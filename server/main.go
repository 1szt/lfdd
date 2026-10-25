package main

import (
	"hxzl/api"
	"hxzl/apiInstall"
	"hxzl/apiMenu"
	"hxzl/apiTest"
	"hxzl/apiWS"
	"hxzl/db"
	"hxzl/env"
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
	api.Start()
	apiWS.Start()
	apiInstall.Start()
	apiMenu.Start()
	apiTest.Start()
	server.Start()
}
