package main

import (
	"server/core/db"
	"server/core/env"
	"server/core/logx"
	"server/core/motd"
	"server/core/server"
	"server/core/setup"
	"server/core/web"
	"server/model"
	"server/router"
)

func main() {
	motd.Start()
	logx.Start()
	env.Start()
	db.Start()
	model.Start()
	setup.Start()
	web.Start()
	router.Start()
	server.Start()
}
