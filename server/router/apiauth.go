package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

type ApiAuthHandler struct {
}

// apiauth 认证接口：guest/login/register 经 GuestGuard（未登录放行、已登录阻止），
// logout/refresh 经 AuthGuard（登录放行、未登录阻止）；guest/revoke 主动注销游客身份
func apiAuthRouter() {
	apiAuthHandler := &handler.ApiAuthHandler{}
	group := web.Engine.Group("/api/auth")
	group.POST("/guest", middleware.GuestGuard(), apiAuthHandler.Guest)
	group.POST("/guest/revoke", apiAuthHandler.RevokeGuest)
	group.POST("/login", middleware.GuestGuard(), apiAuthHandler.Login)
	group.POST("/logout", middleware.AuthGuard(), apiAuthHandler.Logout)
	group.POST("/refresh", middleware.AuthGuard(), apiAuthHandler.Refresh)
	group.POST("/register", middleware.GuestGuard(), apiAuthHandler.Register)
}
