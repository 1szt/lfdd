package home

import (
	"hxzl/web"
	"log"

	"github.com/gin-gonic/gin"
)

func Start() {
	web.Engine.GET("/", home)

	log.Print("✅ [Home] 主页模块 加载完成！")
}

func home(c *gin.Context) {
	c.String(200, "Hello, World!")
}
