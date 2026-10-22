// Package web 中间件初始化模块：注册全局中间件，必须在模块注册路由前调用。
package web

import (
	"hxzl/server"

	g "github.com/gin-gonic/gin"
)

// Start 注册全局中间件（Recovery + Logger）
func Start() {
	server.Router.Use(g.Recovery())
	server.Router.Use(g.Logger())
}
