# apiinstall 模块

安装初始化接口：检查/提交服务器初始化配置，挂载于 `/api/install/*`。须在 `web.Start()` 之后调用。

# 接口列表

GET /api/install/check 检查服务器是否已初始化
POST /api/install/submit 提交初始化配置

# 接口说明

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

# 前置依赖

未初始化时该接口由 web 模块的 InitGuard 中间件放行（`/api/install/` 前缀在白名单内）。
