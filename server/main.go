package main

import (
	"hxzl/apiInstall"
	"hxzl/apiMenu"
	"hxzl/apiTest"
	"hxzl/db"
	"hxzl/env"
	"hxzl/home"
	"hxzl/logx"
	"hxzl/motd"
	"hxzl/server"
	"hxzl/web"
	// "hxzl/ws"
)

func main() {
	motd.Start()
	env.Start()
	logx.Start()
	db.Start()
	web.Start()
	// ws.Start()
	home.Start()
	apiInstall.Start()
	apiMenu.Start()
	apiTest.Start()
	server.Start()
}
