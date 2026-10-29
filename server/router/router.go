// Package router 路由层：统一注册全部 HTTP 路由，handler 只提供处理函数。
// 路由一功能一文件；下线某功能 = 删除对应文件 + 删除 Start 中一行调用。
package router

import (
	"log"
)

// Start 注册全部路由（须在 web.Start 之后、server.Start 之前调用）
func Start() {
	api()
	apiauth()
	apimenu()
	apinstall()
	apiitest()
	apiws()
	log.Print("✅ [ROUTER] 路由注册 加载完成！")
}
