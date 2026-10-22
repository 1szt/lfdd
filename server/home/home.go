package home

import (
	"hxzl/server"
	"log"

	g "github.com/gin-gonic/gin"
)

func Start() {
	server.Router.GET("/", home)

	log.Print("✅ [Home] 主页模块 加载完成！")
}

func home(c *g.Context) {
	c.String(200, "Hello, World!")
}
