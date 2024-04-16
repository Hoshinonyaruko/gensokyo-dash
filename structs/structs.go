package structs

import (
	"encoding/json"
	"fmt"
)

// onebot发来的action调用信息
type ActionMessage struct {
	Action      string        `json:"action"`
	Params      ParamsContent `json:"params"`
	Echo        interface{}   `json:"echo,omitempty"`
	PostType    string        `json:"post_type,omitempty"`
	MessageType string        `json:"message_type,omitempty"`
}

func (a *ActionMessage) UnmarshalJSON(data []byte) error {
	type Alias ActionMessage

	var rawEcho json.RawMessage
	temp := &struct {
		*Alias
		Echo *json.RawMessage `json:"echo,omitempty"`
	}{
		Alias: (*Alias)(a),
		Echo:  &rawEcho,
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if rawEcho != nil {
		var lastErr error

		var intValue int
		if lastErr = json.Unmarshal(rawEcho, &intValue); lastErr == nil {
			a.Echo = intValue
			return nil
		}

		var strValue string
		if lastErr = json.Unmarshal(rawEcho, &strValue); lastErr == nil {
			a.Echo = strValue
			return nil
		}

		var arrValue []interface{}
		if lastErr = json.Unmarshal(rawEcho, &arrValue); lastErr == nil {
			a.Echo = arrValue
			return nil
		}

		var objValue map[string]interface{}
		if lastErr = json.Unmarshal(rawEcho, &objValue); lastErr == nil {
			a.Echo = objValue
			return nil
		}

		return fmt.Errorf("unable to unmarshal echo: %v", lastErr)
	}

	return nil
}

// params类型
type ParamsContent struct {
	BotQQ     string      `json:"botqq,omitempty"`
	ChannelID interface{} `json:"channel_id,omitempty"`
	GuildID   interface{} `json:"guild_id,omitempty"`
	GroupID   interface{} `json:"group_id,omitempty"`   // 每一种onebotv11实现的字段类型都可能不同
	MessageID interface{} `json:"message_id,omitempty"` // 用于撤回信息
	Message   interface{} `json:"message,omitempty"`    // 这里使用interface{}因为它可能是多种类型
	Messages  interface{} `json:"messages,omitempty"`   // 坑爹转发信息
	UserID    interface{} `json:"user_id,omitempty"`    // 这里使用interface{}因为它可能是多种类型
	Duration  int         `json:"duration,omitempty"`   // 可选的整数
	Enable    bool        `json:"enable,omitempty"`     // 可选的布尔值
	// handle quick operation
	Context   Context   `json:"context,omitempty"`   // context 字段
	Operation Operation `json:"operation,omitempty"` // operation 字段
}

// Context 结构体用于存储 context 字段相关信息
type Context struct {
	Avatar      string `json:"avatar,omitempty"`       // 用户头像链接
	Font        int    `json:"font,omitempty"`         // 字体（假设是整数类型）
	MessageID   int    `json:"message_id,omitempty"`   // 消息 ID
	MessageSeq  int    `json:"message_seq,omitempty"`  // 消息序列号
	MessageType string `json:"message_type,omitempty"` // 消息类型
	PostType    string `json:"post_type,omitempty"`    // 帖子类型
	SubType     string `json:"sub_type,omitempty"`     // 子类型
	Time        int64  `json:"time,omitempty"`         // 时间戳
	UserID      int    `json:"user_id,omitempty"`      // 用户 ID
	GroupID     int    `json:"group_id,omitempty"`     // 群号
}

// Operation 结构体用于存储 operation 字段相关信息
type Operation struct {
	Reply    string `json:"reply,omitempty"`     // 回复内容
	AtSender bool   `json:"at_sender,omitempty"` // 是否 @ 发送者
}

// 自定义一个ParamsContent的UnmarshalJSON 让GroupID同时兼容str和int
func (p *ParamsContent) UnmarshalJSON(data []byte) error {
	type Alias ParamsContent
	aux := &struct {
		GroupID   interface{} `json:"group_id"`
		UserID    interface{} `json:"user_id"`
		MessageID interface{} `json:"message_id"`
		ChannelID interface{} `json:"channel_id"`
		GuildID   interface{} `json:"guild_id"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.GroupID.(type) {
	case nil: // 当GroupID不存在时
		p.GroupID = ""
	case float64: // JSON的数字默认被解码为float64
		p.GroupID = fmt.Sprintf("%.0f", v) // 将其转换为字符串，忽略小数点后的部分
	case string:
		p.GroupID = v
	default:
		return fmt.Errorf("GroupID has unsupported type")
	}

	switch v := aux.UserID.(type) {
	case nil: // 当UserID不存在时
		p.UserID = ""
	case float64: // JSON的数字默认被解码为float64
		p.UserID = fmt.Sprintf("%.0f", v) // 将其转换为字符串，忽略小数点后的部分
	case string:
		p.UserID = v
	default:
		return fmt.Errorf("UserID has unsupported type")
	}

	switch v := aux.MessageID.(type) {
	case nil: // 当UserID不存在时
		p.MessageID = ""
	case float64: // JSON的数字默认被解码为float64
		p.MessageID = fmt.Sprintf("%.0f", v) // 将其转换为字符串，忽略小数点后的部分
	case string:
		p.MessageID = v
	default:
		return fmt.Errorf("MessageID has unsupported type")
	}

	switch v := aux.ChannelID.(type) {
	case nil: // 当ChannelID不存在时
		p.ChannelID = ""
	case float64: // JSON的数字默认被解码为float64
		p.ChannelID = fmt.Sprintf("%.0f", v) // 将其转换为字符串，忽略小数点后的部分
	case string:
		p.ChannelID = v
	default:
		return fmt.Errorf("MessageID has unsupported type")
	}

	switch v := aux.GuildID.(type) {
	case nil: // 当GuildID不存在时
		p.GuildID = ""
	case float64: // JSON的数字默认被解码为float64
		p.GuildID = fmt.Sprintf("%.0f", v) // 将其转换为字符串，忽略小数点后的部分
	case string:
		p.GuildID = v
	default:
		return fmt.Errorf("MessageID has unsupported type")
	}

	return nil
}

// Message represents a standardized structure for the incoming messages.
type Message struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
	Echo   interface{}            `json:"echo,omitempty"`
}

type MessageEvent struct {
	PostType    string      `json:"post_type"`
	MessageType string      `json:"message_type"`
	Time        int64       `json:"time"`
	SelfID      int64       `json:"self_id"`
	SubType     string      `json:"sub_type"`
	Message     interface{} `json:"message"`
	RawMessage  string      `json:"raw_message"`
	Sender      struct {
		Age      int    `json:"age"`
		Area     string `json:"area"`
		Card     string `json:"card"`
		Level    string `json:"level"`
		Nickname string `json:"nickname"`
		Role     string `json:"role"`
		Sex      string `json:"sex"`
		Title    string `json:"title"`
		UserID   int64  `json:"user_id"`
	} `json:"sender"`
	UserID    int64 `json:"user_id"`
	Anonymous *struct {
	} `json:"anonymous"`
	Font       int   `json:"font"`
	GroupID    int64 `json:"group_id"`
	MessageSeq int64 `json:"message_seq"`
	MessageID  int64 `json:"message_id"`
}

type MetaEvent struct {
	PostType      string `json:"post_type"`
	MetaEventType string `json:"meta_event_type"`
	Time          int64  `json:"time"`
	SelfID        int64  `json:"self_id"`
	Interval      int    `json:"interval"`
	Status        struct {
		AppEnabled     bool  `json:"app_enabled"`
		AppGood        bool  `json:"app_good"`
		AppInitialized bool  `json:"app_initialized"`
		Good           bool  `json:"good"`
		Online         bool  `json:"online"`
		PluginsGood    *bool `json:"plugins_good"`
		Stat           struct {
			PacketReceived  int   `json:"packet_received"`
			PacketSent      int   `json:"packet_sent"`
			PacketLost      int   `json:"packet_lost"`
			MessageReceived int   `json:"message_received"`
			MessageSent     int   `json:"message_sent"`
			DisconnectTimes int   `json:"disconnect_times"`
			LostTimes       int   `json:"lost_times"`
			LastMessageTime int64 `json:"last_message_time"`
		} `json:"stat"`
	} `json:"status"`
}

type NoticeEvent struct {
	GroupID    int64  `json:"group_id"`
	NoticeType string `json:"notice_type"`
	OperatorID int64  `json:"operator_id"`
	PostType   string `json:"post_type"`
	SelfID     int64  `json:"self_id"`
	SubType    string `json:"sub_type"`
	Time       int64  `json:"time"`
	UserID     int64  `json:"user_id"`
}

type RobotStatus struct {
	SelfID          int64  `json:"self_id"`
	Date            string `json:"date"`
	Online          bool   `json:"online"`
	MessageReceived int    `json:"message_received"`
	MessageSent     int    `json:"message_sent"`
	LastMessageTime int64  `json:"last_message_time"`
	InvitesReceived int    `json:"invites_received"`
	KicksReceived   int    `json:"kicks_received"`
	DailyDAU        int    `json:"daily_dau"`
}
