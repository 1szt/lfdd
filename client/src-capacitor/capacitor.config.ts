import { defineCapacitorConfig } from "@quasar/app-vite/capacitor";

// App 远程加载的前端地址：构建时通过 LFDD_APP_SERVER_URL 注入（GitHub Actions 里用仓库变量传入）。
// 未设置时回退到局域网默认地址，保持本地打包行为不变。
const serverUrl =
  process.env.LFDD_APP_SERVER_URL?.trim() || "http://192.168.2.2:9500";

export default defineCapacitorConfig({
  appId: "com.szt.lfdd",
  appName: "lfdd",
  server: {
    url: serverUrl,
    cleartext: true
  }
});
