package model

import "time"

// Order 订单表：购物车结算生成一条
type Order struct {
	// 自增主键
	ID int64 `json:"id,string" gorm:"primaryKey;autoIncrement"`

	// 订单号（时间戳+随机数，唯一索引）
	OrderNo string `json:"order_no" gorm:"size:32;uniqueIndex"`

	// 合计金额，单位：分（1850 = 18.50 元）
	Total int `json:"total" gorm:"not null;default:0"`

	// 商品总件数
	Count int `json:"count" gorm:"not null;default:0"`

	// 状态：0 = 待结单（价格未锁定）、1 = 已结单（价格已锁定）
	Status int8 `json:"status" gorm:"not null;default:0"`

	// 创建/更新时间，由 gorm 自动维护
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Order 表结构
func (Order) TableName() string {
	return "orders"
}

// OrderItem 订单明细：下单时记录名称/规格/图片快照，价格在「结单」时才按当时价锁定回填
type OrderItem struct {
	// 自增主键
	ID int64 `json:"id,string" gorm:"primaryKey;autoIncrement"`

	// 所属订单（order.id），普通索引
	OrderID int64 `json:"order_id,string" gorm:"not null;index"`

	// 商品 ID（goods.id），结单时按它取当前价
	GoodsID int64 `json:"goods_id,string" gorm:"not null"`

	// 商品名称快照
	Name string `json:"name" gorm:"size:64;not null"`

	// 单价，单位：分；未改价时下单为 0（展示用商品当前价），改价/结单时回填并锁定
	Price int `json:"price" gorm:"not null;default:0"`

	// 是否改过价（0 未改、1 已改）：改价标记，结单时保留改后价
	Edited int8 `json:"edited" gorm:"not null;default:0"`

	// 改价前的原价，单位：分；未改价为 0（仅展示对比用，结单不改写）
	OrigPrice int `json:"orig_price" gorm:"not null;default:0"`

	// 规格快照
	Spec string `json:"spec" gorm:"size:128"`

	// 购买数量
	Qty int `json:"qty" gorm:"not null;default:1"`

	// 图片哈希快照
	Image string `json:"image" gorm:"size:255"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// OrderItem 表结构
func (OrderItem) TableName() string {
	return "order_items"
}
