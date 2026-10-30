package router

import (
	"hxzl/core/web"

	"github.com/gin-gonic/gin"
)

type Api struct {
}

// api API 根路径探活
func api() {
	Api := Api{}	
	Group := web.Engine.Group("/api")
	Group.GET("/", Api.API)
}

// API API 根路径探活
func (Api) API(c *gin.Context) {
	c.String(200, "Hello, API!")
}
