package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// api API 根路径探活
func apiRouter() {
	apiHandler := &handler.ApiHandler{}
	group := web.Engine.Group("/api")
	group.GET("/", apiHandler.API)
}
