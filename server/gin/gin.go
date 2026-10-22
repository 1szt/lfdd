package gin

import (
	"context"
	"hxzl/config"
	"log"
	"net/http"
	"time"

	g "github.com/gin-gonic/gin"
)

var Router = g.New()
var server *http.Server
var lastConfig string

type ServerStatus struct {
	IP     string `json:"ip"`
	Port   string `json:"port"`
	Status string `json:"status"`
}

// Start 启动 web 服务（阻塞，从配置读取 ip 和端口）
func Start() {
	ip := config.Get("server_ip").Value
	port := config.Get("server_port").Value
	if port == "" {
		port = "9081"
	}
	start(ip + ":" + port)
}

// start 在指定地址启动服务
func start(addr string) {
	server = &http.Server{
		Addr:         addr,
		Handler:      Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🌐 [Gin] 访问 http://%s", addr)
	lastConfig = addr

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ [Gin] 致命错误：%v", err)
	}
}

// Restart 停掉当前服务，从配置重新读取 ip 和端口重启
func Restart() {
	ip := config.Get("server_ip").Value
	port := config.Get("server_port").Value
	if port == "" {
		port = "9081"
	}
	addr := ip + ":" + port

	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("❌ [Gin] 停止旧服务失败: %v", err)
		} else {
			log.Printf("🔄 [Gin] 已停止旧端口服务")
		}
	}

	// 在新地址启动（开一个 goroutine 避免阻塞）
	go start(addr)
	log.Printf("✅ [Gin] 服务已切换到 %s", addr)
	lastConfig = addr
}

// Reload 智能重连（配置无变化且服务运行中则跳过，否则重启）
func Reload() {
	ip := config.Get("server_ip").Value
	port := config.Get("server_port").Value
	if port == "" {
		port = "9081"
	}
	addr := ip + ":" + port

	running := server != nil

	// 配置没变且运行中 → 跳过
	if running && addr == lastConfig {
		log.Println("🔄 [Gin] 配置无变化且服务运行中，跳过重载")
		return
	}

	// 重载
	if !running {
		log.Println("🔄 [Gin] 服务未运行，启动服务...")
	} else {
		log.Println("🔄 [Gin] 配置已变化，重启服务...")
	}
	Restart()
}

// Status 获取当前服务状态
func Status() ServerStatus {
	ip := ""
	port := ""
	status := "stopped"

	if server != nil {
		status = "running"
		// 从配置读取，确保准确
		ip = config.Get("server_ip").Value
		port = config.Get("server_port").Value
		if addr := server.Addr; addr != "" {
			// 从 Addr 中解析端口
			for i := len(addr) - 1; i >= 0; i-- {
				if addr[i] == ':' {
					port = addr[i+1:]
					break
				}
			}
		}
	}

	return ServerStatus{
		IP:     ip,
		Port:   port,
		Status: status,
	}
}
