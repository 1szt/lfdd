package router

import (
	"hxzl/core/web"
	"hxzl/service"

	"github.com/gin-gonic/gin"
)

type ApiAuthHandler struct {
}

// apiauth 认证接口
func apiAuthRouter() {
	apiAuthHandler := &ApiAuthHandler{}
	group := web.Engine.Group("/api/auth")
	group.POST("/login", apiAuthHandler.Login)
	group.POST("/logout", apiAuthHandler.Logout)
	group.POST("/refresh", apiAuthHandler.Refresh)
	group.POST("/register", apiAuthHandler.Register)
}

// Login 登录
func (h *ApiAuthHandler) Login(c *gin.Context) {
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

// Register 注册
func (h *ApiAuthHandler) Register(c *gin.Context) {

	var req service.Auth

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	user, err := req.Register()
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

// Logout 退出
func (h *ApiAuthHandler) Logout(c *gin.Context) {

	c.JSON(200, gin.H{
		"message": "退出成功",
	})
}

// Refresh 刷新 token
func (h *ApiAuthHandler) Refresh(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "刷新成功",
	})
}
