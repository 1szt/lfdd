/**
 * 自定义环境变量类型声明（Quasar CLI 会自动生成标准变量，这里只声明项目自定义项）。
 * 变量值来自构建时的 .env 文件 / CI 环境变量。
 */
interface ImportMetaEnv {
  // 后端地址：如 https://api.example.com；留空表示前后端同源（走反向代理）
  readonly VITE_API_BASE?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
