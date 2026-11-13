package router

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

var errorMessages = map[int]map[string]string{
	1001: {
		"en": "Database operation failed",
		"zh": "数据库操作失败",
	},
	1002: {
		"en": "Unauthorized or expired",
		"zh": "未授权或登录过期",
	},
	1003: {
		"en": "Invalid request parameters",
		"zh": "请求参数错误",
	},
}

// langOf 解析 Accept-Language 头，返回受支持的语言代码（如 "zh-CN;q=0.9" -> "zh"）。
// 不支持的语言统一回退到英文。
func langOf(c *gin.Context) string {
	lang := strings.ToLower(c.GetHeader("Accept-Language"))
	if i := strings.IndexAny(lang, ",;"); i >= 0 {
		lang = lang[:i]
	}
	lang = strings.TrimSpace(lang)
	if strings.HasPrefix(lang, "zh") {
		return "zh"
	}
	return "en"
}

// message 按 语言 -> 英文 -> 兜底 的顺序查找错误文案。
// code 未注册时直接返回兜底文案，避免依赖 map 的零值行为。
func message(c *gin.Context, code int) string {
	msgs := errorMessages[code]
	if len(msgs) == 0 {
		return "Unknown error"
	}
	if msg, ok := msgs[langOf(c)]; ok {
		return msg
	}
	if msg, ok := msgs["en"]; ok {
		return msg
	}
	return "Unknown error"
}

// Fail 统一的错误响应：HTTP 200 + 业务错误码，并记录日志。
func Fail(c *gin.Context, code int) {
	msg := message(c, code)

	// 统一在这里记录日志（不依赖调用方）
	log.Printf("[API Error] Code: %d, Message: %s, Path: %s %s", code, msg, c.Request.Method, c.Request.URL.Path)

	c.JSON(200, gin.H{
		"code":    code,
		"message": msg,
	})
}
