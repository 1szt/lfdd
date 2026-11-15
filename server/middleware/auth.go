package middleware

import (
	"errors"
	"strings"

	"server/core/db"
	"server/model"
	"server/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============ 登录拦截系列中间件（AuthGuard / RoleGuard / GuestGuard / IdentityGuard） ============

// GuestTokenFromHeader 从 Authorization 解析游客 token：返回游客身份，ok=false 表示不是游客
func GuestTokenFromHeader(c *gin.Context) (guestToken string, ok bool) {
	token := TokenFromHeader(TokenFromHeaderInput{Authorization: c.GetHeader("Authorization")}).Token
	if !strings.HasPrefix(token, model.GuestTokenPrefix) {
		return "", false
	}
	return token, true
}

// TokenFromHeaderInput 输入
type TokenFromHeaderInput struct {
	Authorization string `json:"authorization"` // 请求原始 Authorization 头
}

// TokenFromHeaderOutput 输出
type TokenFromHeaderOutput struct {
	Token string `json:"token"` // 解析出的 token；空串表示未携带 Bearer 凭证
}

// TokenFromHeader 解析 Authorization: Bearer <token>（只认 Authorization 头，避免读 body 破坏后续参数绑定）
func TokenFromHeader(i TokenFromHeaderInput) (o TokenFromHeaderOutput) {
	if !strings.HasPrefix(i.Authorization, "Bearer ") {
		return o
	}
	o.Token = strings.TrimSpace(strings.TrimPrefix(i.Authorization, "Bearer "))
	return o
}

// AuthGuard 登录拦截中间件：**登录放行，未登录阻止**。
// 解析 Authorization: Bearer <token> → 校验会话 → 通过后把 userID、token 写入上下文
// （handler 用 c.GetInt64("userID") / c.GetString("token") 取）；
// 未登录/会话失效返回 401 并中断请求，数据库异常返回 500。
// 用于必须登录才能访问的接口。
func AuthGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := TokenFromHeader(TokenFromHeaderInput{Authorization: c.GetHeader("Authorization")}).Token
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"message": "未登录"})
			return
		}

		// 校验会话（GetSession 内部已过滤过期时间）
		s, err := service.GetSession(service.GetSessionInput{Token: token})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(401, gin.H{"message": "登录已过期或会话无效"})
				return
			}
			c.AbortWithStatusJSON(500, gin.H{"message": "服务器内部错误"})
			return
		}

		// 登录放行：把 userID、token 写入上下文供 handler 使用
		c.Set("userID", s.Session.UserID)
		c.Set("token", token)
		c.Next()
	}
}

// RoleGuard 管理员拦截中间件：**管理员（role=2）放行，其余阻止**。
// 先做 AuthGuard 的登录校验（userID/token 写入上下文），再查用户角色：
// - 未登录 → 401；已登录但非管理员 → 403；管理员 → 放行。
// 用于用户管理、AI 设置等仅管理员可用的接口。
func RoleGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := TokenFromHeader(TokenFromHeaderInput{Authorization: c.GetHeader("Authorization")}).Token
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"message": "未登录"})
			return
		}

		s, err := service.GetSession(service.GetSessionInput{Token: token})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(401, gin.H{"message": "登录已过期或会话无效"})
				return
			}
			c.AbortWithStatusJSON(500, gin.H{"message": "服务器内部错误"})
			return
		}

		// 查角色：须为管理员（role=2）
		var u model.User
		if err = db.DB.First(&u, "id = ?", s.Session.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(401, gin.H{"message": "用户不存在"})
				return
			}
			c.AbortWithStatusJSON(500, gin.H{"message": "服务器内部错误"})
			return
		}
		if u.Role != 2 {
			c.AbortWithStatusJSON(403, gin.H{"message": "需要管理员权限"})
			return
		}

		c.Set("userID", s.Session.UserID)
		c.Set("token", token)
		c.Next()
	}
}

// GuestGuard 未登录拦截中间件（游客中间件）：**未登录放行，已登录阻止**。
// 用于登录、注册等已登录用户反而不应访问的页面：
// - 未携带 token、携带游客 token，或会话不存在/已过期 → 视为未登录，放行；
// - 会话有效（已登录）→ 403 中断请求。
func GuestGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := TokenFromHeader(TokenFromHeaderInput{Authorization: c.GetHeader("Authorization")}).Token
		if token == "" || strings.HasPrefix(token, model.GuestTokenPrefix) {
			// 未登录（或携带的是游客 token）→ 放行
			c.Next()
			return
		}

		if _, err := service.GetSession(service.GetSessionInput{Token: token}); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 会话失效 → 视为未登录，放行
				c.Next()
				return
			}
			c.AbortWithStatusJSON(500, gin.H{"message": "服务器内部错误"})
			return
		}

		// 已登录 → 阻止
		c.AbortWithStatusJSON(403, gin.H{"message": "已登录，禁止访问该页面"})
	}
}

// IdentityGuard 身份拦截中间件：**登录身份或有效游客身份放行，两者都没有则阻止**。
// 1. Authorization 中的 guest_ 前缀 token → 游客（须为服务端签发且未注销）；
// 2. 其他 token → 登录身份（会话有效才放行，失效返回 401）；
// 3. 未携带 token / 游客身份无效 → 401 + need_guest=true，前端据此重新申请游客身份后重试。
// 用于所有业务接口：未申请游客身份前不返回任何业务数据。
func IdentityGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := TokenFromHeader(TokenFromHeaderInput{Authorization: c.GetHeader("Authorization")}).Token
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"message":    "未登录或游客身份无效，请先申请游客身份",
				"need_guest": true,
			})
			return
		}

		// 1. 游客身份：必须是服务端签发且未注销的
		if strings.HasPrefix(token, model.GuestTokenPrefix) {
			if _, err := service.GetGuest(service.GetGuestInput{GuestID: token}); err == nil {
				c.Set("guestID", token)
				c.Set("isGuest", true)
				c.Next()
				return
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(500, gin.H{"message": "服务器内部错误"})
				return
			}
			c.AbortWithStatusJSON(401, gin.H{
				"message":    "游客身份无效，请重新申请",
				"need_guest": true,
			})
			return
		}

		// 2. 登录身份
		s, err := service.GetSession(service.GetSessionInput{Token: token})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(401, gin.H{"message": "登录已过期或会话无效"})
				return
			}
			c.AbortWithStatusJSON(500, gin.H{"message": "服务器内部错误"})
			return
		}

		c.Set("userID", s.Session.UserID)
		c.Set("token", token)
		c.Next()
	}
}
