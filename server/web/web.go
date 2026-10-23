// Package web web 框架初始化模块：持有全局 Engine 并注册全局中间件，必须在模块注册路由前调用。
package web

import (
	"hxzl/web/middleware"
	"os"

	"github.com/gin-gonic/gin"
)

func init() {
	// LFDD_MODE=release 时关闭 gin 的调试输出（路由打印、开发警告等）
	if os.Getenv("LFDD_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
}

// Engine 全局路由引擎，各业务模块通过它注册自己的路由
var Engine = gin.New()

// Start 注册全局中间件（Logger + Recovery + 初始化检查）
// 必须在所有业务模块注册路由之前调用（gin 中间件对之后注册的路由才生效）
func Start() {
	Engine.Use(gin.Logger(), gin.Recovery(), middleware.InitGuard())
}
