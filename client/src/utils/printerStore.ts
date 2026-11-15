// 打印机设备管理：已保存设备列表，localStorage 持久化
// 打印时直接从这里选设备，不再每次弹系统蓝牙选择框
// 不去重：同一地址可重复添加（不同名称/编码/用途各存一条），key 每次添加唯一生成
import { ref, watch } from "vue";
import type { PrinterConn, PrinterEncoding } from "./printer";

export type PrinterKind = "receipt" | "label"; // 设备类型：小票机 / 标签打印机（标签打印占位）
export type PrinterTransport = "bt" | "wifi"; // 连接方式：蓝牙 / WiFi（WiFi 打印占位）

export interface SavedPrinter {
  key: string; // 唯一标识（每次添加生成，编辑/删除按 key 操作）
  id: string; // 地址：BLE deviceId / SPP MAC / WiFi IP:端口
  name: string; // 自定义名称（可改名）
  conn: PrinterConn; // 蓝牙子类型：ble 低功耗 / spp 传统（transport=bt 时有效）
  transport: PrinterTransport; // 连接方式：bt 蓝牙 / wifi WiFi（占位）
  kind: PrinterKind; // 设备类型：receipt 小票机 / label 标签打印机（占位）
  encoding: PrinterEncoding; // 编码格式
}

const KEY = "lfdd_printers";

function genKey() {
  return `${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`;
}

// 读取存量数据并补全新字段（旧数据无 key / transport / kind）
function load(): SavedPrinter[] {
  try {
    const raw = localStorage.getItem(KEY);
    const arr = raw ? (JSON.parse(raw) as SavedPrinter[]) : [];
    if (!Array.isArray(arr)) return [];
    return arr.filter((p): p is SavedPrinter => !!p && !!p.id).map((p) => ({
      ...p,
      key: p.key || genKey(),
      transport: p.transport || "bt",
      kind: p.kind || "receipt"
    }));
  } catch {
    return [];
  }
}

const printers = ref<SavedPrinter[]>(load());
watch(printers, (v) => localStorage.setItem(KEY, JSON.stringify(v)), { deep: true });

// 新增设备（不去重：每次都生成新 key 追加）
export function savePrinter(p: Omit<SavedPrinter, "key">) {
  printers.value.push({ ...p, key: genKey() });
}

// 修改设备（按 key）
export function updatePrinter(key: string, patch: Partial<Omit<SavedPrinter, "key">>) {
  const p = printers.value.find((x) => x.key === key);
  if (p) Object.assign(p, patch);
}

export function removePrinter(key: string) {
  printers.value = printers.value.filter((p) => p.key !== key);
}

export { printers };
