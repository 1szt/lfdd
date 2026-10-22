# Server 模块

Server 模块负责 web 框架：持有全局 Router 实例，并管理 web 服务的启动/重启/热重载。

# 使用方法

server.Start() 启动 web 服务（从配置读取 server_ip + server_port）如果配置不存在默认监听9081端口
server.Restart() 强制重连（停掉旧服务 → 重新读取配置 → 启动）
server.Reload() 智能重连（配置无变化且运行中则跳过，否则重启）
server.Status() 获取服务状态
server.Router 全局路由实例，各业务模块通过它注册自己的路由

# Status 返回结构体

IP - 服务监听地址
Port - 服务端口
Status - 服务状态（running/stopped）
