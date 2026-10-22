package api

import (
	"hxzl/web"
	"log"

	"github.com/gin-gonic/gin"
)

func Start() {
	apiRoute := web.Engine.Group("/api")
	apiRoute.GET("/", home)
	apiRoute.GET("/menu", GetMenu)

	// 安装初始化接口
	apiRoute.GET("/install/check", InstallCheck)
	apiRoute.POST("/install/submit", InstallSubmit)

	log.Print("✅ [API] API模块 加载完成！")
}

func home(c *gin.Context) {
	c.String(200, "API")
}
