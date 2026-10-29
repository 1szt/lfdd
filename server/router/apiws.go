package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// apiws WebSocket	
func apiws() {
	web.Engine.Group("/api/ws").GET("/", handler.GetWs)
}
