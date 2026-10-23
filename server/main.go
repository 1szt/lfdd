// @title lfdd 管理系统 API
// @version 1.0
// @description lfdd 后端接口文档（swaggo 自动生成）。修改 handler 注释后执行 `swag init` 更新。
// @BasePath /api
package main

import (
	"hxzl/api"
	"hxzl/db"
	"hxzl/env"
	"hxzl/home"
	"hxzl/logx"
	"hxzl/motd"
	"hxzl/server"
	"hxzl/web"
)

func main() {
	motd.Start()
	env.Start()
	logx.Start()
	db.Start()
	web.Start()
	home.Start()
	api.Start()
	server.Start()
}
