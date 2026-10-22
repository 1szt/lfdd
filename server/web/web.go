// Package init 中间件初始化模块：注册全局中间件，必须在模块注册路由前调用。
package web

import (
	"hxzl/gin"

	g "github.com/gin-gonic/gin"
)

// Start 注册全局中间件（Recovery + Logger）
func Start() {
	gin.Router.Use(g.Recovery())
	gin.Router.Use(g.Logger())
}
