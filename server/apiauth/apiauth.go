package apiauth

import (
	"github.com/gin-gonic/gin"
	"hxzl/web"
	"log"
)

func Start() {
	group := web.Engine.Group("/api/auth")
	group.POST("/login", PostLogin)
	group.POST("/register", PostRegister)

	log.Print("✅ [APIAUTH] 认证接口 加载完成！")
}

func PostLogin(c *gin.Context) {
	// 获取请求参数
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}
	// 如果user=1 密码等于1
	if req.Username == "1" && req.Password == "1" {
		c.JSON(200, gin.H{
			"message": "登录成功",
		})
		return
	}

	c.JSON(401, gin.H{
		"message": "登录失败",
	})
}

func PostRegister(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "注册成功",
	})
}
