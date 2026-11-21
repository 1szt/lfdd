import { defineCapacitorConfig } from "@quasar/app-vite/capacitor";

// 前端资源打包进 App（assets/public），不远程加载页面；
// 接口地址由构建时的 VITE_API_BASE 注入，见 src/utils/api.ts。
export default defineCapacitorConfig({
  appId: "com.szt.lfdd",
  appName: "lfdd"
});
