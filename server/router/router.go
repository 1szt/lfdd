// Package handler HTTP 处理层：路由注册与处理函数同包，一功能一文件。
// 下线某功能 = 删除对应文件 + 删除 Start 中一行调用。
package router

import (
	"log"
)

// Start 注册全部路由（须在 web.Start 之后、server.Start 之前调用）
func Start() {
	api()
	apiauth()
	apimenu()
	apiinstall()
	apitest()
	apiws()

	log.Print("✅ [HANDLER] 路由注册 加载完成！")
}
