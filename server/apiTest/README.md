# apiTest 模块

热重载测试页：手动触发 server/db 重载并观察效果，挂载于 `/api/test/*`。须在 `web.Start()` 之后调用。

# 接口列表

GET /api/test/reload 测试页面（HTML）
GET /api/test/reload/status 查询服务与数据库状态
POST /api/test/reload/apply 触发 server + db 重载

# 接口说明

## GET /api/test/reload

返回内置 HTML 测试页，展示服务状态、监听地址、数据库类型/状态，提供"重载"按钮与操作日志。

## GET /api/test/reload/status

返回：

```json
{
  "code": 0,
  "data": {
    "server": { "ip": "", "port": "", "status": "running" },
    "db": { "type": "", "status": "connected" }
  }
}
```

字段来自 `server.Status()` 与 `db.Status()`。

## POST /api/test/reload/apply

异步执行 `server.Reload()` 与 `db.Reload()` 后立即返回（Shutdown 会等待在途请求，同步调用会阻塞到超时）。重载期间请求会短暂失败，页面自动轮询直到服务恢复。

# 使用场景

验证配置热重载链路（改 `data/.env` 后点重载，观察状态变化与日志），无需重启进程。
