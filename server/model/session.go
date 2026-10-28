package model

import "time"

// Session 用户会话表：存储登录态（refresh token），支持多设备登录与退出。
type Session struct {
	ID        int64     `json:"id,string" gorm:"primaryKey;autoIncrement:false"`
	Token     string    `json:"-" gorm:"size:128;uniqueIndex"` // refresh token，唯一索引，不序列化
	UserID    int64     `json:"user_id,string" gorm:"index"`   // 外键，关联 users.id
	Device    string    `json:"device" gorm:"size:128"`        // 设备标识（UA/IP，可空）
	ExpiresAt time.Time `json:"expires_at" gorm:"index"`       // 过期时间，过期即失效
	CreatedAt time.Time `json:"created_at"`
}
