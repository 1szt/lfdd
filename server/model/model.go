// Package model 数据模型层：定义各表结构（gorm，一表一文件），负责启动时自动迁移。
package model

import (
	"log"
	"server/core/db"
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
		&Guest{},
		&Menu{},
		&Goods{},
		&Order{},
		&OrderItem{},
		&SysConfig{},
	}

	// 2. 调用 AutoMigrate 方法迁移所有模型
	if err := db.DB.AutoMigrate(models...); err != nil {
		log.Printf("❌ [MODEL] 表结构迁移失败: %v", err)
		return
	}

	// 3. 自愈：库里一个管理员都没有时，把最早注册的用户提为管理员（role=2）
	//    （role 字段后补的老数据默认为 1，否则无人能进管理功能）
	var admins, users int64
	db.DB.Model(&User{}).Where("role = ?", 2).Count(&admins)
	if admins == 0 {
		if db.DB.Model(&User{}).Count(&users); users > 0 {
			var oldest User
			if err := db.DB.Order("created_at ASC").First(&oldest).Error; err == nil {
				db.DB.Model(&User{}).Where("id = ?", oldest.ID).Update("role", 2)
				log.Printf("✅ [MODEL] 已将最早用户 %s 提升为管理员", oldest.Username)
			}
		}
	}

	log.Print("✅ [MODEL] 数据模型 迁移完成！")
}
