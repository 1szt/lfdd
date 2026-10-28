// Package model 数据模型层：定义各表结构（gorm，一表一文件），负责启动时自动迁移。
package model

import (
	"hxzl/db"
	"log"
	"math/rand/v2"
	"time"
)

// Start 执行所有表的自动迁移（须在 db.Start 之后调用）
func Start() {
	if db.DB == nil {
		log.Println("⚠️ [MODEL] 数据库未连接，跳过表结构迁移")
		return
	}

	if err := db.DB.AutoMigrate(
		&User{},
		&Session{},
	); err != nil {
		log.Printf("❌ [MODEL] 表结构迁移失败: %v", err)
		return
	}

	log.Print("✅ [MODEL] 数据模型 迁移完成！")
}

// NewID 生成 18 位数字 ID：毫秒时间戳(13位) + 随机数(5位)，不依赖数据库自增。
// 多实例部署时如出现冲突，再升级为雪花算法。
func NewID() int64 {
	return time.Now().UnixMilli()*100_000 + int64(rand.IntN(100_000))
}
