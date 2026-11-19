package handler

import (
	"server/service"

	"github.com/gin-gonic/gin"
)

type ApiGoodsHandler struct {
}

// isAdminRequest 当前请求是否来自管理员（游客 / 未登录返回 false）
func isAdminRequest(c *gin.Context) bool {
	uid := c.GetInt64("userID")
	if uid == 0 {
		return false
	}
	me, err := service.GetMe(service.GetMeInput{UserID: uid})
	return err == nil && me.Role == 2
}

// Sidebar 商品侧边菜单（菜单/分类列表，按 sort 升序）
func (h *ApiGoodsHandler) Sidebar(c *gin.Context) {
	out, err := service.GetGoodsSidebar(service.GetGoodsSidebarInput{})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}

// Page 商品页面（指定菜单下商品的分页列表）
// query: menu_id（可选，0/不传 = 全部）、page（默认 1）、page_size（默认 20，上限 100）、
// all（1 = 含下架商品，管理端用；不传/0 = 只看上架）
func (h *ApiGoodsHandler) Page(c *gin.Context) {
	var req struct {
		MenuID   int64 `form:"menu_id"`
		Page     int   `form:"page"`
		PageSize int   `form:"page_size"`
		All      int   `form:"all"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.GetGoodsPage(service.GetGoodsPageInput{
		MenuID:   req.MenuID,
		Page:     req.Page,
		PageSize: req.PageSize,
		All:      req.All == 1,
		WithCost: isAdminRequest(c), // 进价仅管理员可见
	})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}

// All 全部上架商品（前端一次性拉取后本地模糊搜索）
func (h *ApiGoodsHandler) All(c *gin.Context) {
	out, err := service.GetGoodsAll(service.GetGoodsAllInput{WithCost: isAdminRequest(c)})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}

// Create 新建商品（挂到指定菜单下）
// body: {"menu_id":"101","name":"杨枝甘露","price":1850,"spec":"","stock":-1,"image":"","status":1,"sort":0}
// price 单位：分（1850 = 18.50 元）
func (h *ApiGoodsHandler) Create(c *gin.Context) {
	var req service.CreateGoodsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.CreateGoods(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, out)
}

// Update 修改商品（价格、顺序、库存、上下架、所属菜单等全量更新，前端回传完整对象）
func (h *ApiGoodsHandler) Update(c *gin.Context) {
	var req service.UpdateGoodsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.UpdateGoods(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Delete 删除商品
func (h *ApiGoodsHandler) Delete(c *gin.Context) {
	var req service.DeleteGoodsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.DeleteGoods(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Sort 商品顺序整组重写（前端拖拽后调用）
// body: {"ids":["1","2","3"]} 按数组下标写 sort（0,1,2...）
func (h *ApiGoodsHandler) Sort(c *gin.Context) {
	var req service.SortGoodsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.SortGoods(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Upload 商品媒体上传（multipart 字段名 file；图片 ≤5MB jpg/png/gif/webp，视频 ≤50MB mp4/mov/webm）
// 文件按内容 SHA-256 哈希命名存储（无后缀、同内容去重），返回 {"hash":"{sha256}"}
// 数据库 goods.media 只存哈希，访问走 GET /api/upload/{hash}（类型按内容嗅探）
func (h *ApiGoodsHandler) Upload(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{
			"message": "请选择图片或视频文件（multipart 字段名 file）",
		})
		return
	}

	out, err := service.UploadImage(service.UploadImageInput{File: fh})
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, out)
}

// Match AI 商品匹配：前端传自然语言点单文字，后端取库中商品拼提示词调大模型，
// 返回命中的商品完整信息 + 数量，前端据此加购物车
// body: {"text":"两杯柠檬水"} → {"list":[{...商品字段,"qty":2}]}
func (h *ApiGoodsHandler) Match(c *gin.Context) {
	var req service.MatchGoodsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.MatchGoods(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, out)
}

// Image 图片直出：GET /api/upload/{hash}，校验哈希文件名后按内容类型返回文件
func (h *ApiGoodsHandler) Image(c *gin.Context) {
	out, err := service.GetUploadImage(service.GetUploadImageInput{Name: c.Param("name")})
	if err != nil {
		c.JSON(404, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.File(out.Path)
}
