package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// registerMenuRoutes 菜单接口
func registerMenuRoutes() {
	web.Engine.Group("/api").GET("/menu", handler.GetMenu)
}
