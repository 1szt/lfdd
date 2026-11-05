package model

import "time"

// User 用户表
type User struct {
	// 内部主键，由业务层生成（毫秒时间戳 + 随机数），仅系统内部关联使用
	ID int64 `json:"id,string" gorm:"primaryKey;autoIncrement:false"`

	// 用户 ID：对外唯一标识，同时作为登录名
	Username string `json:"username" gorm:"size:32;not null;uniqueIndex"`

	// 邮箱，全局唯一（写入前统一转小写）
	Email string `json:"email" gorm:"size:128;not null;uniqueIndex"`

	// 密码哈希（argon2id PHC 自描述编码串，约 96 字符），永不序列化
	Password string `json:"-" gorm:"size:255;not null"`

	// 名字，展示用，可重复、可空
	Name string `json:"name" gorm:"size:64"`

	// 头像地址，可空
	Avatar string `json:"avatar" gorm:"size:255"`

	// 账号状态：1 正常、0 禁用（默认 1）
	Status int8 `json:"status" gorm:"not null;default:1"`

	// 账号角色/权限等级：1 普通用户、2 管理员（默认 1）
	Role int8 `json:"role" gorm:"not null;default:1"`

	// 最近登录时间，从未登录为 null
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// 创建/更新时间，由 gorm 自动维护
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User 表结构
func (User) TableName() string {
	return "user"
}
