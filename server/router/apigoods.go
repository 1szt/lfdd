package router

import (
	"server/core/web"
	"server/handler"
	"server/middleware"
)

// apigoods 商品接口：侧边菜单 + 商品页面 + 图片直出（须登录或持有有效游客身份，IdentityGuard），
// 新建/修改/删除/顺序管理/图片上传仅管理员（RoleGuard）
func apiGoodsRouter() {
	apiGoodsHandler := &handler.ApiGoodsHandler{}
	group := web.Engine.Group("/api/goods")

	// 商品页（登录用户 / 游客）
	group.GET("/sidebar", middleware.IdentityGuard(), apiGoodsHandler.Sidebar)
	group.GET("/page", middleware.IdentityGuard(), apiGoodsHandler.Page)
	group.GET("/all", middleware.IdentityGuard(), apiGoodsHandler.All)

	// 商品管理（仅管理员：RoleGuard）
	group.POST("/create", middleware.RoleGuard(), apiGoodsHandler.Create)
	group.POST("/update", middleware.RoleGuard(), apiGoodsHandler.Update)
	group.POST("/delete", middleware.RoleGuard(), apiGoodsHandler.Delete)
	group.POST("/sort", middleware.RoleGuard(), apiGoodsHandler.Sort)
	group.POST("/upload", middleware.RoleGuard(), apiGoodsHandler.Upload)
	group.POST("/match", middleware.IdentityGuard(), apiGoodsHandler.Match)

	// 图片直出（免身份）：GET /api/upload/{hash}{ext}（goods.image 存的就是这个路径）
	// 浏览器的 <img> 请求无法携带认证头，故不做身份拦截
	web.Engine.GET("/api/upload/:name", apiGoodsHandler.Image)
}
