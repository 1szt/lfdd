package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

// apiai AI 设置接口：读取/保存/测试连通性，仅管理员（RoleGuard）
func apiAiRouter() {
	apiAiHandler := &handler.ApiAiHandler{}
	group := web.Engine.Group("/api/ai")

	group.GET("/settings", middleware.RoleGuard(), apiAiHandler.Settings)
	group.POST("/save", middleware.RoleGuard(), apiAiHandler.Save)
	group.POST("/test", middleware.RoleGuard(), apiAiHandler.Test)
}
