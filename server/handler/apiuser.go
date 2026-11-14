package handler

import (
	"server/service"

	"github.com/gin-gonic/gin"
)

type ApiUserHandler struct {
}

// Me 当前登录用户信息（角色用来控制管理页签是否展示）
func (h *ApiUserHandler) Me(c *gin.Context) {
	out, err := service.GetMe(service.GetMeInput{UserID: c.GetInt64("userID")})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}

// List 用户分页列表（query: keyword、page、page_size）
func (h *ApiUserHandler) List(c *gin.Context) {
	var req struct {
		Keyword  string `form:"keyword"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.GetUsers(service.GetUsersInput{
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}

// Update 修改用户角色/状态（role：1 普通 2 管理员；status：1 正常 0 禁用）
func (h *ApiUserHandler) Update(c *gin.Context) {
	var req service.UpdateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.UpdateUser(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Delete 删除用户（操作人从上下文取，禁止删自己）
func (h *ApiUserHandler) Delete(c *gin.Context) {
	var req struct {
		ID int64 `json:"id,string"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.DeleteUser(service.DeleteUserInput{
		ID:         req.ID,
		OperatorID: c.GetInt64("userID"),
	}); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// ResetPassword 管理员重置用户密码（body: {"id":"...","password":"..."}）
func (h *ApiUserHandler) ResetPassword(c *gin.Context) {
	var req service.ResetUserPasswordInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.ResetUserPassword(service.ResetUserPasswordInput{
		ID:        req.ID,
		Password:  req.Password,
		KeepToken: c.GetString("token"),
	}); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}
