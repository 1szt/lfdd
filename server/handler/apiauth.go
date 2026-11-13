package handler

import (
	"hxzl/service"

	"github.com/gin-gonic/gin"
)

type ApiAuthHandler struct {
}

// Login 登录
func (h *ApiAuthHandler) Login(c *gin.Context) {
	var req struct {
		Account  string `json:"account"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	//使用新的LoginInput结构体
	token, err := service.Login(service.LoginInput{
		Account:  req.Account,
		Password: req.Password,
		IP:       c.ClientIP(),
		Device:   c.GetHeader("User-Agent"),
	})
	if err != nil {
		c.JSON(400, gin.H{
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	token, err := service.Register(service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		IP:       c.ClientIP(),
		Device:   c.GetHeader("User-Agent"),
	})
	if err != nil {
		c.JSON(400, gin.H{
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
	_, err := service.Logout(service.LogoutInput{
		Token: c.GetString("token"),
	})
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	//Logout方法返回格式不一致，缺少code字段
	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Refresh 刷新 token（旧 token 作废，返回新 token）
func (h *ApiAuthHandler) Refresh(c *gin.Context) {
	token, err := service.Refresh(service.RefreshInput{
		Token: c.GetString("token"),
	})
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"token": token.Token,
	})
}
