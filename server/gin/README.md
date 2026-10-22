# Gin 模块

Gin 模块负责初始化 Gin 框架、中间件、路由和 web 服务。

# 使用方法

gin.Start() 启动 web 服务（从配置读取 server_ip + server_port）如果配置不存在默认监听9081端口
gin.Restart() 强制重连（停掉旧服务 → 重新读取配置 → 启动）
gin.Reload() 智能重连（配置无变化且运行中则跳过，否则重启）
gin.Status() 获取服务状态

# Status 返回结构体

IP - 服务监听地址
Port - 服务端口
Status - 服务状态（running/stopped）


