# API 模块

API 模块是 HTTP 接口的总入口：创建 /api 路由组，汇总注册各业务子包的路由。

# 目录约定

每个业务域一个子包，目录层级天然表达接口分组（如 api/install → /api/install/*）：

api/
├── api.go       总入口：建路由组，依次调用各子包 Register
├── install/     安装初始化接口
└── menu/        菜单接口

子包约定：
1. 只导出 Register(rg *gin.RouterGroup)，由 api.go 传入路由组，子包不直接依赖 web.Engine
2. 子包之间禁止互相 import
3. 子包内代码膨胀时按文件拆分（handler.go / model.go / service.go），保持单一 package
4. 新增接口 = 新建子包 + api.go 里加一行 Register 调用

# 接口列表

GET /api/ 主页接口
GET /api/menu 菜单数据
GET /api/install/check 检查是否需要初始化
POST /api/install/submit 提交初始化配置
