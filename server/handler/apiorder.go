package handler

import (
	"server/service"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ApiOrderHandler struct {
}

// guestIDOf 取上下文游客身份（由 IdentityGuard 校验游客认证头后写入）：空串表示非游客
func guestIDOf(c *gin.Context) string {
	v, ok := c.Get("guestID")
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
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
	req.GuestID = guestIDOf(c) // 游客订单归属取认证头，不接受 body 传入

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
	if guestIDOf(c) != "" {
		c.JSON(403, gin.H{"message": "游客无权结单"})
		return
	}
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

// ChangePrice 改价：修改待结单订单某商品单价（打改价标记，结单不覆盖）
// body: {"id":"1","goods_id":"101","price":1500} → 200
func (h *ApiOrderHandler) ChangePrice(c *gin.Context) {
	if guestIDOf(c) != "" {
		c.JSON(403, gin.H{"message": "游客无权改价"})
		return
	}
	var req service.ChangePriceInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.ChangePrice(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{})
}

// Void 废单：作废未结单订单（回补库存并删除）
// body: {"id":"1"} → 200
func (h *ApiOrderHandler) Void(c *gin.Context) {
	if guestIDOf(c) != "" {
		c.JSON(403, gin.H{"message": "游客无权废单"})
		return
	}
	var req service.VoidOrderInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.VoidOrder(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{})
}

// Delete 删除订单（仅已结单）
// body: {"id":"1"} → 200
func (h *ApiOrderHandler) Delete(c *gin.Context) {
	if guestIDOf(c) != "" {
		c.JSON(403, gin.H{"message": "游客无权删除订单"})
		return
	}
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
		GuestID:  guestIDOf(c), // 游客只看自己的未结单（状态过滤由服务层强制）
	})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}

// Search 订单搜索（仅管理员）：query 同分页，另支持
// keyword（单号模糊匹配）、date（YYYY-MM-DD，按本地时区当天）
func (h *ApiOrderHandler) Search(c *gin.Context) {
	var req struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Status   int    `form:"status"`
		Keyword  string `form:"keyword"`
		Date     string `form:"date"`
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
	req.Keyword = strings.TrimSpace(req.Keyword)
	req.Date = strings.TrimSpace(req.Date)
	if req.Date != "" {
		if _, err := time.Parse("2006-01-02", req.Date); err != nil {
			c.JSON(400, gin.H{
				"message": "日期格式应为 YYYY-MM-DD",
			})
			return
		}
	}

	out, err := service.GetOrders(service.GetOrdersInput{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
		Keyword:  req.Keyword,
		Date:     req.Date,
	})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}
