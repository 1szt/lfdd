// Package router 路由层：统一注册全部 HTTP 路由，handler 只提供处理函数。
package router

import (
	"hxzl/core/web"
	"hxzl/handler"
	"log"
)

// Start 注册全部路由（须在 web.Start 之后、server.Start 之前调用）
func Start() {
	// API 根路径探活
	web.Engine.Group("/api").GET("/", handler.GetAPI)

	// 认证接口
	auth := web.Engine.Group("/api/auth")
	auth.POST("/login", handler.PostLogin)
	auth.POST("/logout", handler.PostLogout)
	auth.POST("/refresh", handler.PostRefresh)
	auth.POST("/register", handler.PostRegister)

	// 菜单接口
	web.Engine.Group("/api").GET("/menu", handler.GetMenu)

	// 安装初始化接口
	install := web.Engine.Group("/api/install")
	install.GET("/check", handler.GetCheck)
	install.POST("/submit", handler.PostSubmit)

	// 热重载测试页
	test := web.Engine.Group("/api")
	test.GET("/test/reload", handler.GetPage)
	test.GET("/test/reload/status", handler.GetStatus)
	test.POST("/test/reload/apply", handler.PostApply)

	// WebSocket
	web.Engine.Group("/api/ws").GET("/", handler.GetWs)

	log.Print("✅ [ROUTER] 路由注册 加载完成！")
}
