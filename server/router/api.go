package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// registerAPIRoutes API 根路径探活
func registerAPIRoutes() {
	web.Engine.Group("/api").GET("/", handler.GetAPI)
}
