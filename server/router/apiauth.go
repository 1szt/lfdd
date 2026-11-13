package router

import (
	"hxzl/core/web"
	"hxzl/middleware"
	"hxzl/service"

	"github.com/gin-gonic/gin"
)

type ApiAuthHandler struct {
}

// apiauth 认证接口：login/register 公开，logout/refresh 经 Auth 中间件解析 token
func apiAuthRouter() {
	apiAuthHandler := &ApiAuthHandler{}
	group := web.Engine.Group("/api/auth")
	group.POST("/login", apiAuthHandler.Login)
	group.POST("/logout", middleware.Auth(), apiAuthHandler.Logout)
	group.POST("/refresh", middleware.Auth(), apiAuthHandler.Refresh)
	group.POST("/register", apiAuthHandler.Register)
}

// Login 登录
func (h *ApiAuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"err":     400,
			"message": "请求参数错误",
		})
		return
	}

	token, err := service.Login(service.AuthInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		IP:       c.ClientIP(),
		Device:   c.GetHeader("User-Agent"),
	})
	if err != nil {
		c.JSON(400, gin.H{
			"code":     400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}

// Register 注册
func (h *ApiAuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":     400,
			"message": "请求参数错误",
		})
		return
	}

	token, err := service.Register(service.AuthInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		IP:       c.ClientIP(),
		Device:   c.GetHeader("User-Agent"),
	})
	if err != nil {
		c.JSON(400, gin.H{
			"code":     400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}

// Logout 退出（token 由 Auth 中间件解析并放入上下文）
func (h *ApiAuthHandler) Logout(c *gin.Context) {
		if err := service.Logout(c.GetString("token")); err != nil {
		c.JSON(400, gin.H{
			"code":     400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Refresh 刷新 token（旧 token 作废，返回新 token）
func (h *ApiAuthHandler) Refresh(c *gin.Context) {
		token, err := service.Refresh(c.GetString("token"))
	if err != nil {
		c.JSON(400, gin.H{
			"code":     400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}
