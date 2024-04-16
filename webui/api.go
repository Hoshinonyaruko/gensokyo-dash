package webui

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-dashboard/apistats"
	"github.com/hoshinonyaruko/gensokyo-dashboard/config"
	"github.com/hoshinonyaruko/gensokyo-dashboard/sqlite"
	"github.com/hoshinonyaruko/gensokyo-dashboard/sys"
)

//go:embed dist/*
//go:embed dist/icons/*
//go:embed dist/assets/*
var content embed.FS

const configFile = "config.json"

// NewCombinedMiddleware 创建并返回一个带有依赖的中间件闭包
func CombinedMiddleware(config config.Config, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/webui/api") {

			// 处理/api/login的POST请求
			if c.Param("filepath") == "/api/login" && c.Request.Method == http.MethodPost {
				HandleLoginRequest(c, config, db)
				return
			}
			// 处理/api/check-login-status的GET请求
			if c.Param("filepath") == "/api/check-login-status" && c.Request.Method == http.MethodGet {
				HandleCheckLoginStatusRequest(db, c)
				return
			}
			// 处理 /api/get-json 的GET请求
			if c.Param("filepath") == "/api/getjson" && c.Request.Method == http.MethodGet {
				HandleGetJSON(c, config, db)
				return
			}
			// 处理 /api/save-json 的POST请求
			if c.Param("filepath") == "/api/savejson" && c.Request.Method == http.MethodPost {
				HandleSaveJSON(c, config, db)
				return
			}
			// 处理 /api/restartself 的GET请求
			if c.Param("filepath") == "/api/restartself" && c.Request.Method == http.MethodGet {
				HandleRestartSelf(c, config, db)
				return
			}
			// 处理 /api/online-robots 的GET请求
			if c.Param("filepath") == "/api/online-robots" && c.Request.Method == http.MethodGet {
				HandleOnlineRobots(c, &config, db)
				return
			}
			// 处理 /api/robot-info 的GET请求
			if c.Param("filepath") == "/api/robot-info" && c.Request.Method == http.MethodGet {
				HandleRobotInfo(c, db)
				return
			}
			// 处理 /api/robot-info-all 的GET请求
			if c.Param("filepath") == "/api/robot-info-all" && c.Request.Method == http.MethodGet {
				HandleRobotInfoAll(c, db)
				return
			}
			// 处理 /api/api-info 的GET请求
			if c.Param("filepath") == "/api/api-info" && c.Request.Method == http.MethodGet {
				HandleApiInfo(c, config, db)
				return
			}
			// 处理 /api/command-all 的GET请求
			if c.Param("filepath") == "/api/command-all" && c.Request.Method == http.MethodGet {
				HandleCommandAll(c, &config, db)
				return
			}
			// 处理 /api/command-daily 的GET请求
			if c.Param("filepath") == "/api/command-daily" && c.Request.Method == http.MethodGet {
				HandleCommandDaily(c, &config, db)
				return
			}
			// 处理 /api/group-all 的GET请求
			if c.Param("filepath") == "/api/group-all" && c.Request.Method == http.MethodGet {
				HandleGroupAll(c, db)
				return
			}
			// 处理 /api/group-daily 的GET请求
			if c.Param("filepath") == "/api/group-daily" && c.Request.Method == http.MethodGet {
				HandleGroupDaily(c, db)
				return
			}
			// 处理 /api/user-all 的GET请求
			if c.Param("filepath") == "/api/user-all" && c.Request.Method == http.MethodGet {
				HandleUserAll(c, db)
				return
			}
			// 处理 /api/user-daily 的GET请求
			if c.Param("filepath") == "/api/user-daily" && c.Request.Method == http.MethodGet {
				HandleUserDaily(c, db)
				return
			}

		} else {
			// 否则，处理静态文件请求
			// 如果请求是 "/webui/" ，默认为 "index.html"
			filepathRequested := c.Param("filepath")
			if filepathRequested == "" || filepathRequested == "/" {
				filepathRequested = "index.html"
			}

			// 使用 embed.FS 读取文件内容
			filepathRequested = strings.TrimPrefix(filepathRequested, "/")
			data, err := content.ReadFile("dist/" + filepathRequested)
			if err != nil {
				c.String(http.StatusNotFound, "File not found: %v", err)
				return
			}

			mimeType := getContentType(filepathRequested)

			c.Data(http.StatusOK, mimeType, data)
		}
		// 调用c.Next()以继续处理请求链
		c.Next()
	}
}

// HandleOnlineRobots returns all online robots' statuses for the current day in JSON
func HandleOnlineRobots(c *gin.Context, config *config.Config, db *sql.DB) {
	jsonData, err := sqlite.FetchOnlineRobots(db, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", jsonData)
}

func getContentType(path string) string {
	// todo 根据需要增加更多的 MIME 类型
	switch filepath.Ext(path) {
	case ".html":
		return "text/html"
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "text/plain"
	}
}

// HandleLoginRequest处理登录请求
func HandleLoginRequest(c *gin.Context, config config.Config, db *sql.DB) {
	var json struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if checkCredentials(json.Username, json.Password, config) {
		// 如果验证成功，设置cookie
		cookieValue, err := GenerateCookie(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate cookie"})
			return
		}

		c.SetCookie("login_cookie", cookieValue, 3600*24, "/", "", false, true)

		c.JSON(http.StatusOK, gin.H{
			"isLoggedIn": true,
			"cookie":     cookieValue,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"isLoggedIn": false,
		})
	}
}

func checkCredentials(username, password string, jsonconfig config.Config) bool {
	serverUsername := jsonconfig.Account
	serverPassword := jsonconfig.Password
	fmt.Printf("有用户正尝试使用 用户名:%v 密码:%v 进行登入\n", username, password)
	fmt.Printf("A user is attempting to log in with Username: %v Password: %v\n", username, password)

	fmt.Printf("请使用默认登入用户[%v] 默认密码[%v] 进行登入,不包含[],遇到问题可到QQ群:196173384 请教\n", serverUsername, serverPassword)
	fmt.Printf("please use default account[%v] default password[%v] to login, not include []\n", serverUsername, serverPassword)
	return username == serverUsername && password == serverPassword
}

// HandleCheckLoginStatusRequest 检查登录状态的处理函数
func HandleCheckLoginStatusRequest(db *sql.DB, c *gin.Context) {
	// 从请求中获取cookie
	cookieValue, err := c.Cookie("login_cookie")
	if err != nil {
		// 如果cookie不存在，而不是返回BadRequest(400)，我们返回一个OK(200)的响应
		c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Cookie not provided"})
		return
	}

	// 使用ValidateCookie函数验证cookie
	isValid, err := ValidateCookie(db, cookieValue)
	if err != nil {
		switch err {
		case ErrCookieNotFound:
			c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Cookie not found"})
		case ErrCookieExpired:
			c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Cookie has expired"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"isLoggedIn": false, "error": "Internal server error"})
		}
		return
	}

	if isValid {
		c.JSON(http.StatusOK, gin.H{"isLoggedIn": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Invalid cookie"})
	}
}

// HandleGetJSON 返回当前的config作为JSON
func HandleGetJSON(c *gin.Context, cfg config.Config, db *sql.DB) {
	// 从请求中获取cookie
	cookieValue, err := c.Cookie("login_cookie")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Cookie not provided"})
		return
	}

	// 使用ValidateCookie函数验证cookie
	isValid, err := ValidateCookie(db, cookieValue)
	if err != nil || !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid cookie"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// HandleSaveJSON 从请求体中读取JSON并更新config
func HandleSaveJSON(c *gin.Context, cfg config.Config, db *sql.DB) {
	// 从请求中获取cookie
	cookieValue, err := c.Cookie("login_cookie")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Cookie not provided"})
		return
	}

	// 使用ValidateCookie函数验证cookie
	isValid, err := ValidateCookie(db, cookieValue)
	if err != nil || !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid cookie"})
		return
	}

	var newConfig config.Config
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用saveFunc来保存config
	writeConfigToFile(newConfig)

	c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully"})

	//重启自身 很快 唰的一下
	sys.RestartApplication()

}

func HandleRestartSelf(c *gin.Context, cfg config.Config, db *sql.DB) {
	// 从请求中获取cookie
	cookieValue, err := c.Cookie("login_cookie")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Cookie not provided"})
		return
	}

	// 使用ValidateCookie函数验证cookie
	isValid, err := ValidateCookie(db, cookieValue)
	if err != nil || !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid cookie"})
		return
	}

	// Cookie验证通过后，执行重启操作
	c.JSON(http.StatusOK, gin.H{"message": "Restart initiated"})
	//重启自身 很快 唰的一下
	sys.RestartApplication()
}

// writeConfigToFile 将配置写回文件
func writeConfigToFile(config config.Config) {
	configJSON, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Fatalf("无法序列化配置: %v", err)
	}

	err = os.WriteFile(configFile, configJSON, 0644)
	if err != nil {
		log.Fatalf("无法写入配置文件: %v", err)
	}
}

// RobotInfo represents the structured information of a single field over multiple days.
type RobotInfo struct {
	Date  string `json:"date"`
	Field string `json:"field"`
	Value string `json:"value"`
}

// HandleRobotInfo handles the GET request to fetch robot info based on the provided parameters.
func HandleRobotInfo(c *gin.Context, db *sql.DB) {
	// Parse URL query parameters
	selfID, err := strconv.ParseInt(c.Query("selfID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid selfID"})
		return
	}

	days, err := strconv.Atoi(c.Query("days"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days"})
		return
	}

	fieldType := c.Query("fieldType")
	if fieldType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fieldType is required"})
		return
	}

	// Fetch field values from the database
	values, err := sqlite.FetchFieldValuesForRobot(db, selfID, days, fieldType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert values into structured format
	endDate := time.Now()
	var robotInfos []RobotInfo
	for i, value := range values {
		date := endDate.AddDate(0, 0, -i).Format("2006-01-02")
		robotInfos = append(robotInfos, RobotInfo{
			Date:  date,
			Field: fieldType,
			Value: value,
		})
	}

	// Return structured data as JSON
	c.JSON(http.StatusOK, robotInfos)
}

func HandleRobotInfoAll(c *gin.Context, db *sql.DB) {
	selfID, err := strconv.ParseInt(c.Query("selfID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid selfID"})
		return
	}

	days, err := strconv.Atoi(c.Query("days"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days"})
		return
	}

	robots, err := sqlite.FetchAllFieldsForRobot(db, selfID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, robots)
}

func HandleApiInfo(c *gin.Context, cfg config.Config, db *sql.DB) {

	// Parse days from the query parameter
	days, err := strconv.Atoi(c.DefaultQuery("days", "7")) // If days is not specified, default to the last 7 days
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days parameter"})
		return
	}

	// Fetch API statuses using the provided function
	apiStatuses, err := apistats.FetchAPIStatuses(db, cfg, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the fetched statuses as JSON
	c.JSON(http.StatusOK, apiStatuses)
}

func HandleCommandAll(c *gin.Context, config *config.Config, db *sql.DB) {
	selfId, err := strconv.ParseInt(c.Query("selfId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing selfId parameter"})
		return
	}

	rank, err := strconv.Atoi(c.Query("rank"))
	if err != nil || rank <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing rank parameter"})
		return
	}

	commands, err := sqlite.FetchTopCommands(db, selfId, rank)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commands)
}

func HandleCommandDaily(c *gin.Context, config *config.Config, db *sql.DB) {
	selfId, err := strconv.ParseInt(c.Query("selfId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing selfId parameter"})
		return
	}

	dateStr := c.Query("date")
	rank, err := strconv.Atoi(c.Query("rank"))
	if err != nil || rank <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rank"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	commands, err := sqlite.FetchTopDailyCommands(db, selfId, date, rank)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commands)
}

func HandleGroupAll(c *gin.Context, db *sql.DB) {
	selfId, err := strconv.ParseInt(c.Query("selfId"), 10, 64)
	if err != nil || selfId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing selfId parameter"})
		return
	}

	rank, err := strconv.Atoi(c.Query("rank"))
	if err != nil || rank <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing rank parameter"})
		return
	}

	groups, err := sqlite.FetchTopGroups(db, selfId, rank)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}

func HandleGroupDaily(c *gin.Context, db *sql.DB) {
	selfId, err := strconv.ParseInt(c.Query("selfId"), 10, 64)
	if err != nil || selfId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing selfId parameter"})
		return
	}

	dateStr := c.Query("date")
	rank, err := strconv.Atoi(c.Query("rank"))
	if err != nil || rank <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rank"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	groups, err := sqlite.FetchTopDailyGroups(db, selfId, date, rank)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}

func HandleUserAll(c *gin.Context, db *sql.DB) {
	selfId, err := strconv.ParseInt(c.Query("selfId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing selfId parameter"})
		return
	}

	rank, err := strconv.Atoi(c.Query("rank"))
	if err != nil || rank <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing rank parameter"})
		return
	}

	users, err := sqlite.FetchTopUsers(db, selfId, rank)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func HandleUserDaily(c *gin.Context, db *sql.DB) {
	selfId, err := strconv.ParseInt(c.Query("selfId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing selfId parameter"})
		return
	}

	dateStr := c.Query("date")
	rank, err := strconv.Atoi(c.Query("rank"))
	if err != nil || rank <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rank"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	users, err := sqlite.FetchTopDailyUsers(db, selfId, date, rank)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
