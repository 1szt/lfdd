package router

import (
	"hxzl/core/web"
	"hxzl/handler"
)

// apiitest 热重载测试页
func apiitest() {
	test := web.Engine.Group("/api")
	test.GET("/test/reload", handler.GetPage)
	test.GET("/test/reload/status", handler.GetStatus)
	test.POST("/test/reload/apply", handler.PostApply)
}
