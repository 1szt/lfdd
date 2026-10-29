package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// api API 根路径探活
func api() {
	web.Engine.Group("/api").GET("/", handler.GetAPI)
}
