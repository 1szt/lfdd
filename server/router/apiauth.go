package router

import (
	"hxzl/core/web"
	"hxzl/handler"
	"hxzl/middleware"
)

type ApiAuthHandler struct {
}

// apiauth 认证接口：login/register 公开，logout/refresh 经 Auth 中间件解析 token
func apiAuthRouter() {
	apiAuthHandler := &handler.ApiAuthHandler{}
	group := web.Engine.Group("/api/auth")
	group.POST("/login", apiAuthHandler.Login)
	group.POST("/logout", middleware.Auth(), apiAuthHandler.Logout)
	group.POST("/refresh", middleware.Auth(), apiAuthHandler.Refresh)
	group.POST("/register", apiAuthHandler.Register)
}
