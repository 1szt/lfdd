// 后端地址绑定：构建时通过 VITE_API_BASE 注入（如 https://api.example.com）。
// 留空则表示前端与后端同源（开发代理或生产反向代理），请求使用相对路径。
export const API_BASE = String(import.meta.env.VITE_API_BASE ?? "").replace(
  /\/+$/,
  ""
);

// apiPath 把以 /api 开头的相对地址拼上后端地址；
// 已是完整 URL（http/https）或非 /api 路径原样返回。
export function apiPath(url: string): string {
  if (!API_BASE || !url.startsWith("/api")) return url;
  return API_BASE + url;
}
