package model

import "time"

// Menu 菜单表
type Menu struct {
	// 自增主键
	ID int64 `json:"id,string" gorm:"primaryKey;autoIncrement"`

	// 菜单名称
	Name string `json:"name" gorm:"size:64;not null"`

	// 同级菜单排序，数值越小越靠前；前端拖拽后由业务层整组重写
	Sort int `json:"sort" gorm:"not null;default:0"`

	// 创建/更新时间，由 gorm 自动维护
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Menu 表结构
func (Menu) TableName() string {
	return "menu"
}
