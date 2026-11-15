package middleware

import (
	"net/http"
	"strings"

	"github.com/1szt/go-pkg/env"
	"github.com/gin-gonic/gin"
)

// Cors 跨域中间件：前后端分域部署时放行浏览器请求。
// 鉴权走 Authorization 头（不使用 Cookie），因此允许来源可为 *。
// 允许来源由 LFDD_CORS_ORIGINS 配置：* 表示全部；也可填逗号分隔的域名白名单。
func Cors() gin.HandlerFunc {
	allowAll := false
	allowed := map[string]bool{}
	for _, origin := range strings.Split(env.Get("LFDD_CORS_ORIGINS"), ",") {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			allowAll = true
		}
		allowed[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			switch {
			case allowAll:
				c.Header("Access-Control-Allow-Origin", "*")
			case allowed[origin]:
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		// 预检请求直接返回 204，不进入后续鉴权/业务中间件
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
