package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

// apiorder 订单接口：须登录或持有有效游客身份（IdentityGuard）。
// 游客：只能查自己的未结单，结单/改价/废单/删除一律拒绝；登录用户忽略游客身份。
func apiOrderRouter() {
	apiOrderHandler := &handler.ApiOrderHandler{}
	group := web.Engine.Group("/api/order")
	group.POST("/create", middleware.IdentityGuard(), apiOrderHandler.Create)
	group.POST("/finish", middleware.IdentityGuard(), apiOrderHandler.Finish)
	group.POST("/price", middleware.IdentityGuard(), apiOrderHandler.ChangePrice)
	group.POST("/void", middleware.IdentityGuard(), apiOrderHandler.Void)
	group.POST("/delete", middleware.IdentityGuard(), apiOrderHandler.Delete)
	group.GET("/page", middleware.IdentityGuard(), apiOrderHandler.Page)
	// 订单搜索（单号/日期）仅管理员可用
	group.GET("/search", middleware.RoleGuard(), apiOrderHandler.Search)
}
