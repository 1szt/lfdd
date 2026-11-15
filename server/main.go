package main

import (
	"log"
	"server/core/db"
	"server/core/env"
	"server/core/logx"
	"server/core/motd"
	"server/core/server"
	"server/core/web"
	"server/model"
	"server/router"
	"server/service"
	"time"
)

func main() {
	motd.Start()
	logx.Start()
	env.Start()
	db.Start()
	model.Start()
	// 首次启动（库里没有用户）自动创建管理员：随机邮箱 + 随机密码，凭据仅打印一次
	if out, err := service.EnsureAdmin(service.EnsureAdminInput{}); err != nil {
		log.Printf("⚠️ [MAIN] 初始化管理员失败: %v", err)
	} else if out.Created {
		log.Printf("🔑 [MAIN] 首次启动已创建管理员 -> 账号: %s  密码: %s（请立即保存，之后可在管理页修改）",
			out.Email, out.Password)
	}
	// 过期会话清理：启动时清一次，之后每小时一次（避免 session 表无限增长）
	go cleanSessionTask()
	web.Start()
	router.Start()
	server.Start()
}

// cleanSessionTask 定期删除已过期的登录会话
func cleanSessionTask() {
	for {
		if err := service.CleanExpiredSessions(service.CleanExpiredSessionsInput{}); err != nil {
			log.Printf("⚠️ [MAIN] 清理过期会话失败: %v", err)
		}
		time.Sleep(time.Hour)
	}
}
