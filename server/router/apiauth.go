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

}

// Register 注册
func (h *ApiAuthHandler) Register(c *gin.Context) {

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{
			"err":     400,
			"message": "请求参数错误",
		})
		return
	}

	token, err := service.Register(service.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		IP:       c.ClientIP(),
		Device:   c.GetHeader("User-Agent"),
	})
	if err != nil {
		c.JSON(500, gin.H{
			"err":     500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}

// Logout 退出
func (h *ApiAuthHandler) Logout(c *gin.Context) {

}

// Refresh 刷新 token
func (h *ApiAuthHandler) Refresh(c *gin.Context) {

}
