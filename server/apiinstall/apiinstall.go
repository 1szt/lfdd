// Package apiinstall 安装初始化接口：检查/提交服务器初始化配置，挂载于 /api/install/*。
package apiinstall

import (
	"hxzl/db"
	"hxzl/server"
	"hxzl/web"
	"log"

	"github.com/1szt/go-pkg/env"
	"github.com/gin-gonic/gin"
)

// Start 注册安装初始化路由（须在 web.Start 之后调用）
// 将apiinstall模块的路由注册改为使用命名路由组，保持与已完成重构模块的一致性
func Start() {
	group := web.Engine.Group("/api/install")
	group.GET("/check", GetCheck)
	group.POST("/submit", PostSubmit)	
	log.Print("✅ [APINSTALL] 安装接口 加载完成！")
}

// -------- 请求结构 --------

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

// -------- 接口1: 检查是否已初始化 --------
// GetCheck 检查服务器是否已初始化
// data.need_install=true 表示需要初始化
func GetCheck(c *gin.Context) {
	install := env.Get("LFDD_SERVER_INSTALL")

	// 不存在 / 空 / false → 需要安装
	needInstall := install == "" || install == "false"

	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data":    gin.H{"need_install": needInstall},
	})
}

// -------- 接口2: 提交初始化信息 --------
// InstallSubmit 提交初始化配置（写入配置并热重载数据库连接与监听端口）
// sqlite 必填 server_ip/server_port/db_type/db_path，
// 其余数据库类型必填 server_ip/server_port/db_type/db_host/db_port/db_name/db_user/db_password
func PostSubmit(c *gin.Context) {
	// 防重复提交：已初始化完成则直接拒绝，避免覆盖线上配置
	if env.Get("LFDD_SERVER_INSTALL") == "true" {
		c.JSON(400, gin.H{
			"code":    -1,
			"message": "服务器已初始化，不能重复提交",
		})
		return
	}

	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    -1,
			"message": "参数错误",
		})
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
			c.JSON(500, gin.H{
				"code":    -1,
				"message": "写入配置失败: " + item.name,
			})
			return
		}
	}

	// 最后设置 LFDD_SERVER_INSTALL = true，表示初始化完成
	if err := env.Set("LFDD_SERVER_INSTALL", "true"); err != nil {
		log.Printf("❌ [API] 初始化写入 server_install 失败: %v", err)
		c.JSON(500, gin.H{
			"code":    -1,
			"message": "写入初始化标记失败",
		})
		return
	}

	// 初始化完成后重载数据库连接
	db.Reload()

	// 先返回响应，再异步热切换到新地址：
	// Restart 会 Shutdown 当前服务并等待在途请求结束，当前请求就是本请求，
	// 同步调用会阻塞到超时；放 goroutine 里让响应先送达前端。
	c.JSON(200, gin.H{
		"code":    0,
		"message": "初始化完成",
	})
	go server.Restart()
}
