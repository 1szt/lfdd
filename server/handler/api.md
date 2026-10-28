# Handler 接口文档

`handler/` 目录全部接口的文档。路由注册统一在 `../router/router.go`，本目录只放处理函数。

# 路由总览

| 路由 | 方法 | 处理函数 | 所在文件 | 说明 |
| --- | --- | --- | --- | --- |
| /api/ | GET | GetAPI | api.go | 探活，返回 "Hello, API!" |
| /api/auth/login | POST | PostLogin | auth.go | 登录 |
| /api/auth/register | POST | PostRegister | auth.go | 注册（service.Register 落库） |
| /api/auth/logout | POST | PostLogout | auth.go | 退出 |
| /api/auth/refresh | POST | PostRefresh | auth.go | 刷新 token |
| /api/menu | GET | GetMenu | menu.go | 菜单（内置模拟数据） |
| /api/install/check | GET | GetCheck | install.go | 检查是否已初始化 |
| /api/install/submit | POST | PostSubmit | install.go | 提交初始化配置 |
| /api/test/reload | GET | GetPage | test.go | 热重载测试页（HTML） |
| /api/test/reload/status | GET | GetStatus | test.go | 服务与数据库状态 |
| /api/test/reload/apply | POST | PostApply | test.go | 触发 server + db 重载 |
| /api/ws/ | GET | GetWs | ws.go | WebSocket 连接（query 带 token） |

# 认证接口（auth.go）

- login：当前仅测试账号（username=1 / password=1 返回 token=1），接数据库后改为查 User + bcrypt 比对
- register：绑定 username/email/password → `service.Register`（查重 → bcrypt 哈希 → 写 users 表）
- logout / refresh：占位，接 Session 表后实现

# 菜单接口（menu.go）

返回：

```json
{
  "code": 200,
  "message": "success",
  "categories": [
    {
      "id": 1,
      "name": "热销推荐",
      "goods": [
        { "id": 101, "name": "招牌炒饭", "icon": "🍚", "desc": "大火爆炒 粒粒分明", "price": 18 }
      ]
    }
  ]
}
```

数据结构：CategoryItem（分类：ID、Name、Goods）/ GoodsItem（商品：ID、Name、Icon、Desc、Price，单位元）。

数据为内置模拟数据 `menuData`，共 4 个分类（热销推荐/经典套餐/凉菜小食/饮品甜点），接入数据库时替换 GetMenu 的数据来源即可。

# 安装初始化（install.go）

## GET /api/install/check

返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "need_install": true }
}
```

`need_install=true` 表示需要初始化（`LFDD_SERVER_INSTALL` 不存在、为空或 `false`）。

## POST /api/install/submit

请求结构 InstallRequest：

| 字段 | 说明 |
| --- | --- |
| server_ip | 服务监听地址 |
| server_port | 服务端口 |
| db_type | 数据库类型（sqlite/pgsql/mysql） |
| db_path | SQLite 数据库文件路径（sqlite 必填） |
| db_host / db_port / db_name / db_user / db_password | 非 sqlite 数据库必填 |

行为：

1. 防重复提交：`LFDD_SERVER_INSTALL=true` 时直接拒绝（400）
2. 写入配置到 `data/.env`（最后才写 `LFDD_SERVER_INSTALL=true`，中途失败视为未初始化）
3. `db.Reload()` 重载数据库连接
4. 先返回响应，再异步 `server.Restart()` 热切换监听地址（同步调用会阻塞到 Shutdown 超时）

未初始化时本组接口由 InitGuard 中间件放行（`/api/install/` 前缀在白名单内）。

# 热重载测试页（test.go）

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

使用场景：验证配置热重载链路（改 `data/.env` 后点重载，观察状态变化与日志），无需重启进程。

# WebSocket（ws.go）

GET /api/ws/：从 query 取 `token`、`c.ClientIP()`，升级为 WebSocket 连接（AcceptOptions 设 `InsecureSkipVerify: true` 允许跨域）。当前仅完成升级，连接即关闭；消息收发逻辑待扩展。
