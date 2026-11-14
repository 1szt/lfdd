package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

// apiorder 订单接口：下单/查询均免登录（OptionalGuard，暂无用户身份区分）
func apiOrderRouter() {
	apiOrderHandler := &handler.ApiOrderHandler{}
	group := web.Engine.Group("/api/order")
	group.POST("/create", middleware.OptionalGuard(), apiOrderHandler.Create)
	group.POST("/finish", middleware.OptionalGuard(), apiOrderHandler.Finish)
	group.POST("/price", middleware.OptionalGuard(), apiOrderHandler.ChangePrice)
	group.POST("/void", middleware.OptionalGuard(), apiOrderHandler.Void)
	group.POST("/delete", middleware.OptionalGuard(), apiOrderHandler.Delete)
	group.GET("/page", middleware.OptionalGuard(), apiOrderHandler.Page)
}
