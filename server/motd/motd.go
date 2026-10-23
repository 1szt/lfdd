package motd

import "fmt"

// banner 启动横幅（原生字符串，所见即所得；中文按 2 列宽对齐）
const banner = `
========================================
           lfdd 管理系统启动！
========================================
正在启动各个模块...`

func Start() {
	fmt.Println(banner)
}
