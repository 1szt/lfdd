// Package web web 框架初始化模块：持有全局 Engine 并注册全局中间件，必须在模块注册路由前调用。
package web

import (
	"hxzl/web/middleware"
	"strconv"
	"time"

	"github.com/1szt/go-pkg/env"
	"github.com/gin-gonic/gin"
)

// Engine 全局路由引擎，各业务模块通过它注册自己的路由
var Engine = gin.New()

// Start 注册全局中间件（Logger + Recovery + 初始化检查）
// 必须在所有业务模块注册路由之前调用（gin 中间件对之后注册的路由才生效）
func Start() {
	// 非调试模式切换 gin 为 release，屏蔽路由注册等调试日志刷屏
	if env.Get("LFDD_DEBUG") != "true" {
		gin.SetMode(gin.ReleaseMode)
	}
	Engine.Use(gin.LoggerWithFormatter(accessLogFormat))
	Engine.Use(gin.Recovery())
	Engine.Use(middleware.InitGuard())
}

// accessLogFormat 自定义访问日志格式（去掉 gin 默认的 [GIN] 前缀）
func accessLogFormat(p gin.LogFormatterParams) string {
	return p.TimeStamp.Format("2006/01/02 - 15:04:05") +
		" | " + strconv.Itoa(p.StatusCode) +
		" | " + fmtLatency(p.Latency) +
		" | " + p.ClientIP +
		" | " + p.Method +
		" " + p.Path +
		errorMsg(p.ErrorMessage)
}

// fmtLatency 去掉尾随零的耗时（如 1.5ms、0s）
func fmtLatency(d time.Duration) string {
	return d.Round(time.Microsecond).String()
}

// errorMsg 请求出错时附加错误信息
func errorMsg(msg string) string {
	if msg == "" {
		return ""
	}
	return " | error: " + msg
}
