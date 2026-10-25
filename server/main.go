package main

import (
	"hxzl/api"
	"hxzl/apiinstall"
	"hxzl/apimenu"
	"hxzl/apitest"
	"hxzl/apiauth"
	"hxzl/apiws"
	"hxzl/db"
	"hxzl/env"
	"hxzl/logx"
	"hxzl/motd"
	"hxzl/server"
	"hxzl/web"
)

func main() {
	motd.Start()
	logx.Start()
	env.Start()
	db.Start()
	web.Start()
	api.Start()
	apiws.Start()
	apiinstall.Start()
	apiauth.Start()
	apimenu.Start()
	apitest.Start()
	server.Start()
}
