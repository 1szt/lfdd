package main

import (
	"hxzl/api"
	"hxzl/config"
	"hxzl/gin"
	"hxzl/gorm"
	"hxzl/home"
	"hxzl/web"
	"hxzl/motd"
)

func main() {
	motd.Start()
	config.Start()
	gorm.Start()

	web.Start()
	home.Start()
	api.Start()
	gin.Start()
}
