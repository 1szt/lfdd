package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// apiauth 认证接口
func apiauth() {
	auth := web.Engine.Group("/api/auth")
	auth.POST("/login", handler.PostLogin)
	auth.POST("/logout", handler.PostLogout)
	auth.POST("/refresh", handler.PostRefresh)
	auth.POST("/register", handler.PostRegister)
}
