package model

import "time"

// GuestTokenPrefix 游客 token 前缀：游客与登录用户共用 Authorization: Bearer <token> 头，
// 服务端按前缀分流（guest_ 开头查游客表，其余查会话表）
const GuestTokenPrefix = "guest_"

// Guest 游客身份表：未登录用户申请游客身份时生成一条，之后通过 Authorization 头携带该 token。
// 登录后该身份作废（登录时主动删除），业务接口一律以登录身份为准。
type Guest struct {
	// 游客身份（随机 token），直接作为主键：全局唯一，永不序列化到订单里
	ID string `json:"-" gorm:"primaryKey;size:64"`

	// 创建设备信息（UA 等），便于排查，可空
	Device string `json:"device" gorm:"size:255"`

	// 创建时间，由 gorm 自动维护
	CreatedAt time.Time `json:"created_at"`
}

// Guest 表结构
func (Guest) TableName() string {
	return "guests"
}
