package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// apinstall 安装初始化接口
func apinstall() {
	install := web.Engine.Group("/api/install")
	install.GET("/check", handler.GetCheck)
	install.POST("/submit", handler.PostSubmit)
}
