package motd

import "log"

// banner 启动横幅（原生字符串，所见即所得；中文按 2 列宽对齐）
const banner = `========================================
           lfdd 管理系统启动！
========================================
正在启动各个模块...`

// Start 输出启动横幅。走标准库 log 以便被日志写盘收集，临时关闭时间标志打印后恢复。
// 需在 logx.Start() 之后调用，否则横幅不会进日志文件。
func Start() {
	log.SetFlags(0)
	log.Print(banner)
	log.SetFlags(log.LstdFlags)
}
