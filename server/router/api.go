package router

import (
	"hxzl/core/web"

	"github.com/gin-gonic/gin"
)

// api API 根路径探活
func api() {
	Group := web.Engine.Group("/api")
	Group.GET("/", GetAPI)
}

// GetAPI API 根路径探活
func GetAPI(c *gin.Context) {
	c.String(200, "Hello, API!")
}
