package api

import (
	"hxzl/web"
	"log"

	"github.com/gin-gonic/gin"
)

func Start() {
	group := web.Engine.Group("/api")
	group.GET("/", GetAPI)
	log.Print("✅ [API] API模块 加载完成！")
}

func GetAPI(c *gin.Context) {
	c.String(200, "Hello, API!")
}
