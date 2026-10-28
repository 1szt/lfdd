package apiauth

import (
	"hxzl/service"
	"hxzl/web"
	"log"

	"github.com/gin-gonic/gin"
)

type LoginReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func Start() {
	group := web.Engine.Group("/api/auth")
	group.POST("/login", PostLogin)
	group.POST("/logout", PostLogout)
	group.POST("/refresh", PostRefresh)
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
			"token": "1",
		})
		return
	}

	c.JSON(401, gin.H{
		"message": "登录失败",
	})
}

func PostRegister(c *gin.Context) {

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	user, err := service.Register(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "注册成功",
		"user":    user,
	})
}

func PostLogout(c *gin.Context) {

	c.JSON(200, gin.H{
		"message": "退出成功",
	})
}

func PostRefresh(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "刷新成功",
	})
}
