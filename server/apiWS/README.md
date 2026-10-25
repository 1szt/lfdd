# apiWS 模块

WebSocket 模块：升级 HTTP 为 WebSocket 连接，挂载于 `/ws`。须在 `web.Start()` 之后调用。

# 接口列表

GET /ws WebSocket 连接（query 带 token）

# 函数

apiWS.Start()
注册 `/ws` 路由，从 query 取 `token`、`c.ClientIP()` 后升级为 WebSocket（AcceptOptions 设 InsecureSkipVerify: true 允许跨域）

apiWS.Send(id string, msg []byte)
发送消息给指定 id 的客户端

# 注意

- 客户端 id 分配与 Send 实现位于 [hub.go1](./hub.go1)，扩展名非 .go 当前未参与编译；现行 [apiWS.go](./apiWS.go) 仅完成升级并关闭连接
