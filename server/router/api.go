package router

import (
	"hxzl/core/web"

	"github.com/gin-gonic/gin"
)

type ApiHandler struct {
}

// api API 根路径探活
func apiRouter() {
	apiHandler := &ApiHandler{}
	group := web.Engine.Group("/api")
	group.GET("/", apiHandler.API)
}

// API API 根路径探活
func (h *ApiHandler) API(c *gin.Context) {
	c.String(200, "Hello, API!")
}
