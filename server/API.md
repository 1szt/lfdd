# API 使用文档

后端全部 HTTP 接口的使用说明（请求/响应结构、初始化流程、拦截器约定）。

- 基础地址：`http://{server_ip}:{server_port}`（默认 `9081`）
- 所有接口为 JSON 交互（除主页/探活的纯文本接口）
- 未初始化时，除下方"白名单"外的所有接口都会被全局中间件拦截（见 [初始化流程](#2-初始化流程)）

---

## 1. 接口总览

| 方法 | 路径 | 说明 | 未初始化时 |
| --- | --- | --- | --- |
| GET | `/` | 主页探活，返回 `Hello, World!` | 放行（白名单） |
| GET | `/api/` | API 探活，返回 `API` | 拦截 |
| GET | `/api/install/check` | 检查是否需要初始化 | 放行（白名单） |
| POST | `/api/install/submit` | 提交初始化配置 | 放行（白名单） |
| GET | `/api/menu` | 菜单数据（当前为 mock） | 拦截 |

---

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

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| server_ip | string | 是 | 服务监听 IP（如 `0.0.0.0`） |
| server_port | string | 是 | 服务端口（如 `9081`） |
| db_type | string | 是 | 数据库类型：`sqlite` / `mysql` / `postgres` |
| db_path | string | sqlite 必填 | sqlite 文件路径 |
| db_host | string | mysql/pg 必填 | 数据库主机 |
| db_port | string | mysql/pg 必填 | 数据库端口 |
| db_name | string | mysql/pg 必填 | 数据库名 |
| db_user | string | mysql/pg 必填 | 数据库用户 |
| db_password | string | mysql/pg 必填 | 数据库密码 |
| user | string | 是 | 操作者（写入配置的 user 字段） |

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

---

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

---

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

---

## 5. 业务码约定

| code | 含义 |
| --- | --- |
| 0 | 成功（install 系列） |
| 200 | 成功（menu 系列，沿用 HTTP 风格） |
| 400 | 参数错误 / 重复提交 |
| 40001 | 未初始化拦截 |
| -1 | 服务端错误（写库、配置失败） |

> 注意：`40001` 只出现在**被拦截**的响应里，业务接口自身不会返回它。后续新增业务码在 [REFACTOR.md](file:///c:/Users/jingxuanwei/Desktop/code/lfdd/server/REFACTOR.md) 集中登记。

---

## 6. Swagger 接口文档（本地生成）

接口文档用 swaggo 从 handler 注释生成，产物在 `docs/` 目录，**只在本地生成和查看，不在服务里挂载文档路由**（生产服务不暴露接口目录）。

### 6.1 日常工作流（三步）

```text
① 修改 handler 上的注释（见下方模板）
② server 目录双击 swag-init.bat（或手动执行 swag init）
③ 查看 docs/swagger.json（编辑器直接打开，或导入 Apifox/Postman 可视化浏览）
```

注意顺序：**先改注释 → 再 swag init**，生成结果不依赖业务代码，但路由变了页面上的试调路径要对得上。

### 6.2 注释模板

GET 接口（放在 handler 函数上方）：

```go
// GetMenu 获取菜单
// @Summary 获取菜品分类与商品列表
// @Tags 菜单                ← 文档页左侧的分组名
// @Produce json
// @Success 200 {object} MenuResponse
// @Router /menu [get]       ← 相对 @BasePath(/api)，完整路径 = /api/menu
func GetMenu(c *gin.Context) { ... }
```

POST 接口（带请求体）：

```go
// @Accept json
// @Param data body InstallRequest true "初始化参数说明"
// @Success 200 {object} InstallResponse
// @Failure 400 {object} InstallResponse
// @Router /install/submit [post]
```

规则要点：

- `{object}` 引用**当前包内的结构体名**，swag 会自动展开字段（结构体改了字段，重新 `swag init` 即同步）
- `@Tags` 决定文档页分组，新增业务子包时用子包名作 Tag（如 `菜单`、`安装`、`用户`）
- `@Router` 路径不带 `/api` 前缀（`@BasePath` 定义在 [main.go](file:///c:/Users/jingxuanwei/Desktop/code/lfdd/server/main.go) 顶部）

### 6.3 相关文件

| 文件 | 说明 |
| --- | --- |
| `docs/`（docs.go / swagger.json / swagger.yaml） | `swag init` 自动生成，**不要手改**，改注释重新生成即可 |
| `swag-init.bat` | 双击即可执行 `swag init` 的快捷脚本，窗口停住按任意键关闭 |
| main.go 顶部 `// @title` 注释 | 文档标题、版本、@BasePath 全局定义 |

新增/修改接口后忘记 `swag init`，生成的 docs/ 内容不会更新（git 提交时把 `docs/` 一起提交，方便 diff 查看接口变更）。
