package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// apimenu 菜单接口
func apimenu() {
	web.Engine.Group("/api").GET("/menu", handler.GetMenu)
}
