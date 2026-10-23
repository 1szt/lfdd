// Package logx 日志模块：标准库 log 与 gin 访问日志统一输出（控制台 + 可选写盘）。
package logx

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/1szt/go-pkg/env"
	"github.com/gin-gonic/gin"
)

// Start 初始化日志输出。必须在 env.Start() 之后、其余模块之前调用。
// 控制台始终输出；配置了 LFDD_LOG_PATH 则同时追加写入该文件。
func Start() {
	var w io.Writer = os.Stdout

	if path := env.Get("LFDD_LOG_PATH"); path != "" {
		f, err := openLogFile(path)
		if err != nil {
			log.Printf("⚠️ [Log] 打开日志文件失败，仅输出控制台: %v", err)
		} else {
			w = io.MultiWriter(os.Stdout, f)
		}
	}

	log.SetOutput(w)      // 标准库 log（业务日志）
	gin.DefaultWriter = w // gin 访问日志
	gin.DefaultErrorWriter = w

	log.Println("✅ [Log] 日志模块 初始化完成")
}

// openLogFile 创建目录并以追加模式打开日志文件；启动时超限则轮转一次。
func openLogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	rotate(path)
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

// maxLogSize 单个日志文件上限，启动时超过即轮转（仅保留一份 .old）
const maxLogSize = 50 << 20 // 50MB

// rotate 日志文件超过 maxLogSize 时重命名为 .old（旧 .old 被覆盖），仅在启动时执行
func rotate(path string) {
	info, err := os.Stat(path)
	if err != nil || info.Size() <= maxLogSize {
		return
	}
	os.Rename(path, path+".old")
}
