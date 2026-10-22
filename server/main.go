package main

import (
	"hxzl/api"
	"hxzl/config"
	"hxzl/gorm"
	"hxzl/home"
	"hxzl/motd"
	"hxzl/server"
	"hxzl/web"
)

func main() {
	motd.Start()
	config.Start()
	gorm.Start()

	web.Start()
	home.Start()
	api.Start()
	server.Start()
}
