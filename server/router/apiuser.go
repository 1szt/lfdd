package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

// apiuser 用户接口：me 登录即可（AuthGuard），
// 列表/新建/改角色状态/改用户名密码/删除/重置密码/设备（会话）管理仅管理员（RoleGuard）
func apiUserRouter() {
	apiUserHandler := &handler.ApiUserHandler{}
	group := web.Engine.Group("/api/user")

	group.GET("/me", middleware.AuthGuard(), apiUserHandler.Me)
	group.GET("/list", middleware.RoleGuard(), apiUserHandler.List)
	group.POST("/create", middleware.RoleGuard(), apiUserHandler.Create)
	group.POST("/update", middleware.RoleGuard(), apiUserHandler.Update)
	group.POST("/update-info", middleware.RoleGuard(), apiUserHandler.UpdateInfo)
	group.POST("/delete", middleware.RoleGuard(), apiUserHandler.Delete)
	group.POST("/reset-password", middleware.RoleGuard(), apiUserHandler.ResetPassword)
	group.GET("/sessions", middleware.RoleGuard(), apiUserHandler.Sessions)
	group.POST("/kick", middleware.RoleGuard(), apiUserHandler.Kick)
}
