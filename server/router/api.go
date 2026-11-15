package router

import (
	"server/core/web"
	"server/handler"
)

// api API 根路径探活
func apiRouter() {
	apiHandler := &handler.ApiHandler{}
	group := web.Engine.Group("/api")
	group.GET("/", apiHandler.API)
}
