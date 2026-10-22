package api

import (
	"hxzl/config"
	"hxzl/db"
	"hxzl/server"

	"github.com/gin-gonic/gin"
)

// -------- 请求/响应结构 --------

type InstallRequest struct {
	ServerIP   string `json:"server_ip"`
	ServerPort string `json:"server_port"`
	DBType     string `json:"db_type"`
	DBPath     string `json:"db_path"`
	DBHost     string `json:"db_host"`
	DBPort     string `json:"db_port"`
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
}

type InstallResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// -------- 接口1: 检查是否已初始化 --------
// GET /api/install/check
func InstallCheck(c *gin.Context) {
	install := config.Get("server_install").Value

	// 不存在 / 空 / false → 需要安装
	needInstall := install == "" || install == "false"

	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"need_install": needInstall,
		},
	})
}

// -------- 接口2: 提交初始化信息 --------
// POST /api/install/submit
func InstallSubmit(c *gin.Context) {
	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, InstallResponse{Code: -1, Message: "参数错误"})
		return
	}

	user := "admin" // 默认初始化用户

	// 写入所有配置（Set 会自动判断：不存在则创建，存在则修改）
	config.Set("server_ip", req.ServerIP, user)
	config.Set("server_port", req.ServerPort, user)
	config.Set("db_type", req.DBType, user)
	config.Set("db_path", req.DBPath, user)
	config.Set("db_host", req.DBHost, user)
	config.Set("db_port", req.DBPort, user)
	config.Set("db_name", req.DBName, user)
	config.Set("db_user", req.DBUser, user)
	config.Set("db_password", req.DBPassword, user)

	// 最后设置 server_install = true，表示初始化完成
	config.Set("server_install", "true", user)

	// 初始化完成后重载数据库连接
	db.Reload()

	// 初始化完成后热切换到新地址
	server.Restart()

	c.JSON(200, InstallResponse{Code: 0, Message: "初始化完成"})
}
