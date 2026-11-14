package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

// apiuser 用户接口：me 登录即可（AuthGuard），
// 列表/改角色状态/删除/重置密码仅管理员（RoleGuard）
func apiUserRouter() {
	apiUserHandler := &handler.ApiUserHandler{}
	group := web.Engine.Group("/api/user")

	group.GET("/me", middleware.AuthGuard(), apiUserHandler.Me)
	group.GET("/list", middleware.RoleGuard(), apiUserHandler.List)
	group.POST("/update", middleware.RoleGuard(), apiUserHandler.Update)
	group.POST("/delete", middleware.RoleGuard(), apiUserHandler.Delete)
	group.POST("/reset-password", middleware.RoleGuard(), apiUserHandler.ResetPassword)
}
