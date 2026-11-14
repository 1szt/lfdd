package handler

import (
	"strconv"

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

// Create 管理员创建用户（body: username/email/password/role）
func (h *ApiUserHandler) Create(c *gin.Context) {
	var req service.CreateUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.CreateUser(req)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, out)
}

// UpdateInfo 修改用户名/密码（body: id、username、password 留空不改）
func (h *ApiUserHandler) UpdateInfo(c *gin.Context) {
	var req service.UpdateUserInfoInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	req.KeepToken = c.GetString("token")
	if _, err := service.UpdateUserInfo(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}

// Sessions 某用户的登录设备/会话列表（query: user_id）
func (h *ApiUserHandler) Sessions(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil || userID == 0 {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	out, err := service.GetUserSessions(service.GetUserSessionsInput{
		UserID:       userID,
		CurrentToken: c.GetString("token"),
	})
	if err != nil {
		c.JSON(500, gin.H{
			"message": "服务器内部错误",
		})
		return
	}

	c.JSON(200, out)
}

// Kick 踢指定设备下线（body: user_id、token）
func (h *ApiUserHandler) Kick(c *gin.Context) {
	var req service.KickSessionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": "请求参数错误",
		})
		return
	}

	if _, err := service.KickSession(req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "ok",
	})
}
