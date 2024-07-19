package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
)

// 配置文件路径
const configFile = "config.json"

type Config struct {
	Account        string    `json:"account"`        // 登入用户名
	Password       string    `json:"password"`       // 登入密码
	Title          string    `json:"title"`          // 自定义标题
	WsPath         string    `json:"wspath"`         // 默认监听裸端点
	Port           string    `json:"port"`           // WebUI端口
	UseHttps       bool      `json:"useHttps"`       // 使用 https
	StoreMsgs      bool      `json:"storeMsgs"`      // 储存每条信息 用于详细分析
	PrintLogs      bool      `json:"printLogs"`      // 输出日志开关
	Cert           string    `json:"cert"`           // 证书
	Key            string    `json:"key"`            // 密钥
	EnableWSServer bool      `json:"enableWsServer"` // 是否启用正向WS服务器
	WSServerToken  string    `json:"wsServerToken"`  // 正向WS的Token
	ApisInfos      []Apis    `json:"apis"`           // api信息数组
	BotInfos       []BotInfo `json:"botInfos"`       // 机器人信息数组
}

type BotInfo struct {
	BotID       string `json:"botId"`       // 机器人的唯一标识
	BotNickname string `json:"botNickname"` // 机器人的昵称
	BotHead     string `json:"botHead"`     // 机器人的头像链接
}

type Apis struct {
	APIPaths string `json:"apiPaths"` // API地址 检测存活
	APINames string `json:"apiNames"` // API名称 一一对应
}

// 默认配置
var defaultConfig = Config{
	UseHttps:       false,
	StoreMsgs:      true,
	PrintLogs:      true,
	Cert:           "",
	Key:            "",
	Account:        "admin",
	Password:       "admin",
	Title:          "",
	WsPath:         "nil",
	Port:           "18630",
	EnableWSServer: true,
	WSServerToken:  "",
	ApisInfos: []Apis{
		{
			APIPaths: "http://127.0.0.1:18630",
			APINames: "自身",
		},
	},
}

// readConfig 尝试读取配置文件，如果失败则创建并自动配置默认配置
func ReadConfig() Config {
	var config Config

	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println("无法读取配置文件, 正在创建默认配置...")
		config = createDefaultConfig()
	} else {
		err = json.Unmarshal(data, &config)
		if err != nil {
			fmt.Println("配置解析失败, 正在使用默认配置...")
			config = defaultConfig
		}
	}

	// 检查并设置默认值
	if checkAndSetDefaults(&config) {
		// 如果配置被修改，写回文件
		WriteConfigToFile(config)
	}

	return config
}

// checkAndSetDefaults 检查并设置默认值，返回是否做了修改
func checkAndSetDefaults(config *Config) bool {
	// 通过反射获取Config的类型和值
	val := reflect.ValueOf(config).Elem()
	typ := val.Type()

	// 记录是否进行了修改
	var modified bool

	// 遍历所有字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		defaultField := reflect.ValueOf(defaultConfig).Field(i)
		fieldType := field.Type()

		// 跳过布尔类型的字段
		if fieldType.Kind() == reflect.Bool {
			continue
		}

		fieldName := typ.Field(i).Name

		// 特殊处理RestartInterval字段
		if fieldName == "RestartInterval" || fieldName == "WhiteCheckTime" || fieldName == "MemoryCleanupInterval" || fieldName == "BackupInterval" || fieldName == "MemoryCheckInterval" {
			continue
		}

		// 如果字段是零值，设置为默认值
		if isZeroOfUnderlyingType(field.Interface()) {
			field.Set(defaultField)
			modified = true
		}
	}

	return modified
}

// isZeroOfUnderlyingType 检查一个值是否为其类型的零值
func isZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

// WriteConfigToFile 将配置写回文件
func WriteConfigToFile(config Config) {
	configJSON, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Fatalf("无法序列化配置: %v", err)
	}

	err = os.WriteFile(configFile, configJSON, 0644)
	if err != nil {
		log.Fatalf("无法写入配置文件: %v", err)
	}
}

// createDefaultConfig 创建一个带有默认值的配置文件，并返回这个配置
func createDefaultConfig() Config {
	// 序列化默认配置
	data, err := json.MarshalIndent(defaultConfig, "", "    ")
	if err != nil {
		fmt.Println("无法创建默认配置文件:", err)
		os.Exit(1)
	}

	// 将默认配置写入文件
	err = os.WriteFile(configFile, data, 0666)
	if err != nil {
		fmt.Println("无法写入默认配置文件:", err)
		os.Exit(1)
	}

	fmt.Println("默认配置文件已创建:", configFile)

	// 返回默认配置
	return defaultConfig
}
