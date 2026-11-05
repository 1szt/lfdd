package model

import "time"

// Session 用户会话表：一次登录一条记录（refresh token），支持多设备登录与退出。
type Session struct {
	// 会话 token（refresh token），直接作为主键：全局唯一，永不序列化
	Token string `json:"-" gorm:"primaryKey;size:128"`

	// 所属用户的内部 ID（user.id），普通索引：同一用户可同时存在多条会话
	UserID int64 `json:"user_id,string" gorm:"not null;index"`

	// 登录端 IP，IPv6 最长 45 字符，可空
	IP string `json:"ip" gorm:"size:45"`

	// 设备信息（UA 等），可空
	Device string `json:"device" gorm:"size:255"`

	// 过期时间，过期即失效；查询有效会话时按此过滤
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`

	// 创建/更新时间，由 gorm 自动维护
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Session 表结构
func (Session) TableName() string {
	return "session"
}
