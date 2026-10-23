// Package install 安装初始化接口：检查/提交服务器初始化配置。
package install

import (
	"hxzl/db"
	"hxzl/server"
	"log"

	"github.com/1szt/go-pkg/env"
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

type CheckData struct {
	NeedInstall bool `json:"need_install"`
}

type CheckResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    CheckData `json:"data"`
}

// -------- 接口1: 检查是否已初始化 --------
// InstallCheck 检查服务器是否已初始化
// @Summary 检查服务器是否已初始化
// @Tags 安装
// @Produce json
// @Success 200 {object} CheckResponse "data.need_install=true 表示需要初始化"
// @Router /install/check [get]
func InstallCheck(c *gin.Context) {
	install := env.Get("LFDD_SERVER_INSTALL")

	// 不存在 / 空 / false → 需要安装
	needInstall := install == "" || install == "false"

	c.JSON(200, CheckResponse{
		Code:    0,
		Message: "ok",
		Data:    CheckData{NeedInstall: needInstall},
	})
}

// -------- 接口2: 提交初始化信息 --------
// InstallSubmit 提交初始化配置
// @Summary 提交初始化配置（写入配置并热重载数据库连接与监听端口）
// @Tags 安装
// @Accept json
// @Param data body InstallRequest true "初始化参数（sqlite 必填 server_ip/server_port/db_type/db_path，其余数据库类型必填 server_ip/server_port/db_type/db_host/db_port/db_name/db_user/db_password）"
// @Success 200 {object} InstallResponse "初始化完成，随后服务热切换"
// @Failure 400 {object} InstallResponse "参数错误或已初始化"
// @Failure 500 {object} InstallResponse "写入配置失败"
// @Router /install/submit [post]
func InstallSubmit(c *gin.Context) {
	// 防重复提交：已初始化完成则直接拒绝，避免覆盖线上配置
	if env.Get("LFDD_SERVER_INSTALL") == "true" {
		c.JSON(400, InstallResponse{Code: -1, Message: "服务器已初始化，不能重复提交"})
		return
	}

	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, InstallResponse{Code: -1, Message: "参数错误"})
		return
	}

	// 写入所有配置（Set 会自动判断：不存在则创建，存在则修改）
	// 注意顺序：最后才写 LFDD_SERVER_INSTALL，中途失败则视为未完成初始化
	configs := []struct {
		name  string
		value string
	}{
		{"LFDD_SERVER_IP", req.ServerIP},
		{"LFDD_PORT", req.ServerPort},
		{"LFDD_DB_TYPE", req.DBType},
		{"LFDD_DB_PATH", req.DBPath},
		{"LFDD_DB_HOST", req.DBHost},
		{"LFDD_DB_PORT", req.DBPort},
		{"LFDD_DB_NAME", req.DBName},
		{"LFDD_DB_USER", req.DBUser},
		{"LFDD_DB_PASSWORD", req.DBPassword},
	}
	for _, item := range configs {
		if err := env.Set(item.name, item.value); err != nil {
			log.Printf("❌ [API] 初始化写入配置 %s 失败: %v", item.name, err)
			c.JSON(500, InstallResponse{Code: -1, Message: "写入配置失败: " + item.name})
			return
		}
	}

	// 最后设置 LFDD_SERVER_INSTALL = true，表示初始化完成
	if err := env.Set("LFDD_SERVER_INSTALL", "true"); err != nil {
		log.Printf("❌ [API] 初始化写入 server_install 失败: %v", err)
		c.JSON(500, InstallResponse{Code: -1, Message: "写入初始化标记失败"})
		return
	}

	// 初始化完成后重载数据库连接
	db.Reload()

	// 先返回响应，再异步热切换到新地址：
	// Restart 会 Shutdown 当前服务并等待在途请求结束，当前请求就是本请求，
	// 同步调用会阻塞到超时；放 goroutine 里让响应先送达前端。
	c.JSON(200, InstallResponse{Code: 0, Message: "初始化完成"})
	go server.Restart()
}
