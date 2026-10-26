package apiauth

import (
	"github.com/gin-gonic/gin"
	"hxzl/web"
	"log"
)

func Start() {
	group := web.Engine.Group("/api/auth")
	group.POST("/login", PostLogin)
	log.Print("✅ [APIAUTH] 认证接口 加载完成！")
}

func PostLogin(c *gin.Context) {	
	c.JSON(200, gin.H{
		"message": "登录成功",
	})
}
