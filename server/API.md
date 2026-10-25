# API 使用文档

后端全部 HTTP 接口的使用说明（请求/响应结构、初始化流程、拦截器约定）。s

- 基础地址：`https://{server_ip}:{server_port}`（默认 `9081`）
- 所有接口为 JSON 交互（除主页/探活的纯文本接口）
- 未初始化时，除下方"白名单"外的所有接口都会被全局中间件拦截（见 [初始化流程](#2-初始化流程)）

***

## 1. 接口总览

| 方法   | 路径                    | 说明                      | 未初始化时   |
| ---- | --------------------- | ----------------------- | ------- |
| GET  | `/`                   | 主页探活，返回 `Hello, World!` | 放行（白名单） |
| GET  | `/api/`               | API 探活，返回 `API`         | 拦截      |
| GET  | `/api/install/check`  | 检查是否需要初始化               | 放行（白名单） |
| POST | `/api/install/submit` | 提交初始化配置                 | 放行（白名单） |
| GET  | `/api/menu`           | 菜单数据（当前为 mock）          | 拦截      |

***

## 2. 初始化流程

后端启动时若未初始化（`server_install != true`），全局中间件 `InitGuard` 会拦截**白名单以外**的所有请求。

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
  ├── server_install = true（拦截器自动解除）
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

`need_install`：`true` 需要初始化，`false` 已初始化。

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
| user         | string | 是           | 操作者（写入配置的 user 字段）                    |

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
3. **拦截器解除**：`server_install` 已写入，之后的请求直接放行

**失败响应**（HTTP 500）：

```json
{
  "code": -1,
  "message": "写入配置失败: server_port"
}
```

部分字段写入失败时按 `message` 里的字段名排查，配置写入是顺序执行的，失败即中断。

### 2.3 重复提交保护

已初始化（`server_install == "true"`）时调用 submit：

```json
{
  "code": 400,
  "message": "服务器已完成初始化，禁止重复提交"
}
```

***

## 3. 未初始化拦截

**触发条件**：`server_install != true` 且路径不在白名单内。

**白名单**（见 [middleware.go](file:///c:/Users/jingxuanwei/Desktop/code/lfdd/server/web/middleware/init_guard.go)）：

- `/` —— 安装页入口（静态页）
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

## 4. GET /api/menu

返回菜品分类与商品（当前为 mock 数据）。

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

***

## 5. 业务码约定

| code  | 含义                     |
| ----- | ---------------------- |
| 0     | 成功（install 系列）         |
| 200   | 成功（menu 系列，沿用 HTTP 风格） |
| 400   | 参数错误 / 重复提交            |
| 40001 | 未初始化拦截                 |
| -1    | 服务端错误（写库、配置失败）         |

> 注意：`40001` 只出现在**被拦截**的响应里，业务接口自身不会返回它。后续新增业务码在 [REFACTOR.md](file:///c:/Users/jingxuanwei/Desktop/code/lfdd/server/REFACTOR.md) 集中登记。

***

## 6. 接口文档维护方式

接口文档就是本文件（API.md），手写维护：改了接口顺手更新对应章节即可，响应示例直接贴 `curl` 或前端实际收到的 JSON。

新增接口时按现有章节结构补一节：请求方法 + 路径、参数表、响应示例（json 代码块）、错误码。
