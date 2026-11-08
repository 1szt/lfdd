// Package model 数据模型层：定义各表结构（gorm，一表一文件），负责启动时自动迁移。
package model

import (
	"hxzl/core/db"
	"log"
)

// Start 执行所有表的自动迁移（须在 db.Start 之后调用）
func Start() {
	if db.DB == nil {
		log.Println("⚠️ [MODEL] 数据库未连接，跳过表结构迁移")
		return
	}

	// 1. 将所有需要迁移的模型放入一个切片中
	models := []any{
		&User{},
		&Session{},
		&Menu{},
		&Goods{},
	}

	// 2. 调用 AutoMigrate 方法迁移所有模型
	if err := db.DB.AutoMigrate(models...); err != nil {
		log.Printf("❌ [MODEL] 表结构迁移失败: %v", err)
		return
	}

	log.Print("✅ [MODEL] 数据模型 迁移完成！")
}
