// Package setup 启动初始化模块：首次启动创建管理员账号、定期清理过期登录会话。
// 须在 db.Start 与 model.Start 之后调用。
package setup

import (
	"log"
	"time"

	"server/service"
)

// Start 启动时初始化：创建首个管理员 + 开启过期会话清理
func Start() {
	ensureAdmin()
	// 过期会话清理：启动时清一次，之后每小时一次（避免 session 表无限增长）
	go cleanSessionTask()
	log.Print("✅ [SETUP] 启动初始化 加载完成！")
}

// ensureAdmin 首次启动（库里没有用户）自动创建管理员：随机邮箱 + 随机密码，凭据仅打印一次
func ensureAdmin() {
	out, err := service.EnsureAdmin(service.EnsureAdminInput{})
	if err != nil {
		log.Printf("⚠️ [SETUP] 初始化管理员失败: %v", err)
		return
	}
	if out.Created {
		log.Printf("🔑 [SETUP] 首次启动已创建管理员 -> 账号: %s  密码: %s（请立即保存，之后可在管理页修改）",
			out.Email, out.Password)
	}
}

// cleanSessionTask 定期删除已过期的登录会话
func cleanSessionTask() {
	for {
		if err := service.CleanExpiredSessions(service.CleanExpiredSessionsInput{}); err != nil {
			log.Printf("⚠️ [SETUP] 清理过期会话失败: %v", err)
		}
		time.Sleep(time.Hour)
	}
}
