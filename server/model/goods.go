package model

import "time"

// Goods 商品表：挂在菜单（分类）下的商品
type Goods struct {
	// 自增主键
	ID int64 `json:"id,string" gorm:"primaryKey;autoIncrement"`

	// 所属菜单 ID（menu.id），普通索引：同一菜单下可有多个商品
	MenuID int64 `json:"menu_id,string" gorm:"not null;index"`

	// 商品名称
	Name string `json:"name" gorm:"size:64;not null"`

	// 商品别名（多个用逗号分隔），供 AI 点单匹配时一起识别，可空
	Alias string `json:"alias" gorm:"size:255"`

	// 商品价格，单位：分（1850 = 18.50 元），避免浮点精度问题
	Price int `json:"price" gorm:"not null;default:0"`

	// 商品进价（成本价），单位：分；仅管理员可见
	Cost int `json:"cost" gorm:"not null;default:0"`

	// 商品规格（如"大杯/中杯"、"500g"），可空
	Spec string `json:"spec" gorm:"size:128"`

	// 库存，-1 表示不限库存
	Stock int `json:"stock" gorm:"not null;default:0"`

	// 商品图片地址，可空（保留字段：取第一张图片，供购物车等旧逻辑使用）
	Image string `json:"image" gorm:"size:255"`

	// 商品媒体列表（图片/视频），JSON 数组：[{"hash":"...","type":"image"}]，可空
	Media string `json:"media" gorm:"type:text"`

	// 同级排序，数值越小越靠前；前端拖拽后由业务层整组重写
	Sort int `json:"sort" gorm:"not null;default:0"`

	// 上架状态：1 上架、0 下架（默认 1）
	Status int8 `json:"status" gorm:"not null;default:1"`

	// 创建/更新时间，由 gorm 自动维护
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Goods 表结构
func (Goods) TableName() string {
	return "goods"
}
