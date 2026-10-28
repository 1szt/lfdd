package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// registerInstallRoutes 安装初始化接口
func registerInstallRoutes() {
	install := web.Engine.Group("/api/install")
	install.GET("/check", handler.GetCheck)
	install.POST("/submit", handler.PostSubmit)
}
