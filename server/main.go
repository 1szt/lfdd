// @title lfdd 管理系统 API
// @version 1.0
// @description lfdd 后端接口文档（swaggo 自动生成）。修改 handler 注释后执行 `swag init` 更新。
// @BasePath /api
package main

import (
	"hxzl/api"
	"hxzl/config"
	"hxzl/db"
	"hxzl/home"
	"hxzl/motd"
	"hxzl/server"
	"hxzl/web"

	_ "hxzl/docs" // 导入 swaggo 生成的接口文档（swag init 生成）
)

func main() {
	motd.Start()
	config.Start()
	db.Start()

	web.Start()
	home.Start()
	api.Start()
	server.Start()
}
