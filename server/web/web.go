// Package web web 框架初始化模块：持有全局 Engine 并注册全局中间件，必须在模块注册路由前调用。
package web

import (
	"github.com/gin-gonic/gin"
)

// Engine 全局路由引擎，各业务模块通过它注册自己的路由
var Engine = gin.New()

// Start 注册全局中间件（Recovery + Logger）
func Start() {
	Engine.Use(gin.Recovery())
	Engine.Use(gin.Logger())
}
