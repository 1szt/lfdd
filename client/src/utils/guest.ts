// 游客身份：未登录访问时由后端签发（POST /api/auth/guest），与登录 token 共用 Authorization 头，
// 游客 token 以 guest_ 开头，服务端按前缀分流。未持有效身份的请求，业务接口一律不返回数据
// （401 + need_guest）；登录后忽略并作废游客身份，也可调用 revokeGuestIdentity 主动注销。

// 登录态存储键（与 admin 页共用）
export const TOKEN_KEY = "lfdd_token";

const GUEST_KEY = "lfdd_guest_token";

// 读取登录 token（空串表示未登录）
export function getToken(): string {
  try {
    return localStorage.getItem(TOKEN_KEY) ?? "";
  } catch {
    return "";
  }
}

// 是否已登录
export function hasToken(): boolean {
  return getToken() !== "";
}

// 读取本地游客 token（空串表示还没申请）
export function getGuestToken(): string {
  try {
    return localStorage.getItem(GUEST_KEY) ?? "";
  } catch {
    return "";
  }
}

// 保存/清除游客 token
export function setGuestToken(token: string) {
  try {
    if (token) localStorage.setItem(GUEST_KEY, token);
    else localStorage.removeItem(GUEST_KEY);
  } catch {
    // 本地存储不可用：仅本次会话有效
  }
}

// 当前生效的身份 token：已登录用登录 token，否则用游客 token
export function activeToken(): string {
  return getToken() || getGuestToken();
}

// 是否游客：未登录且有游客身份（游客不显示/不使用结单、打印、改价、废单，也看不到已结单）
export function isGuestMode(): boolean {
  return !hasToken() && getGuestToken() !== "";
}

// 申请游客身份（已有则直接返回）：唯一允许无身份调用的接口
export async function ensureGuestIdentity(): Promise<string> {
  const token = getGuestToken();
  if (token) return token;

  const res = await fetch("/api/auth/guest", { method: "POST" });
  const data = await res.json().catch(() => ({}));
  if (!res.ok || typeof data.guest_id !== "string" || !data.guest_id) {
    throw new Error(data.message || `申请游客身份失败（${res.status}）`);
  }
  setGuestToken(data.guest_id);
  return data.guest_id;
}

// 主动注销游客身份：注销后该身份立即失效，下次请求会自动重新申请
export async function revokeGuestIdentity(): Promise<void> {
  const token = getGuestToken();
  if (!token) return;
  try {
    await fetch("/api/auth/guest/revoke", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` }
    });
  } catch {
    // 后端不可用时只清本地身份
  }
  setGuestToken("");
}

// 登录成功：清掉本地游客身份（后端登录时已作废该身份，登录后一律以登录身份为准）
export function clearGuestIdentity() {
  setGuestToken("");
}
