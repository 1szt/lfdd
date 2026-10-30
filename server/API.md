# API 使用文档

后端全部 HTTP 接口的使用说明（请求/响应结构、初始化流程、拦截器约定）。

- 基础地址：`https://{server_ip}:{server_port}`（默认 `9081`）
- 所有接口为 JSON 交互（除探活/测试页的纯文本/HTML 接口）
- 未初始化时，除下方"白名单"外的所有接口都会被全局中间件拦截（见 [初始化流程](#2-初始化流程)）
- 路由注册与处理函数同在 `handler/`（一功能一文件：注册 + 处理）

***

## 1. 接口总览

| 方法 | 路径 | 说明 | 未初始化时 | 处理函数 | 所在文件 |
| ---- | --------------------- | ----------------------- | ------- | ---- | ---- |
| GET  | `/api/`               | API 探活，返回 `Hello, API!`    | 拦截 | GetAPI | api.go |
| POST | `/api/auth/login`     | 登录                      | 拦截 | PostLogin | apiauth.go |
| POST | `/api/auth/register`  | 注册（落库）                  | 拦截 | PostRegister | apiauth.go |
| POST | `/api/auth/logout`    | 退出                      | 拦截 | PostLogout | apiauth.go |
| POST | `/api/auth/refresh`   | 刷新 token                | 拦截 | PostRefresh | apiauth.go |
| GET  | `/api/install/check`  | 检查是否需要初始化               | 放行（白名单） | GetCheck | apiinstall.go |
| POST | `/api/install/submit` | 提交初始化配置                 | 放行（白名单） | PostSubmit | apiinstall.go |
| GET  | `/api/menu`           | 菜单数据（当前为 mock）          | 拦截 | GetMenu | apimenu.go |
| GET  | `/api/test/reload`    | 热重载测试页（HTML）            | 拦截 | GetPage | apitest.go |
| GET  | `/api/test/reload/status` | 服务与数据库状态            | 拦截 | GetStatus | apitest.go |
| POST | `/api/test/reload/apply` | 触发 server + db 重载     | 拦截 | PostApply | apitest.go |
| GET  | `/api/ws/`            | WebSocket 连接（query 带 token） | 拦截 | GetWs | apiws.go |

***

## 2. 初始化流程

后端启动时若未初始化（`LFDD_SERVER_INSTALL != true`），全局中间件 `InitGuard` 会拦截**白名单以外**的所有请求。

```text
前端启动
  │
  ▼
GET /api/install/check ──► need_install=true ──► 前端自行跳转安装页
  │
  ▼
用户填写表单
  │
  ▼
POST /api/install/submit ──► 200 初始化完成
  │
  ├── 写入全部配置（ip/端口/数据库）
  ├── LFDD_SERVER_INSTALL = true（拦截器自动解除）
  ├── 数据库连接热重载（db.Reload，按配置指纹判断是否重连）
  └── 服务热切换到新端口（server.Restart，异步执行）
  │
  ▼
后续请求正常放行
```

### 2.1 GET /api/install/check

检查服务器是否需要初始化。

**响应**（HTTP 200）：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "need_install": true
  }
}
```

`need_install`：`true` 需要初始化（`LFDD_SERVER_INSTALL` 不存在、为空或 `false`），`false` 已初始化。

### 2.2 POST /api/install/submit

提交初始化配置。**已初始化后再次调用会被拒绝（防重复提交），不会覆盖线上配置。**

**请求体**（`application/json`）：

| 字段           | 类型     | 必填          | 说明                                    |
| ------------ | ------ | ----------- | ------------------------------------- |
| server\_ip   | string | 是           | 服务监听 IP（如 `0.0.0.0`）                  |
| server\_port | string | 是           | 服务端口（如 `9081`）                        |
| db\_type     | string | 是           | 数据库类型：`sqlite` / `mysql` / `postgres` |
| db\_path     | string | sqlite 必填   | sqlite 文件路径                           |
| db\_host     | string | mysql/pg 必填 | 数据库主机                                 |
| db\_port     | string | mysql/pg 必填 | 数据库端口                                 |
| db\_name     | string | mysql/pg 必填 | 数据库名                                  |
| db\_user     | string | mysql/pg 必填 | 数据库用户                                 |
| db\_password | string | mysql/pg 必填 | 数据库密码                                 |

**成功响应**（HTTP 200）：

```json
{
  "code": 0,
  "message": "初始化完成"
}
```

响应返回后，后端会异步执行：

1. **数据库热重载**：配置指纹有变化才真正重连，无变化跳过
2. **服务热切换**：读取新 `server_ip:server_port`，先优雅关闭旧监听再启动新监听；**若端口没变则行为不变，前端无感知**
3. **拦截器解除**：`LFDD_SERVER_INSTALL` 已写入，之后的请求直接放行

**失败响应**（HTTP 500）：

```json
{
  "code": -1,
  "message": "写入配置失败: server_port"
}
```

部分字段写入失败时按 `message` 里的字段名排查，配置写入是顺序执行的，失败即中断。

### 2.3 重复提交保护

已初始化（`LFDD_SERVER_INSTALL == "true"`）时调用 submit（HTTP 400）：

```json
{
  "code": -1,
  "message": "服务器已初始化，不能重复提交"
}
```

***

## 3. 未初始化拦截

**触发条件**：`LFDD_SERVER_INSTALL != true` 且路径不在白名单内。

**白名单**（见 [init_guard.go](file:///c:/Users/jingxuanwei/Desktop/code/lfdd/server/middleware/init_guard.go)）：

- `/` —— 预留安装页入口/健康检查
- 前缀 `/api/install/` —— 初始化接口自身

**拦截响应**（HTTP **503**）：

```json
{
  "code": 40001,
  "message": "服务器未初始化，请先完成初始化",
  "need_install": true
}
```

**前端约定**：

- 收到 `code == 40001`（或 `need_install == true`）→ 自行跳转安装页，不区分具体被拦截的接口
- 与 `GET /api/install/check` 的判断逻辑复用同一个 `need_install` 字段
- 拦截发生在业务逻辑之前，接口本身不需要重复判断初始化状态

***

## 4. 认证接口（/api/auth/*）

### 4.1 POST /api/auth/register

注册新用户（查重 → bcrypt 哈希 → 写 users 表）。

**请求体**：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| username | string | 用户名，唯一 |
| email | string | 邮箱，唯一 |
| password | string | 密码 |

**成功响应**（HTTP 200）：

```json
{
  "message": "注册成功",
  "user": {
    "id": "1780000000000000000",
    "username": "demo",
    "email": "demo@example.com",
    "created_at": "2026-01-01T00:00:00+08:00",
    "updated_at": "2026-01-01T00:00:00+08:00"
  }
}
```

`id` 为 18 位数字 ID（JSON 中是字符串，防 JS 精度丢失）；密码字段永不返回。

**失败响应**（HTTP 400）：

```json
{ "message": "用户名或邮箱已存在" }
```

数据库未连接（未初始化）时同样返回 400。

### 4.2 POST /api/auth/login

当前仅测试账号：`username=1`、`password=1` 时返回 `{"token": "1"}`（HTTP 200）。

失败返回 HTTP 401：`{"message": "登录失败"}`。接数据库后改为查 User + bcrypt 比对 + 写 Session 发 token。

### 4.3 logout / refresh

占位接口，接 Session 表后实现。

***

## 5. GET /api/menu

返回菜品分类与商品（当前为内置模拟数据 `menuData`，共 4 个分类）。

**响应**（HTTP 200）：

```json
{
  "code": 200,
  "message": "success",
  "categories": [
    {
      "id": 1,
      "name": "热销推荐",
      "goods": [
        {
          "id": 101,
          "name": "招牌炒饭",
          "icon": "🍚",
          "desc": "大火爆炒 粒粒分明",
          "price": 18
        }
      ]
    }
  ]
}
```

数据结构：CategoryItem（分类：id、name、goods）/ GoodsItem（商品：id、name、icon、desc、price，单位元）。接入数据库时替换 GetMenu 的数据来源即可。

***

## 6. 热重载测试页（/api/test/*）

### 6.1 GET /api/test/reload

返回内置 HTML 测试页：展示服务状态、监听地址、数据库类型/状态，提供"重载"按钮与操作日志，自动轮询状态。

### 6.2 GET /api/test/reload/status

**响应**（HTTP 200）：

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

### 6.3 POST /api/test/reload/apply

异步执行 `server.Reload()` 与 `db.Reload()` 后立即返回（Shutdown 会等待在途请求，同步调用会阻塞到超时）。重载期间请求会短暂失败，页面自动轮询直到服务恢复。

使用场景：验证配置热重载链路（改 `data/.env` 后点重载，观察状态变化与日志），无需重启进程。

***

## 7. WebSocket（GET /api/ws/）

从 query 取 `token`、`c.ClientIP()`，升级为 WebSocket 连接（`InsecureSkipVerify: true` 允许跨域）。当前仅完成升级，连接即关闭；消息收发逻辑待扩展。

***

## 8. 业务码约定

| code  | 含义                     |
| ----- | ---------------------- |
| 0     | 成功（install / test 系列）  |
| 200   | 成功（menu 系列，沿用 HTTP 风格） |
| 400   | 参数错误 / 重复提交 / 业务校验失败   |
| 401   | 登录失败                   |
| 40001 | 未初始化拦截                 |
| -1    | 服务端错误（写库、配置失败）         |

> 注意：`40001` 只出现在**被拦截**的响应里，业务接口自身不会返回它。后续新增业务码在本节集中登记。

***

## 9. 接口文档维护方式

接口文档就是本文件（API.md），手写维护：改了接口顺手更新对应章节即可，响应示例直接贴 `curl` 或前端实际收到的 JSON。

新增接口时按现有章节结构补一节：请求方法 + 路径、参数表、响应示例（json 代码块）、错误码，并在第 1 节总览表加一行。
