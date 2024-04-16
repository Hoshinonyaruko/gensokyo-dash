package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-dashboard/apistats"
	"github.com/hoshinonyaruko/gensokyo-dashboard/config"
	"github.com/hoshinonyaruko/gensokyo-dashboard/mylog"
	"github.com/hoshinonyaruko/gensokyo-dashboard/server"
	"github.com/hoshinonyaruko/gensokyo-dashboard/sqlite"
	"github.com/hoshinonyaruko/gensokyo-dashboard/sys"
	"github.com/hoshinonyaruko/gensokyo-dashboard/webui"

	_ "github.com/mattn/go-sqlite3" // 只导入，作为驱动
)

// APIStatus 结构用于保存API的URL和它的状态
type APIStatus struct {
	URL    string
	Status string
}

func main() {

	// 读取或创建配置
	jsonconfig := config.ReadConfig()

	// 打印配置以确认
	fmt.Printf("当前配置: %#v\n", jsonconfig)
	fmt.Printf("作者 早苗狐 答疑群:196173384\n")

	//给程序整个标题
	sys.SetTitle(jsonconfig.Title + " 作者 早苗狐 答疑群:196173384")

	// 打开数据库，使用参数启动SQLite
	db, err := sql.Open("sqlite3", "file:mydb.sqlite?cache=shared&mode=rwc")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 启动WAL模式
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatalf("Failed to set WAL mode: %v", err)
	}

	// 检查当前的journal_mode
	var journalMode string
	row := db.QueryRow("PRAGMA journal_mode;")
	if err := row.Scan(&journalMode); err != nil {
		log.Fatalf("Failed to fetch journal mode: %v", err)
	}
	fmt.Printf("Database journal mode is set to: %s\n", journalMode)

	// 确保表存在 函数需要符合幂等性
	err = sqlite.EnsureCookieTablesExist(db) //网页登入cookie表
	if err != nil {
		log.Fatalf("sqlite.EnsureCookieTablesExist: %v", err)
	}
	err = sqlite.EnsureMessagesTableExists(db) //全量收信息表
	if err != nil {
		log.Fatalf("sqlite.EnsureMessagesTableExists: %v", err)
	}
	err = sqlite.EnsureRobotStatusTableExists(db) //机器人状态表
	if err != nil {
		log.Fatalf("sqlite.EnsureRobotStatusTableExists: %v", err)
	}
	err = sqlite.EnsureUserStatsTableExists(db) //用户统计表
	if err != nil {
		log.Fatalf(" sqlite.EnsureUserStatsTableExists: %v", err)
	}
	err = sqlite.EnsureGroupStatsTableExists(db) //群统计表
	if err != nil {
		log.Fatalf("sqlite.EnsureGroupStatsTableExists %v", err)
	}
	err = sqlite.EnsureCommandStatsTables(db) //指令统计表
	if err != nil {
		log.Fatalf("sqlite.EnsureCommandStatsTableExists: %v", err)
	}
	err = sqlite.EnsureAPITableExists(db) //api状态表
	if err != nil {
		log.Fatalf("sqlite.EnsureAPITableExists: %v", err)
	}

	r := gin.Default()

	//webui和它的api
	webuiGroup := r.Group("/webui")
	{
		webuiGroup.GET("/*filepath", webui.CombinedMiddleware(jsonconfig, db))
		webuiGroup.POST("/*filepath", webui.CombinedMiddleware(jsonconfig, db))
		webuiGroup.PUT("/*filepath", webui.CombinedMiddleware(jsonconfig, db))
		webuiGroup.DELETE("/*filepath", webui.CombinedMiddleware(jsonconfig, db))
		webuiGroup.PATCH("/*filepath", webui.CombinedMiddleware(jsonconfig, db))
	}
	//正向ws

	wspath := jsonconfig.WsPath
	if wspath == "nil" {
		r.GET("", server.WsHandlerWithDependencies(jsonconfig, db))
		mylog.Println("正向ws启动成功,监听0.0.0.0:" + jsonconfig.Port + "请注意设置ws_server_token(可空),并对外放通端口...")
	} else {
		r.GET("/"+wspath, server.WsHandlerWithDependencies(jsonconfig, db))
		mylog.Println("正向ws启动成功,监听0.0.0.0:" + jsonconfig.Port + "/" + wspath + "请注意设置ws_server_token(可空),并对外放通端口...")
	}

	// 创建一个http.Server实例(主服务器)
	httpServer := &http.Server{
		Addr:    "0.0.0.0:" + jsonconfig.Port,
		Handler: r,
	}

	if jsonconfig.UseHttps {
		fmt.Printf("webui-api运行在 HTTPS 端口 %v\n", jsonconfig.Port)
		// 在一个新的goroutine中启动主服务器
		go func() {
			// 定义默认的证书和密钥文件名 自签名证书
			certFile := "cert.pem"
			keyFile := "key.pem"
			if jsonconfig.Cert != "" && jsonconfig.Key != "" {
				certFile = jsonconfig.Cert
				keyFile = jsonconfig.Key
			}
			// 使用 HTTPS
			if err := httpServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}

		}()
	} else {
		fmt.Printf("webui-api运行在 HTTP 端口 %v\n", jsonconfig.Port)
		// 在一个新的goroutine中启动主服务器
		go func() {
			// 使用HTTP
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		}()
	}

	// 运行api监测
	apistats.MonitorAPIs(db, jsonconfig)

	// 设置信号捕获
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-sigChan
	// 可以执行退出程序
	// 正常退出程序
	os.Exit(0)

}
