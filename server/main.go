package main

import (
	"hxzl/api"
	"hxzl/config"
	"hxzl/gin"
	"hxzl/gorm"
	"hxzl/home"
	"hxzl/initpkg"
	"hxzl/motd"
)

func main() {
	motd.Start()
	config.Start()
	gorm.Start()
	initpkg.Start()
	home.Start()
	api.Start()
	gin.Start()
}
