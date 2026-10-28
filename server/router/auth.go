package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// registerAuthRoutes 认证接口
func registerAuthRoutes() {
	auth := web.Engine.Group("/api/auth")
	auth.POST("/login", handler.PostLogin)
	auth.POST("/logout", handler.PostLogout)
	auth.POST("/refresh", handler.PostRefresh)
	auth.POST("/register", handler.PostRegister)
}
