package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

// apimenu 菜单管理接口：新建/改名/删除/顺序整组重写，仅管理员（RoleGuard）
func apiMenuRouter() {
	apiMenuHandler := &handler.ApiMenuHandler{}
	group := web.Engine.Group("/api/menu")
	group.POST("/create", middleware.RoleGuard(), apiMenuHandler.Create)
	group.POST("/update", middleware.RoleGuard(), apiMenuHandler.Update)
	group.POST("/delete", middleware.RoleGuard(), apiMenuHandler.Delete)
	group.POST("/sort", middleware.RoleGuard(), apiMenuHandler.Sort)
}
