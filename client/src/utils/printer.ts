// 蓝牙小票机打印服务
// - 原生（Capacitor Android）：经典蓝牙 SPP + BLE 双通道，走原生插件（WebView 没有 Web Bluetooth API）
// - 桌面 / 网页（Electron、Chrome）：Web Bluetooth（仅 BLE）

// ==================== ESC/POS 指令 ====================
import str2gbk from "str2gbk";

export const ESC_INIT = [0x1b, 0x40]; // ESC @ 初始化打印机
export const ESC_ALIGN_CENTER = [0x1b, 0x61, 0x01]; // ESC a 1 居中对齐
export const ESC_FEED = [0x1b, 0x64, 0x04]; // ESC d 4 走纸 4 行

// ==================== 运行环境 ====================
export function isNative(): boolean {
  return Boolean((window as unknown as { Capacitor?: any }).Capacitor?.isNativePlatform?.());
}

// ==================== 文本编码 ====================
// 连接方式：ble 低功耗 / spp 经典蓝牙
export type PrinterConn = "ble" | "spp";
// 编码格式：gbk 国产热敏机默认 / utf8 部分机型支持
export type PrinterEncoding = "gbk" | "utf8";

// 国产热敏机按 GBK 解释字节流，发 UTF-8 中文会乱码
// str2gbk：纯浏览器 GBK 编码（<1KB，基于 TextDecoder 反查实现）
export function encodeGbk(text: string): Uint8Array {
  return str2gbk(text, { onAlloc: (len) => new Uint8Array(len) });
}

// 按设备编码格式编码文本（utf8 机型发 UTF-8，gbk 机型发 GBK）
export function encodeText(text: string, enc: PrinterEncoding = "gbk"): Uint8Array {
  return enc === "utf8" ? new TextEncoder().encode(text) : encodeGbk(text);
}

// 组装测试小票数据：初始化 + 居中 + 文案 + 换行 + 走纸
export function buildTestPayload(encoding: PrinterEncoding = "gbk"): number[] {
  return [
    ...ESC_INIT,
    ...ESC_ALIGN_CENTER,
    ...encodeText("你好 蓝牙小票机", encoding),
    0x0a,
    0x0a,
    ...ESC_FEED
  ];
}

// ==================== 订单小票 ====================
export interface ReceiptItem {
  name: string;
  spec: string;
  qty: number;
  price: number; // 单价，分
}
export interface ReceiptOrder {
  orderNo: string;
  createdAt: string; // 2026-11-14 12:00:00
  count: number; // 商品总件数
  total: number; // 合计，分
  items: ReceiptItem[];
}

// 按视觉宽度折行（文字可以换行，全字符不切半）
function wrapByWidth(s: string, maxW: number): string[] {
  const rows: string[] = [];
  let cur = "";
  let w = 0;
  for (const ch of s) {
    const cw = displayWidth(ch);
    if (cur !== "" && w + cw > maxW) {
      rows.push(cur);
      cur = ch;
      w = cw;
    } else {
      cur += ch;
      w += cw;
    }
  }
  if (cur !== "") rows.push(cur);
  return rows;
}

// 视觉宽度：全角字符（中日韩等）按 2 列计算，58mm 纸约 32 列
function displayWidth(s: string): number {
  let w = 0;
  for (const ch of s) {
    const code = ch.codePointAt(0) ?? 0;
    // 中日韩表意文字/全角形式 + 全角符号（￥ 等 U+FFE0~EE）
    if ((code > 0x1100 && code <= 0xff60) || (code >= 0xffe0 && code <= 0xffee)) w += 2;
    else w += 1;
  }
  return w;
}
function padEndW(s: string, width: number): string {
  let out = s;
  while (displayWidth(out) < width) out += " ";
  return out;
}
function padStartW(s: string, width: number): string {
  let out = s;
  while (displayWidth(out) < width) out = " " + out;
  return out;
}

// 组装订单小票：初始化 + 居中放大标题 + 单号/时间 + 明细 + 合计 + 走纸切纸
export function buildOrderPayload(o: ReceiptOrder, encoding: PrinterEncoding = "gbk"): number[] {
  const lines: string[] = [
    `${o.orderNo}`,
    "--------------------------------"
  ];
  for (const it of o.items) {
    const label = it.spec ? `${it.name}(${it.spec})` : it.name;
    // 明细：价格固定在右侧列（数量x单价=小计，单价一位小数）；
    // 名称过长时折行续打（文字可错位），价格列不错位
    const unit = (it.price / 100).toFixed(1);
    const subtotal = ((it.price * it.qty) / 100).toFixed(1);
    const right = padStartW(`${it.qty}x${unit}=`, 9) + padStartW(subtotal, 7);
    const nameW = Math.max(32 - displayWidth(right), 4);
    const segs = wrapByWidth(label, nameW);
    for (let j = 0; j < segs.length; j++) {
      // 只有最后一行补空格接价格，前面的名称行顶格换行
      const seg = segs[j]!;
      lines.push(j === segs.length - 1 ? padEndW(seg, nameW) + right : seg);
    }
  }
  lines.push("--------------------------------");
  // 合计：总价右对齐到纸宽（与明细小计同列），按价格宽度动态留空，超长不换行
  const totalStr = `￥${(o.total / 100).toFixed(1)}`;
  lines.push(padEndW("合计", Math.max(32 - displayWidth(totalStr), 4)) + totalStr);
  lines.push(padEndW(`共 ${o.count} 件`, 24));

  const parts: Uint8Array[] = [
    new Uint8Array(ESC_INIT),
    // ESC ! 选字体：本机实测 0x10 = 瘦长体（0x08 是加粗、0x20 变宽），
    // 紧跟在 INIT 后、且不发 GS ! 等可能覆盖样式的大命令（与字体测试票同一路径）
    new Uint8Array([0x1b, 0x21, 0x10]),
    new Uint8Array(ESC_ALIGN_CENTER),
    encodeText(lines[0] ?? "订单小票", encoding),
    new Uint8Array([0x0a]),
    new Uint8Array([0x1b, 0x61, 0x00]), // ESC a 0 左对齐
    encodeText(lines.slice(1).join("\n"), encoding),
    new Uint8Array(ESC_FEED),
    new Uint8Array([0x1d, 0x56, 0x41, 0x03]) // GS V A 3 走纸并切纸
  ];
  const total = parts.reduce((n, p) => n + p.length, 0);
  const payload = new Uint8Array(total);
  let offset = 0;
  for (const p of parts) {
    payload.set(p, offset);
    offset += p.length;
  }
  return Array.from(payload);
}

// ==================== 原生：经典蓝牙 SPP ====================
export interface SppDevice {
  address: string;
  name: string;
}

function sppApi(): any {
  const api = (window as unknown as { bluetoothSerial?: any }).bluetoothSerial;
  if (!api) throw new Error("SPP 插件未加载，请执行 npx cap sync 后重启应用");
  return api;
}

// cordova 插件是回调风格，包一层 Promise 并加超时
function sppCall<T>(
  run: (ok: (v: T) => void, fail: (e: unknown) => void) => void,
  timeoutMs = 15000
): Promise<T> {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error("蓝牙操作超时")), timeoutMs);
    run(
      (v) => {
        clearTimeout(timer);
        resolve(v);
      },
      (e) => {
        clearTimeout(timer);
        reject(new Error(typeof e === "string" ? e : ((e as any)?.message || String(e))));
      }
    );
  });
}

// Android 12+ 运行时权限兜底：老 cordova SPP 插件（2018 年）不申请 BLUETOOTH_CONNECT/SCAN，
// 未授权直接调 getBondedDevices/startDiscovery 会抛 SecurityException → 应用闪退。
// 借 BLE 插件的 initialize（会弹「附近设备」授权并等授权结果）先把权限拿到手。
async function ensureBtPermission(): Promise<void> {
  if (!isNative()) return;
  try {
    await bleApi().initialize({ androidNeverForLocation: true });
  } catch (e) {
    const msg = String((e as any)?.message || e || "");
    if (/BLE is not supported/i.test(msg)) return; // 无 BLE 的设备无法经此渠道授权，只能直接尝试
    if (/permission/i.test(msg)) {
      throw new Error("蓝牙权限未授予，请在系统设置里允许本应用的「附近设备」权限后重试");
    }
    throw e;
  }
}

// 已配对的经典蓝牙设备
export async function sppList(): Promise<SppDevice[]> {
  await ensureBtPermission();
  return sppCall<SppDevice[]>((ok, fail) => sppApi().list(ok, fail));
}

// 经典蓝牙搜索：扫描附近「未配对」设备（不返回已保存的设备）
// onDevice 每发现一个新设备回调一次；扫描结束（约 12 秒）时 resolve 完整列表
export async function sppDiscover(onDevice?: (d: SppDevice) => void): Promise<SppDevice[]> {
  await ensureBtPermission();
  const api = sppApi();
  const found = new Map<string, SppDevice>();
  const emit = (raw: SppDevice) => {
    if (!raw?.address || found.has(raw.address)) return;
    const d: SppDevice = { address: raw.address, name: raw.name || "" };
    found.set(d.address, d);
    onDevice?.(d);
  };
  api.setDeviceDiscoveredListener(emit);
  try {
    const list = await sppCall<SppDevice[]>((ok, fail) => api.discoverUnpaired(ok, fail), 30000);
    list?.forEach(emit);
    return Array.from(found.values());
  } finally {
    try {
      api.clearDeviceDiscoveredListener();
    } catch {
      // 忽略：部分版本无此方法
    }
  }
}

// 经典蓝牙打印：连接 → 写入 → 等排空 → 断开，返回设备名
export async function sppPrint(device: SppDevice, payload: number[]): Promise<string> {
  await ensureBtPermission(); // connect 同样需要 BLUETOOTH_CONNECT，缺失会闪退
  const api = sppApi();
  await sppCall<void>((ok, fail) => api.connect(device.address, ok, fail));
  try {
    // 连接回调刚返回时底层通道未必就绪，稍等再写，减少首包丢失
    await new Promise((r) => setTimeout(r, 150));
    // 插件的 write 是「入队即 success」（BluetoothSerial.java:128-132），此时数据
    // 还在队列里没发到打印机，立刻 disconnect 会把 socket 关掉 → 提示已发送但不打印。
    // 分块写入 + 按包大小等排空（SPP 速率远高于此，够用），再断开。
    const chunkSize = 512;
    const drainMs = Math.min(1000, 300 + Math.ceil(payload.length / 8));
    for (let off = 0; off < payload.length; off += chunkSize) {
      if (off > 0) await new Promise((r) => setTimeout(r, 100));
      const slice = new Uint8Array(payload.slice(off, off + chunkSize));
      await sppCall<void>((ok, fail) => api.write(slice.buffer, ok, fail));
    }
    await new Promise((r) => setTimeout(r, drainMs));
    return device.name || "未知设备";
  } finally {
    // 尽力断开，失败忽略
    sppCall<void>((ok, fail) => api.disconnect(ok, fail), 3000).catch(() => undefined);
  }
}

// ==================== 原生：BLE ====================
interface BleCharacteristic {
  uuid: string;
  properties: { write?: boolean; writeWithoutResponse?: boolean };
}
interface BleService {
  uuid: string;
  characteristics: BleCharacteristic[];
}
interface BleApi {
  initialize(o?: { androidNeverForLocation?: boolean }): Promise<void>;
  requestDevice(o?: Record<string, unknown>): Promise<{ deviceId: string; name?: string }>;
  connect(o: { deviceId: string }): Promise<void>;
  disconnect(o: { deviceId: string }): Promise<void>;
  getServices(o: { deviceId: string }): Promise<{ services: BleService[] }>;
  getMtu(o: { deviceId: string }): Promise<{ value: number }>;
  write(o: { deviceId: string; service: string; characteristic: string; value: string }): Promise<void>;
  writeWithoutResponse(o: {
    deviceId: string;
    service: string;
    characteristic: string;
    value: string;
  }): Promise<void>;
  // 蓝牙搜索
  addListener(event: string, cb: (r: any) => void): Promise<{ remove(): Promise<void> }>;
  requestLEScan(o: Record<string, unknown>): Promise<void>;
  stopLEScan(): Promise<void>;
}

function bleApi(): BleApi {
  const api = (window as unknown as { Capacitor?: any }).Capacitor?.Plugins?.BluetoothLe;
  if (!api) throw new Error("BLE 插件未加载，请执行 npx cap sync 后重启应用");
  return api;
}

// 初始化 BLE：Android 上会弹「附近设备」运行时权限（SPP 也依赖这次授权）
export function bleInit(): Promise<void> {
  return bleApi().initialize({ androidNeverForLocation: true });
}

// 系统设备选择弹窗 → 连接 → 打印
export async function blePrint(payload: number[]): Promise<string> {
  const device = await bleApi().requestDevice({}); // 弹系统设备选择框
  return blePrintTo(device.deviceId, payload, device.name);
}

// 连接指定 BLE 设备并打印（用于搜索列表点击打印），返回设备名
export async function blePrintTo(deviceId: string, payload: number[], name = ""): Promise<string> {
  const api = bleApi();
  await api.connect({ deviceId });
  try {
    // GATT 服务发现常滞后于连接返回（打印机有时要几百毫秒才暴露特征），
    // 立即 getServices 会拿到空/不全的列表 → 重试几次直到找到可写特征
    let target: { service: string; characteristic: string; useResponse: boolean } | null = null;
    let svcCount = 0;
    for (let i = 0; i < 6 && !target; i++) {
      const { services } = await api.getServices({ deviceId });
      svcCount = services?.length ?? 0;
      target = findWritable(services ?? []);
      if (!target) await new Promise((r) => setTimeout(r, 300));
    }
    if (!target) {
      throw new Error(
        svcCount === 0
          ? "已连接但未发现任何蓝牙服务，可能是连错了设备或打印机正忙，请重试"
          : `已连接但未找到可写特征（发现 ${svcCount} 个服务），请确认这是 BLE 小票机`
      );
    }

    // 单次写入上限 = MTU - 3（ATT 协议头）
    let mtu = 23;
    try {
      mtu = (await api.getMtu({ deviceId })).value;
    } catch {
      // 取不到就按最小值保守分片
    }
    const chunkSize = Math.max(20, mtu - 3);

    const bytes = new Uint8Array(payload);
    for (let i = 0; i < bytes.length; i += chunkSize) {
      const opts = {
        deviceId,
        service: target.service,
        characteristic: target.characteristic,
        value: toHex(bytes.subarray(i, i + chunkSize)) // 原生插件的 value 要求 hex 字符串
      };
      if (target.useResponse) {
        await api.write(opts);
      } else {
        await api.writeWithoutResponse(opts);
      }
    }
    return name || "未知设备";
  } finally {
    api.disconnect({ deviceId }).catch(() => undefined);
  }
}

// ================= 蓝牙搜索 =================
export interface BleScanDevice {
  deviceId: string;
  name: string;
  rssi?: number;
}

// BLE 搜索：开始扫描附近广播设备（不含已保存配对关系也能发现）
// 返回停止函数；onDevice 每发现一个新设备回调一次
export async function bleScanStart(onDevice: (d: BleScanDevice) => void): Promise<() => Promise<void>> {
  const api = bleApi();
  const seen = new Set<string>();
  const sub = await api.addListener("onScanResult", (r: any) => {
    const dev = r?.device ?? {};
    const id: string = dev.deviceId ?? dev.id ?? "";
    if (!id || seen.has(id)) return;
    seen.add(id);
    onDevice({ deviceId: id, name: r.localName || dev.name || "", rssi: r.rssi });
  });
  try {
    await api.requestLEScan({ services: [], allowDuplicates: false, scanMode: 2 });
  } catch (e) {
    await sub.remove().catch(() => undefined);
    throw e;
  }
  return async () => {
    await sub.remove().catch(() => undefined);
    await api.stopLEScan().catch(() => undefined);
  };
}

// 找第一个具备写能力的特征
function findWritable(services: BleService[]) {
  for (const svc of services) {
    for (const c of svc.characteristics) {
      if (c.properties?.write) return { service: svc.uuid, characteristic: c.uuid, useResponse: true };
      if (c.properties?.writeWithoutResponse)
        return { service: svc.uuid, characteristic: c.uuid, useResponse: false };
    }
  }
  return null;
}

function toHex(bytes: Uint8Array): string {
  let hex = "";
  for (const b of bytes) hex += b.toString(16).padStart(2, "0");
  return hex;
}

// ==================== Web Bluetooth（Electron / Chrome） ====================
// 常见蓝牙小票机 GATT 服务 UUID（不同厂商差异大，尽量多列）
const PRINTER_SERVICES = [
  0xff00,
  0x18f0,
  0xffe0,
  0xfff0,
  0xffe5,
  "0000ff00-0000-1000-8000-00805f9b34fb",
  "000018f0-0000-1000-8000-00805f9b34fb",
  "0000ffe0-0000-1000-8000-00805f9b34fb",
  "0000fff0-0000-1000-8000-00805f9b34fb",
  "49535343-fe7d-4ae5-8fa9-9fafd205e455"
];

// Web Bluetooth 没有被动扫描 API，只能弹系统设备选择框（用户勾选即"搜索"）
// 已选择的设备缓存在此，供后续点击打印
const webDevices = new Map<string, any>();

// 弹出 Web Bluetooth 设备选择框，返回选中设备（同时缓存以便打印）
export async function webBluetoothPick(): Promise<{ deviceId: string; name: string }> {
  const bluetooth = (navigator as Navigator & { bluetooth?: any }).bluetooth;
  if (!bluetooth) {
    throw new Error("当前环境不支持 Web Bluetooth（需 Chrome/Edge/Chromium 内核）");
  }
  if (!window.isSecureContext) {
    throw new Error("页面不在安全上下文：改用 file:// 或 localhost 加载即可，无需 HTTPS");
  }

  const device = await bluetooth.requestDevice({
    acceptAllDevices: true,
    optionalServices: PRINTER_SERVICES
  });
  webDevices.set(device.id, device);
  return { deviceId: device.id, name: device.name || "" };
}

// 向已选择的 Web Bluetooth 设备打印
export async function webBluetoothPrintTo(deviceId: string, payload: number[]): Promise<string> {
  const device = webDevices.get(deviceId);
  if (!device) {
    // 没缓存（如刷新过页面）则重新弹选择框
    const picked = await webBluetoothPick();
    if (picked.deviceId !== deviceId) throw new Error("设备不一致，请重新选择");
    return webBluetoothPrintTo(deviceId, payload);
  }
  const server = await device.gatt.connect();
  try {
    const ch = await findWritableChar(server);
    if (!ch) throw new Error("已连接但未找到可写服务特征，请确认这是 BLE 小票机");

    const bytes = new Uint8Array(payload);
    const CHUNK = 180; // BLE 单包容量有限
    for (let i = 0; i < bytes.length; i += CHUNK) {
      const chunk = bytes.slice(i, i + CHUNK);
      if (ch.writeValueWithResponse) {
        await ch.writeValueWithResponse(chunk);
      } else {
        await ch.writeValue(chunk);
      }
    }
    return device.name || "未知设备";
  } finally {
    server.disconnect();
  }
}

export async function webBluetoothPrint(payload: number[]): Promise<string> {
  const picked = await webBluetoothPick();
  return webBluetoothPrintTo(picked.deviceId, payload);
}

// 遍历所有服务，返回第一个具备写能力的特征
async function findWritableChar(server: any) {
  const services = await server.getPrimaryServices();
  for (const svc of services) {
    const chars = await svc.getCharacteristics();
    for (const c of chars) {
      if (c.properties?.write || c.properties?.writeWithoutResponse) return c;
    }
  }
  return null;
}

// ==================== 错误文案 ====================
export function describeError(e: any): string {
  const name = e?.name;
  const msg = String(e?.message || e || "");
  if (name === "NotFoundError" && /cancel/i.test(msg)) return "已取消选择设备";
  if (name === "NotFoundError") return "未发现蓝牙设备，请确认打印机已开机并进入配对模式";
  if (name === "SecurityError") {
    return "安全限制：请确认页面以 file:// 或 localhost 加载（无需 HTTPS），且由点击按钮触发";
  }
  if (/permission/i.test(msg)) {
    return `蓝牙权限被拒绝，请在系统设置中允许本应用的「附近设备」权限：${msg}`;
  }
  return msg || "测试失败";
}
