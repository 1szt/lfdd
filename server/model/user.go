package model

import "time"

// User 用户表
type User struct {
	ID        int64     `json:"id,string" gorm:"primaryKey;autoIncrement:false"`
	Username  string    `json:"username" gorm:"size:64;uniqueIndex"`
	Email     string    `json:"email" gorm:"size:128;uniqueIndex"`
	Password  string    `json:"-"` // 密码永不序列化到响应
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
