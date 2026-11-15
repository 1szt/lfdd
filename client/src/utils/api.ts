// 后端地址绑定：
// 开发环境固定走 devServer 代理（/api → http://127.0.0.1:34334），请求保持相对路径；
// 生产环境使用构建时通过 VITE_API_BASE 注入的绝对地址（如 https://api.example.com），
// 留空则表示与后端同源（生产反向代理）。
export const API_BASE = import.meta.env.DEV
  ? ""
  : String(import.meta.env.VITE_API_BASE ?? "").replace(/\/+$/, "");

// apiPath 把以 /api 开头的相对地址拼上后端地址；
// 已是完整 URL（http/https）或非 /api 路径原样返回。
export function apiPath(url: string): string {
  if (!API_BASE || !url.startsWith("/api")) return url;
  return API_BASE + url;
}
