package service

import (
	"errors"
	"fmt"
	"math/rand"
	"server/core/db"
	"server/model"
	"time"

	"gorm.io/gorm"
)

// ------------------------ CreateOrder 创建订单 ------------------------

type CreateOrderItemInput struct {
	ID  int64 `json:"id,string"` // 商品 ID
	Qty int   `json:"qty"`       // 购买数量
}
type CreateOrderInput struct {
	Items []CreateOrderItemInput `json:"items"` // 购物车明细
}
type CreateOrderOutput struct {
	ID      int64  `json:"id,string"`
	OrderNo string `json:"order_no"` // 订单号（打印小票用）
}

// CreateOrder 下单：只记录商品/数量快照，不记录价格（价格待结单时锁定），
// 校验上架/库存，并在事务中扣减库存、写订单与明细（明细 price=0，订单 total=0、status=0）
func CreateOrder(i CreateOrderInput) (o CreateOrderOutput, err error) {
	if len(i.Items) == 0 {
		return o, errors.New("购物车是空的")
	}

	// 1. 合并同商品数量并校验基本值
	qtyByID := map[int64]int{}
	ids := []int64{}
	for _, it := range i.Items {
		if it.ID <= 0 || it.Qty <= 0 {
			return o, errors.New("购物车数据异常")
		}
		if _, seen := qtyByID[it.ID]; !seen {
			ids = append(ids, it.ID)
		}
		qtyByID[it.ID] += it.Qty
		if qtyByID[it.ID] > matchMaxQty {
			return o, fmt.Errorf("单商品数量不能超过 %d", matchMaxQty)
		}
	}

	// 2. 取上架商品（只校验上架与库存，不算价格）
	var goods []model.Goods
	if err = db.DB.Where("id IN ? AND status = ?", ids, 1).Find(&goods).Error; err != nil {
		return o, err
	}
	byID := make(map[int64]model.Goods, len(goods))
	for _, g := range goods {
		byID[g.ID] = g
	}

	count := 0
	for _, id := range ids {
		g, ok := byID[id]
		if !ok {
			return o, fmt.Errorf("商品已下架或不存在，请刷新购物车")
		}
		qty := qtyByID[id]
		if g.Stock > 0 && qty > g.Stock {
			return o, fmt.Errorf("「%s」库存不足（剩 %d）", g.Name, g.Stock)
		}
		count += qty
	}

	// 3. 事务：扣库存 → 建订单（待结单）→ 写明细（价格留空待结单回填）
	o.OrderNo = time.Now().Format("20060102150405") + fmt.Sprintf("%06d", rand.Intn(1e6))
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			g, qty := byID[id], qtyByID[id]
			if g.Stock <= 0 {
				continue // -1 不限库存
			}
			res := tx.Model(&model.Goods{}).
				Where("id = ? AND stock >= ?", id, qty).
				Update("stock", gorm.Expr("stock - ?", qty))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("「%s」库存不足", g.Name)
			}
		}

		ord := model.Order{OrderNo: o.OrderNo, Count: count, Status: 0}
		if e := tx.Create(&ord).Error; e != nil {
			return e
		}
		items := make([]model.OrderItem, 0, len(ids))
		for _, id := range ids {
			g, qty := byID[id], qtyByID[id]
			items = append(items, model.OrderItem{
				OrderID: ord.ID, GoodsID: g.ID, Name: g.Name,
				Spec: g.Spec, Qty: qty, Image: g.Image, // Price 留 0，结单时回填
			})
		}
		if e := tx.Create(&items).Error; e != nil {
			return e
		}
		o.ID = ord.ID
		return nil
	})
	if err != nil {
		return o, err
	}
	return o, nil
}

// ------------------------ FinishOrder 结单（锁定价格） ------------------------

type FinishOrderInput struct {
	ID int64 `json:"id,string"` // 订单 ID
}
type FinishOrderOutput struct {
	Total int `json:"total"` // 锁定后的合计，分
	Count int `json:"count"` // 商品总件数
}

// FinishOrder 结单：按库中当前价锁定本单全部明细，合计回写订单并置为已结单。
// 只能结一次（status=0 才允许，事务内防并发重复结单）；之后商品再改价不影响本单。
func FinishOrder(i FinishOrderInput) (o FinishOrderOutput, err error) {
	// 1. 查订单（仅待结单可结）
	var ord model.Order
	if err = db.DB.First(&ord, i.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return o, errors.New("订单不存在")
		}
		return o, err
	}
	if ord.Status != 0 {
		return o, errors.New("该订单已结单")
	}

	// 2. 查明细与当前商品价
	var items []model.OrderItem
	if err = db.DB.Where("order_id = ?", ord.ID).Order("id ASC").Find(&items).Error; err != nil {
		return o, err
	}
	if len(items) == 0 {
		return o, errors.New("订单没有明细")
	}
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	var goods []model.Goods
	if err = db.DB.Where("id IN ?", ids).Find(&goods).Error; err != nil {
		return o, err
	}
	byID := make(map[int64]model.Goods, len(goods))
	for _, g := range goods {
		byID[g.ID] = g
	}

	total := 0
	for _, it := range items {
		if it.Edited == 1 { // 改过价的保留改后价
			total += it.Price * it.Qty
			continue
		}
		g, ok := byID[it.GoodsID]
		if !ok {
			return o, fmt.Errorf("「%s」已删除，无法结单", it.Name)
		}
		total += g.Price * it.Qty
	}

	// 3. 事务：回填未改价明细的价 → 合计与状态一次更新
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		for _, it := range items {
			if it.Edited == 1 {
				continue // 改价行的价格已锁定，不覆盖
			}
			g := byID[it.GoodsID]
			if e := tx.Model(&model.OrderItem{}).Where("id = ?", it.ID).
				Update("price", g.Price).Error; e != nil {
				return e
			}
		}
		res := tx.Model(&model.Order{}).Where("id = ? AND status = ?", ord.ID, 0).
			Updates(map[string]any{"total": total, "status": 1})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("该订单已结单")
		}
		return nil
	})
	if err != nil {
		return o, err
	}

	o.Total = total
	o.Count = ord.Count
	return o, nil
}

// ------------------------ ChangePrice 改价 ------------------------

type ChangePriceInput struct {
	ID      int64 `json:"id,string"`      // 订单 ID
	GoodsID int64 `json:"goods_id,string"` // 商品 ID（订单内明细）
	Price   int   `json:"price"`          // 改后单价，分
}
type ChangePriceOutput struct{}

// ChangePrice 改价：修改待结单订单里某个商品的单价（改后价立即生效并锁定，结单不覆盖）。
// 只允许 status=0；价格必须 >= 0。
func ChangePrice(i ChangePriceInput) (o ChangePriceOutput, err error) {
	if i.Price < 0 {
		return o, errors.New("价格不能为负数")
	}
	var ord model.Order
	if err = db.DB.First(&ord, i.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return o, errors.New("订单不存在")
		}
		return o, err
	}
	if ord.Status != 0 {
		return o, errors.New("已结单的订单不能改价")
	}

	// 原价 = 商品表当前价（即「没改的价格」），用于改价前后对比展示
	orig := 0
	var g model.Goods
	if e := db.DB.First(&g, i.GoodsID).Error; e == nil {
		orig = g.Price
	}

	res := db.DB.Model(&model.OrderItem{}).
		Where("order_id = ? AND goods_id = ?", ord.ID, i.GoodsID).
		Updates(map[string]any{"price": i.Price, "edited": 1, "orig_price": orig})
	if res.Error != nil {
		return o, res.Error
	}
	if res.RowsAffected == 0 {
		return o, errors.New("订单里没有该商品")
	}
	return o, nil
}

// ------------------------ VoidOrder 废单 ------------------------

type VoidOrderInput struct {
	ID int64 `json:"id,string"` // 订单 ID
}
type VoidOrderOutput struct{}

// VoidOrder 废单：作废未结单订单——回补库存后删除订单与明细。已结单的不可废单。
func VoidOrder(i VoidOrderInput) (o VoidOrderOutput, err error) {
	var ord model.Order
	if err = db.DB.First(&ord, i.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return o, errors.New("订单不存在")
		}
		return o, err
	}
	if ord.Status != 0 {
		return o, errors.New("已结单的订单不能废单")
	}

	var items []model.OrderItem
	if err = db.DB.Where("order_id = ?", ord.ID).Find(&items).Error; err != nil {
		return o, err
	}
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	qtyByGoods := map[int64]int{}
	for _, it := range items {
		qtyByGoods[it.GoodsID] += it.Qty
	}

	var gs []model.Goods
	if len(ids) > 0 {
		if err = db.DB.Where("id IN ?", ids).Find(&gs).Error; err != nil {
			return o, err
		}
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// 回补库存（商品已删则无法回补；stock=-1 不限库存）
		for _, g := range gs {
			if g.Stock < 0 {
				continue
			}
			if e := tx.Model(&model.Goods{}).Where("id = ?", g.ID).
				Update("stock", gorm.Expr("stock + ?", qtyByGoods[g.ID])).Error; e != nil {
				return e
			}
		}
		if e := tx.Where("order_id = ?", ord.ID).Delete(&model.OrderItem{}).Error; e != nil {
			return e
		}
		return tx.Delete(&model.Order{}, ord.ID).Error
	})
	if err != nil {
		return o, err
	}
	return o, nil
}

// ------------------------ DeleteOrder 删除订单 ------------------------

type DeleteOrderInput struct {
	ID int64 `json:"id,string"` // 订单 ID
}
type DeleteOrderOutput struct{}

// DeleteOrder 删除订单：只允许删除已结单（status=1）的订单，连明细一并删除
func DeleteOrder(i DeleteOrderInput) (o DeleteOrderOutput, err error) {
	var ord model.Order
	if err = db.DB.First(&ord, i.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return o, errors.New("订单不存在")
		}
		return o, err
	}
	if ord.Status != 1 {
		return o, errors.New("未结单的订单不能删除")
	}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if e := tx.Where("order_id = ?", ord.ID).Delete(&model.OrderItem{}).Error; e != nil {
			return e
		}
		return tx.Delete(&model.Order{}, ord.ID).Error
	})
	if err != nil {
		return o, err
	}
	return o, nil
}

// ------------------------ GetOrders 订单分页 ------------------------

type GetOrdersInput struct {
	Page     int `json:"page"`      // 页码，从 1 开始
	PageSize int `json:"page_size"` // 每页条数
	Status   int `json:"status"`    // -1 全部（默认）、0 未结单、1 已结单
}
type OrderItemView struct {
	GoodsID    int64  `json:"goods_id,string"`
	Name       string `json:"name"`
	Price      int    `json:"price"`      // 单价，分（未改价的待结单单为商品当前价；改价/结单后为锁定价）
	Edited     int8   `json:"edited"`     // 是否改过价（改价标记）
	OrigPrice  int    `json:"orig_price"` // 改价前原价，分（仅改过价的行有值，展示对比用）
	Spec       string `json:"spec"`
	Qty        int    `json:"qty"`
	Image      string `json:"image"`
}
type OrderView struct {
	ID        int64           `json:"id,string"`
	OrderNo   string          `json:"order_no"`
	Total     int             `json:"total"`   // 合计，分（待结单时为 0）
	Count     int             `json:"count"`   // 商品总件数
	Status    int8            `json:"status"`  // 0 待结单、1 已结单
	CreatedAt string          `json:"created_at"`
	Items     []OrderItemView `json:"items"`
}
type GetOrdersOutput struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	List     []OrderView `json:"list"`
}

// GetOrders 订单分页：最新在前，附带每单明细（供列表展示与小票打印）
func GetOrders(i GetOrdersInput) (o GetOrdersOutput, err error) {
	// 分页参数：默认 1 页 20 条，单页上限 100
	if i.Page < 1 {
		i.Page = 1
	}
	if i.PageSize < 1 {
		i.PageSize = 20
	}
	if i.PageSize > 100 {
		i.PageSize = 100
	}

	q := db.DB.Model(&model.Order{})
	if i.Status == 0 || i.Status == 1 {
		q = q.Where("status = ?", i.Status)
	}
	if err = q.Count(&o.Total).Error; err != nil {
		return o, err
	}
	var orders []model.Order
	if err = q.Order("id DESC").Limit(i.PageSize).Offset((i.Page - 1) * i.PageSize).
		Find(&orders).Error; err != nil {
		return o, err
	}
	o.Page = i.Page
	o.PageSize = i.PageSize
	o.List = make([]OrderView, 0, len(orders))
	if len(orders) == 0 {
		return o, nil
	}

	// 批量取明细，按订单分组
	orderIDs := make([]int64, 0, len(orders))
	for _, ord := range orders {
		orderIDs = append(orderIDs, ord.ID)
	}
	var items []model.OrderItem
	if err = db.DB.Where("order_id IN ?", orderIDs).Order("id ASC").Find(&items).Error; err != nil {
		return o, err
	}
	itemsByOrder := map[int64][]OrderItemView{}
	for _, it := range items {
		itemsByOrder[it.OrderID] = append(itemsByOrder[it.OrderID], OrderItemView{
			GoodsID: it.GoodsID, Name: it.Name, Price: it.Price, Edited: it.Edited,
			OrigPrice: it.OrigPrice,
			Spec:      it.Spec, Qty: it.Qty, Image: it.Image,
		})
	}

	// 待结单订单：未改价的明细按商品表当前价实时展示、合计实时算（结单时才回写锁定）；
	// 改过价的明细保留改后价；未改价且商品已删除的行价格留 0（结单时会拒绝）
	pendingIDs := []int64{}
	for _, ord := range orders {
		if ord.Status == 0 {
			pendingIDs = append(pendingIDs, ord.ID)
		}
	}
	pendingTotal := map[int64]int{} // 订单 ID → 实时合计
	if len(pendingIDs) > 0 {
		pendingGoods := []int64{}
		for _, id := range pendingIDs {
			for _, v := range itemsByOrder[id] {
				pendingGoods = append(pendingGoods, v.GoodsID)
			}
		}
		var gs []model.Goods
		if err = db.DB.Where("id IN ?", pendingGoods).Find(&gs).Error; err != nil {
			return o, err
		}
		priceByGoods := make(map[int64]int, len(gs))
		for _, g := range gs {
			priceByGoods[g.ID] = g.Price
		}
		for _, id := range pendingIDs {
			total := 0
			for n, v := range itemsByOrder[id] {
				if v.Edited == 1 {
					total += v.Price * v.Qty // 改过价：用改后价
					continue
				}
				if p, ok := priceByGoods[v.GoodsID]; ok {
					itemsByOrder[id][n].Price = p
					total += p * v.Qty
				}
			}
			pendingTotal[id] = total
		}
	}

	for _, ord := range orders {
		list := itemsByOrder[ord.ID]
		if list == nil {
			list = []OrderItemView{} // 保证输出 [] 而非 null
		}
		total := ord.Total
		if ord.Status == 0 {
			total = pendingTotal[ord.ID] // 待结单用实时合计
		}
		o.List = append(o.List, OrderView{
			ID:        ord.ID,
			OrderNo:   ord.OrderNo,
			Total:     total,
			Count:     ord.Count,
			Status:    ord.Status,
			CreatedAt: ord.CreatedAt.Format("2006-01-02 15:04:05"),
			Items:     list,
		})
	}
	return o, nil
}
