package middleware

import (
	"errors"
	"strings"

	"hxzl/service"

	"github.com/gin-gonic/gin"
)

// Auth 登录鉴权中间件：解析 Authorization: Bearer <token> → 校验会话 →
// 通过后把 userID、token 写入上下文（handler 用 GetUserID / c.GetString("token") 取），
// 未通过返回对应状态码并中断请求。
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析 token（只认 Authorization 头，避免读 body 破坏后续参数绑定）
		token := ""
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		}

		userID, err := service.SessionUserID(token)
		if err != nil {
			code := 500
			var be *service.APIError
			if errors.As(err, &be) {
				code = be.Code
			}
			c.AbortWithStatusJSON(code, gin.H{
				"err":     code,
				"message": err.Error(),
			})
			return
		}

		c.Set("userID", userID)
		c.Set("token", token)
		c.Next()
	}
}