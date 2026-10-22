// Package api 接口模块：HTTP 接口的总入口，创建 /api 路由组并汇总注册各业务子包。
// 约定：每个业务域一个子包（如 install、menu），子包只导出 Register(rg) 注册自己的路由。
package api

import (
	"hxzl/api/install"
	"hxzl/api/menu"
	"hxzl/web"
	"log"

	"github.com/gin-gonic/gin"
)

func Start() {
	apiRoute := web.Engine.Group("/api")
	apiRoute.GET("/", home)

	// 各业务子包注册自己的路由
	install.Register(apiRoute)
	menu.Register(apiRoute)

	log.Print("✅ [API] API模块 加载完成！")
}

func home(c *gin.Context) {
	c.String(200, "API")
}
