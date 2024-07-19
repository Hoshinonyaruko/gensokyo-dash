package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hoshinonyaruko/gensokyo-dashboard/config"
	"github.com/hoshinonyaruko/gensokyo-dashboard/mylog"
	"github.com/hoshinonyaruko/gensokyo-dashboard/sqlite"
	"github.com/hoshinonyaruko/gensokyo-dashboard/structs"
)

type WebSocketServerClient struct {
	Conn *websocket.Conn
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 使用闭包结构 因为gin需要c *gin.Context固定签名
func WsHandlerWithDependencies(config config.Config, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsHandler(c, config, db)
	}
}

// 处理正向ws客户端的连接
func wsHandler(c *gin.Context, config config.Config, db *sql.DB) {
	// 先从请求头中尝试获取token
	tokenFromHeader := c.Request.Header.Get("Authorization")
	token := ""
	if tokenFromHeader != "" {
		if strings.HasPrefix(tokenFromHeader, "Token ") {
			// 从 "Token " 后面提取真正的token值
			token = strings.TrimPrefix(tokenFromHeader, "Token ")
		} else if strings.HasPrefix(tokenFromHeader, "Bearer ") {
			// 从 "Bearer " 后面提取真正的token值
			token = strings.TrimPrefix(tokenFromHeader, "Bearer ")
		} else {
			// 直接使用token值
			token = tokenFromHeader
		}
	} else {
		// 如果请求头中没有token，则从URL参数中获取
		token = c.Query("access_token")
	}

	// 获取配置中的有效 token
	validToken := config.WSServerToken

	// 如果配置的 token 不为空，但提供的 token 为空或不匹配
	if validToken != "" && (token == "" || token != validToken) {
		if token == "" {
			mylog.Printf("Connection failed due to missing token. Headers: %v", c.Request.Header)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		} else {
			mylog.Printf("Connection failed due to incorrect token. Headers: %v, Provided token: %s", c.Request.Header, token)
			c.JSON(http.StatusForbidden, gin.H{"error": "Incorrect token"})
		}
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		mylog.Printf("Failed to set websocket upgrade: %+v", err)
		return
	}

	clientIP := c.ClientIP()
	mylog.Printf("WebSocket client connected. IP: %s", clientIP)

	// 创建WebSocketServerClient实例
	client := &WebSocketServerClient{
		Conn: conn,
	}

	botID := 123

	// 发送连接成功的消息
	message := map[string]interface{}{
		"meta_event_type": "lifecycle",
		"post_type":       "meta_event",
		"self_id":         botID,
		"sub_type":        "connect",
		"time":            int(time.Now().Unix()),
	}
	err = client.SendMessage(message)
	if err != nil {
		mylog.Printf("Error sending connection success message: %v\n", err)
	}

	//退出时候的清理
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			mylog.Printf("Error reading message: %v", err)
			return
		}

		if messageType == websocket.TextMessage {
			processWSMessage(p, db, config)
		}
	}
}

// 处理收到的信息
func processWSMessage(msg []byte, db *sql.DB, config config.Config) {
	var genericMap map[string]interface{}
	if err := json.Unmarshal(msg, &genericMap); err != nil {
		log.Printf("Error unmarshalling message to map: %v, Original message: %s\n", err, string(msg))
		return
	}

	// Assuming there's a way to distinguish notice messages, for example, checking if notice_type exists
	if noticeType, ok := genericMap["notice_type"].(string); ok && noticeType != "" {
		var noticeEvent structs.NoticeEvent
		if err := json.Unmarshal(msg, &noticeEvent); err != nil {
			log.Printf("Error unmarshalling notice event: %v\n", err)
			return
		}
		fmt.Printf("Processed a notice event of type '%s' from group %d.\n", noticeEvent.NoticeType, noticeEvent.GroupID)
		//进入快乐的处理流程 write
		err := sqlite.ProcessNoticeEvent(db, noticeEvent, config)
		if err != nil {
			fmt.Printf("sqlite.ProcessNoticeEvent error %v.\n", err)
		}
	} else if postType, ok := genericMap["post_type"].(string); ok {
		switch postType {
		case "message":
			var messageEvent structs.MessageEvent
			if err := json.Unmarshal(msg, &messageEvent); err != nil {
				log.Printf("Error unmarshalling message event: %v\n", err)
				return
			}

			if config.PrintLogs {
				fmt.Printf("Processed a message event from group %d.\n", messageEvent.GroupID)
			}

			//进入快乐的处理流程 write
			err := sqlite.ProcessMessageEvent(db, messageEvent, config)
			if err != nil {
				fmt.Printf("sqlite.ProcessMessageEvent error %v.\n", err)
			}
		case "meta_event":
			var metaEvent structs.MetaEvent
			if err := json.Unmarshal(msg, &metaEvent); err != nil {
				log.Printf("Error unmarshalling meta event: %v\n", err)
				return
			}
			fmt.Printf("Processed a meta event, heartbeat interval: %d.\n", metaEvent.Interval)
			//进入快乐的处理流程 write
			err := sqlite.ProcessMetaEvent(db, metaEvent)
			if err != nil {
				fmt.Printf("sqlite.ProcessMetaEvent error %v.\n", err)
			}
		}
	} else {
		log.Printf("Unknown message type or missing post type\n")
	}
}

// 发信息给client
func (c *WebSocketServerClient) SendMessage(message map[string]interface{}) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		mylog.Println("Error marshalling message:", err)
		return err
	}
	return c.Conn.WriteMessage(websocket.TextMessage, msgBytes)
}

func (client *WebSocketServerClient) Close() error {
	return client.Conn.Close()
}

// 截断信息
func TruncateMessage(message structs.ActionMessage, maxLength int) string {
	paramsStr, err := json.Marshal(message.Params)
	if err != nil {
		return "Error marshalling Params for truncation."
	}

	// Truncate Params if its length exceeds maxLength
	truncatedParams := string(paramsStr)
	if len(truncatedParams) > maxLength {
		truncatedParams = truncatedParams[:maxLength] + "..."
	}

	return fmt.Sprintf("Action: %s, Params: %s, Echo: %v", message.Action, truncatedParams, message.Echo)
}
