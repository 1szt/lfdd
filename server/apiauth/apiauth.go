package apiauth

import (
	"github.com/gin-gonic/gin"
	"hxzl/web"
	"log"
)

func Start() {
	auth := web.Engine.Group("/api/auth")
	auth.POST("/login", Login)
	log.Print("✅ [APIAUTH] 认证接口 加载完成！")
}

func Login(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "登录成功",
	})
}
