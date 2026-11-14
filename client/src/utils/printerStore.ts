// 打印机设备管理：已保存设备列表（名称 / 连接方式 / 编码格式），localStorage 持久化
// 打印时直接从这里选设备，不再每次弹系统蓝牙选择框
import { ref, watch } from "vue";
import type { PrinterConn, PrinterEncoding } from "./printer";

export interface SavedPrinter {
  id: string; // BLE: deviceId；SPP: 蓝牙 MAC 地址
  name: string; // 自定义名称（可改名）
  conn: PrinterConn; // 连接方式
  encoding: PrinterEncoding; // 编码格式
}

const KEY = "lfdd_printers";

function load(): SavedPrinter[] {
  try {
    const raw = localStorage.getItem(KEY);
    const arr = raw ? (JSON.parse(raw) as SavedPrinter[]) : [];
    return Array.isArray(arr) ? arr.filter((p) => p && p.id) : [];
  } catch {
    return [];
  }
}

const printers = ref<SavedPrinter[]>(load());
watch(printers, (v) => localStorage.setItem(KEY, JSON.stringify(v)), { deep: true });

// 新增/更新设备（同 id 覆盖，重复扫描同一设备不产生重复项）
export function savePrinter(p: SavedPrinter) {
  const i = printers.value.findIndex((x) => x.id === p.id);
  if (i >= 0) printers.value[i] = { ...p };
  else printers.value.push({ ...p });
}

// 修改名称 / 编码格式
export function updatePrinter(id: string, patch: Partial<Omit<SavedPrinter, "id">>) {
  const p = printers.value.find((x) => x.id === id);
  if (p) Object.assign(p, patch);
}

export function removePrinter(id: string) {
  printers.value = printers.value.filter((p) => p.id !== id);
}

export { printers };
