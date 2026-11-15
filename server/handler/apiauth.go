package handler

import (
	"server/middleware"
	"server/service"

	"github.com/gin-gonic/gin"
)

type ApiAuthHandler struct {
}

// Guest 申请游客身份：未登录用户自动调用，返回的 token 由前端保存并在 Authorization 头中携带
func (h *ApiAuthHandler) Guest(c *gin.Context) {
	out, err := service.CreateGuest(service.CreateGuestInput{
		Device: c.GetHeader("User-Agent"),
	})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, gin.H{
		"guest_id": out.GuestID,
	})
}

// RevokeGuest 注销游客身份：主动注销后该身份立即失效，业务接口需重新申请
func (h *ApiAuthHandler) RevokeGuest(c *gin.Context) {
	guestToken, ok := middleware.GuestTokenFromHeader(c)
	if !ok {
		c.JSON(400, gin.H{"message": "缺少游客身份"})
		return
	}

	if _, err := service.DeleteGuest(service.DeleteGuestInput{GuestID: guestToken}); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{})
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
	out, err := service.Login(service.LoginInput{
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

	// 登录成功：作废本次携带的游客身份（登录后一律以登录身份为准）
	if guestToken, ok := middleware.GuestTokenFromHeader(c); ok {
		_, _ = service.DeleteGuest(service.DeleteGuestInput{GuestID: guestToken})
	}

	c.JSON(200, gin.H{
		"token": out.Token,
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
		"token": token.Token,
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
