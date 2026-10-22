package main

import (
	"hxzl/api"
	"hxzl/config"
	"hxzl/gin"
	"hxzl/gorm"
	"hxzl/home"
	"hxzl/initmod"
	"hxzl/motd"
)

func main() {
	motd.Start()
	config.Start()
	gorm.Start()
	initmod.Start()
	home.Start()
	api.Start()
	gin.Start()
}
