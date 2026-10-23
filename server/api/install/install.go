// Package install 安装初始化接口：检查/提交服务器初始化配置。
package install

import (
	"hxzl/config"
	"hxzl/db"
	"hxzl/server"
	"log"

	"github.com/gin-gonic/gin"
)

// Register 注册安装初始化相关路由（由 api 模块传入路由组）
func Register(rg *gin.RouterGroup) {
	rg.GET("/install/check", InstallCheck)
	rg.POST("/install/submit", InstallSubmit)
}

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
	// 防重复提交：已初始化完成则直接拒绝，避免覆盖线上配置
	if config.Get("server_install").Value == "true" {
		c.JSON(400, InstallResponse{Code: -1, Message: "服务器已初始化，不能重复提交"})
		return
	}

	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, InstallResponse{Code: -1, Message: "参数错误"})
		return
	}

	user := "admin" // 默认初始化用户

	// 写入所有配置（Set 会自动判断：不存在则创建，存在则修改）
	// 注意顺序：最后才写 server_install，中途失败则视为未完成初始化
	configs := []struct {
		name  string
		value string
	}{
		{"server_ip", req.ServerIP},
		{"server_port", req.ServerPort},
		{"db_type", req.DBType},
		{"db_path", req.DBPath},
		{"db_host", req.DBHost},
		{"db_port", req.DBPort},
		{"db_name", req.DBName},
		{"db_user", req.DBUser},
		{"db_password", req.DBPassword},
	}
	for _, item := range configs {
		if err := config.Set(item.name, item.value, user); err != nil {
			log.Printf("❌ [API] 初始化写入配置 %s 失败: %v", item.name, err)
			c.JSON(500, InstallResponse{Code: -1, Message: "写入配置失败: " + item.name})
			return
		}
	}

	// 最后设置 server_install = true，表示初始化完成
	if err := config.Set("server_install", "true", user); err != nil {
		log.Printf("❌ [API] 初始化写入 server_install 失败: %v", err)
		c.JSON(500, InstallResponse{Code: -1, Message: "写入初始化标记失败"})
		return
	}

	// 初始化完成后重载数据库连接
	db.Reload()

	// 初始化完成后热切换到新地址
	server.Restart()

	c.JSON(200, InstallResponse{Code: 0, Message: "初始化完成"})
}
