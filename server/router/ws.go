package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// registerWSRoutes WebSocket
func registerWSRoutes() {
	web.Engine.Group("/api/ws").GET("/", handler.GetWs)
}
