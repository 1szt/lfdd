package main

import (
	"hxzl/core/db"
	"hxzl/core/env"
	"hxzl/core/logx"
	"hxzl/core/motd"
	"hxzl/core/server"
	"hxzl/core/web"
	"hxzl/model"
	"hxzl/router"
)

func main() {
	motd.Start()
	logx.Start()
	env.Start()
	db.Start()
	model.Start()
	web.Start()
	router.Start()
	server.Start()
}
