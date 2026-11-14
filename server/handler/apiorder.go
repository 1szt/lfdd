package handler

import (
	"server/service"

	"github.com/gin-gonic/gin"
)

type ApiOrderHandler struct {
}

// Create 创建订单（购物车下单）：不记录价格、扣库存，明细价格待结单锁定
// body: {"items":[{"id":"101","qty":2}]} → {"id":"1","order_no":"20261114120000123456"}
func (h *ApiOrderHandler) Create(c *gin.Context) {
	var req service.CreateOrderInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.CreateOrder(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, out)
}

// Finish 结单：按当前价锁定本单明细并合计，只能结一次
// body: {"id":"1"} → {"total":3700,"count":2}
func (h *ApiOrderHandler) Finish(c *gin.Context) {
	var req service.FinishOrderInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.FinishOrder(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, out)
}

// Update 编辑未结单订单（整体替换单品数量/增删行，库存按差额调整）
// body: {"id":"1","items":[{"id":"101","qty":2}]} → {"count":2}
func (h *ApiOrderHandler) Update(c *gin.Context) {
	var req service.UpdateOrderInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.UpdateOrder(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, out)
}

// Delete 删除订单（仅已结单）
// body: {"id":"1"} → 200
func (h *ApiOrderHandler) Delete(c *gin.Context) {
	var req service.DeleteOrderInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.DeleteOrder(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{})
}

// Page 订单分页（最新在前，含明细，供列表展示与小票打印）
// query: page（默认 1）、page_size（默认 20，上限 100）、status（-1 全部，0 未结单，1 已结单）
func (h *ApiOrderHandler) Page(c *gin.Context) {
	var req struct {
		Page     int `form:"page"`
		PageSize int `form:"page_size"`
		Status   int `form:"status"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}
	if _, ok := c.GetQuery("status"); !ok {
		req.Status = -1 // 未传 = 全部
	}

	out, err := service.GetOrders(service.GetOrdersInput{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
	})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}
