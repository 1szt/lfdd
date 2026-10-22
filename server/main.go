package main

import (
	"hxzl/api"
	"hxzl/config"
	"hxzl/db"
	"hxzl/home"
	"hxzl/motd"
	"hxzl/server"
	"hxzl/web"
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
