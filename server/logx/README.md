# logx 模块

日志模块：标准库 `log` 与 gin 访问日志统一输出（控制台 + 可选写盘）。

# 使用方法

logx.Start() 初始化日志输出

# 调用顺序

在 `env.Start()` 之前调用（main 中紧跟 `motd.Start()`）。它直接从 `data/.env` 读 `LFDD_LOG_PATH`（不依赖 env 补全），为空则使用默认路径 `data/log/lfdd.log`，保证后续模块日志都能落盘。

# 行为

- 控制台始终输出
- 同时追加写入日志文件，路径为 `LFDD_LOG_PATH`（未配置或为空时用默认 `data/log/lfdd.log`）
- 打开文件失败时降级为仅控制台，并打印告警
- 同时接管 `log.SetOutput`、`gin.DefaultWriter`、`gin.DefaultErrorWriter`

# 日志轮转

- 单文件上限 `maxLogSize = 50MB`
- 仅在启动时执行：超过上限则重命名为 `.old`（旧 `.old` 被覆盖），只保留一份
