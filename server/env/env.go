package env

import (
	"github.com/1szt/go-pkg/env"
)

// Start 初始化配置文件并补全缺失的配置项（不覆盖已有值）。
// 配置文件默认路径为 data/.env，可用环境变量覆盖任意项。
func Start() {
	env.Path("data/.env")

	env.Add("LFDD_PORT", "9081", "端口")
	env.Add("LFDD_SERVER_IP", "", "服务器 IP")

	env.Add("LFDD_SERVER_HTTPS", "", "是否启用 HTTPS（true/false）")
	env.Add("LFDD_SERVER_HTTPS_AUTO", "", "是否自签名证书（true/false）")
	env.Add("LFDD_SERVER_CERT", "", "HTTPS 证书路径")
	env.Add("LFDD_SERVER_KEY", "", "HTTPS 私钥路径")
	env.Add("LFDD_SERVER_INSTALL", "false", "是否已完成初始化安装（true/false）")

	env.Add("LFDD_DB_TYPE", "", "数据库类型（sqlite/pgsql/mysql）")
	env.Add("LFDD_DB_PATH", "", "SQLite 数据库文件路径")
	env.Add("LFDD_DB_HOST", "", "数据库主机地址")
	env.Add("LFDD_DB_PORT", "", "数据库端口")
	env.Add("LFDD_DB_NAME", "", "数据库名")
	env.Add("LFDD_DB_USER", "", "数据库用户名")
	env.Add("LFDD_DB_PASSWORD", "", "数据库密码")
}
