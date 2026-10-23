// Package middleware 全局中间件集合：一个中间件一个文件，由 web.Start 统一装配。
package middleware

import (
	"hxzl/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// initFreePaths 未初始化时精确放行的路径
var initFreePaths = []string{
	"/", // 根路径（安装页入口/健康检查）
}

// initFreePrefixes 未初始化时按前缀放行的路径
var initFreePrefixes = []string{
	"/api/install/", // 安装接口自身必须可访问（check/submit）
	"/swagger",      // 接口文档页免检（开发期需要在未初始化时查阅文档）
}

// InitGuard 全局初始化检查中间件：server_install != true 时，
// 拦截白名单以外的所有请求，返回统一的未初始化提示。
func InitGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 已初始化直接放行
		if config.Get("server_install").Value == "true" {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		for _, p := range initFreePaths {
			if path == p {
				c.Next()
				return
			}
		}
		for _, p := range initFreePrefixes {
			if strings.HasPrefix(path, p) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"code":         40001,
			"message":      "服务器未初始化，请先完成初始化",
			"need_install": true,
		})
	}
}
