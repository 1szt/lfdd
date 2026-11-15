package handler

import (
	"server/service"

	"github.com/gin-gonic/gin"
)

type ApiMenuHandler struct {
}

// Create 新建菜单
func (h *ApiMenuHandler) Create(c *gin.Context) {
	var req service.CreateMenuInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.CreateMenu(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, out)
}

// Update 修改菜单名称
func (h *ApiMenuHandler) Update(c *gin.Context) {
	var req service.UpdateMenuInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.UpdateMenu(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Delete 删除菜单（菜单下有商品时会被业务层拒绝）
func (h *ApiMenuHandler) Delete(c *gin.Context) {
	var req service.DeleteMenuInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.DeleteMenu(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Sort 菜单顺序整组重写（前端拖拽后调用）
// body: {"ids":["1","2","3"]} 按数组下标写 sort（0,1,2...）
func (h *ApiMenuHandler) Sort(c *gin.Context) {
	var req service.SortMenusInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.SortMenus(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}
