客户端连接分配一个随机的id，用于标识客户端

函数
ws.Start()
加载WebSocket模块

发送消息给指定id的客户端
ws.Send(id string, msg []byte)

# 注意

- ws.Start() 注册路由 `/ws`，从 query 取 `token`、`c.ClientIP()` 后升级为 WebSocket（AcceptOptions 设 InsecureSkipVerify: true 允许跨域）
- 客户端 id 分配与 Send 实现位于 [hub.go1](./hub.go1)，扩展名非 .go 当前未参与编译；现行 [ws.go](./ws.go) 仅完成升级并关闭连接
- main.go 中 `ws.Start()` 目前被注释，模块未启用
