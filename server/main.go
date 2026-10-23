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
