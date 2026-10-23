# Server 代码整理记录

按时间倒序记录 server 后端的重构/整理动作，只做代码质量优化，不改变目录结构。

# 2026-10-22 模块命名与结构梳理

模块全部按职责命名，消除与第三方库撞名的历史遗留：

gin 模块 → server 模块 → web + server 拆分
原来叫 gin，与 github.com/gin-gonic/gin 撞名，调用方被迫用别名。拆成两个模块：web 持有 Engine 与全局中间件（初始化职责），server 负责服务启动/重启/热重载。

gorm 模块 → db 模块
原来叫 gorm，与 gorm.io/gorm 撞名。改名 db，日志前缀统一 [DB]，修正原来两处 [GROM] 拼写错误。

initpkg 模块 → web 模块
中间件注册模块，原来叫 initpkg 看不出职责。

api 模块子包化
单包改为业务域子包（install、menu），为上百个 API 的规模做准备：

api.go 只做汇总注册，调用各子包 Register(rg)
子包只导出 Register(rg *gin.RouterGroup)，路由组由 api.go 注入，不直接依赖 web.Engine
子包之间禁止互相 import，边界由编译器强制

全局缩写清理
代码里的 g 缿写全部还原为标准包名 gin。

# 2026-10-22 代码质量修复（本次）

config 模块
Get：配置不存在（ErrRecordNotFound）不再打错误日志，返回零值。首次安装前大量缺配置是正常状态，原来每次读取都刷 ❌ 日志。
内部 DB 变量改为小写 db，不再对外导出（原本就无外部引用）。
Start：panic 信息带上底层错误详情，原来只 panic 一个裸字符串，排障时看不到真实原因。
Set：返回的 error 不再内嵌 emoji（emoji 属于日志展示层，不属于错误值）。

db 模块
提取 configFingerprint()，消除 Reload 与 connect 里重复拼接配置指纹的代码。
删除 log.Fatal 后面不可达的 return。

server 模块
加 sync.Mutex 保护 server / lastAddr：Restart 可从 HTTP handler（安装提交）并发触发，与 Status、Reload 存在数据竞态。
lastConfig 改名 lastAddr（它存的是地址不是配置），删除与 start 重复的赋值。
Status 用 net.SplitHostPort 替代手写倒序扫描字符串解析端口。
提取 addrFromConfig()，消除 Start/Restart/Reload 三处重复的读配置拼地址逻辑。

api/install 模块
config.Set 的 error 原来全部丢弃：写配置失败会静默继续，最后照样写入 server_install=true，导致半初始化状态被标记为完成。现在逐步检查，失败即中断返回 500，且 server_install 永远最后写。

motd 模块
修正横幅中英文混排导致的错位。

# 2026-10-23 后续优化落地

把上一节"后续可做"的 4 项全部实施：

server 模块实现 HTTPS/TLS 监听
新增 tlsFromConfig()：读 server_https / server_cert / server_key，开启 HTTPS 但证书密钥缺任一时打警告并回退 HTTP。start() 按开关走 ListenAndServeTLS 或 ListenAndServe。监听状态改为"地址+TLS 指纹"（listenDesc），Reload 对比指纹，证书路径变化也能触发重载；Status 解析时截掉指纹部分。

web 模块支持 ReleaseMode（后按需删除）
曾新增 init()：环境变量 LFDD_MODE=release 时调用 gin.SetMode。后应用户要求删除，gin 保持默认 DebugMode，启动会打印路由注册信息。

api/install 防重复提交
InstallSubmit 开头检查 server_install == "true"，已初始化直接返回 400，避免重复提交覆盖线上配置。

config 模块内存缓存
Get 走 sync.RWMutex 保护的 map 缓存，命中则不查库；未命中查库后回填。Set 写库成功后同步 putCache，保证读写一致。

新增 InitGuard 全局初始化检查中间件
未初始化（server_install != true）时拦截白名单外所有请求，返回 503 + code 40001 + need_install。白名单：/（安装页入口）、/api/install/（安装接口自身）。
中间件挪入 web/middleware 子包（为上百个中间件做结构准备）：一个中间件一个文件、按域拆子包、web.Start() 唯一装配点、依赖单向（middleware → web 禁止）。演进规则见 web/README.md。

# 后续可做（未动）

首个跨模块业务（如用户管理）出现时再设计领域模块：领域层（如 user/）持有业务逻辑与查询导出函数供他人 import，api/user 只做 HTTP 皮；进程内复用走函数调用不走 HTTP，api 子包间仍禁止互相 import
建 web/resp 统一响应出口（OK/Fail 封装 code+message+data，等接口多了再做）
config.Get 的缓存目前进程内有效，若未来出现外部改库（绕过 Set）的场景需要加过期时间或刷新入口
server_https_auto 字段已存在但未实现 ACME 自动签发证书，目前 HTTPS 需手动提供证书路径
