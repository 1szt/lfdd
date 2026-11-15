import { defineBoot } from "#q-app";
import { apiPath } from "@/utils/api";
import {
  activeToken,
  ensureGuestIdentity,
  hasToken,
  setGuestToken
} from "@/utils/guest";

// 是否已给全局 fetch 打过补丁（HMR 可能重复执行 boot）
const PATCHED = "__lfddFetchPatched";

// 统一给 /api 请求带上身份（登录 token 或游客 token，均走 Authorization）：
// 未登录且还没有游客身份时先自动申请，游客身份失效时自动补签一次后重试。
export default defineBoot(async () => {
  if (!hasToken()) {
    try {
      await ensureGuestIdentity();
    } catch {
      // 后端不可用时忽略，页面会给出加载失败提示
    }
  }
  patchFetch();
});

function patchFetch() {
  const w = window as unknown as Record<string, unknown>;
  if (w[PATCHED]) return;
  w[PATCHED] = true;

  const rawFetch = window.fetch.bind(window);

  window.fetch = async (input, init = {}) => {
    const url =
      typeof input === "string"
        ? input
        : input instanceof Request
          ? input.url
          : String(input);
    if (!url.includes("/api/")) return rawFetch(input, init);

    // 前后端分离部署时把 /api 相对路径改写到 VITE_API_BASE（同源部署时 apiPath 原样返回）
    const target = typeof input === "string" ? apiPath(input) : input;

    const send = () => {
      const headers = new Headers(init.headers);
      const token = activeToken();
      if (token && !headers.has("Authorization"))
        headers.set("Authorization", `Bearer ${token}`);
      return rawFetch(target, { ...init, headers });
    };

    let res = await send();

    // 游客身份失效（已注销/服务端无此身份）：自动补签一次后重试
    if (res.status === 401 && !hasToken()) {
      const data = await res
        .clone()
        .json()
        .catch(() => ({}) as { need_guest?: boolean });
      if (data && data.need_guest) {
        setGuestToken("");
        try {
          await ensureGuestIdentity();
          res = await send();
        } catch {
          // 补签失败：保留原 401 响应，由页面提示
        }
      }
    }
    return res;
  };
}
