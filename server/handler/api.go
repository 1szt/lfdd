// Package handler HTTP 处理层：绑定请求参数 → 调用 service → 返回响应，不含路由注册（路由在 router 模块）。
package handler

import "github.com/gin-gonic/gin"

// GetAPI API 根路径探活
func GetAPI(c *gin.Context) {
	c.String(200, "Hello, API!")
}
