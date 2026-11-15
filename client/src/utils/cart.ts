import { computed, ref, watch } from "vue";

import { apiPath } from "./api";

// 购物车共享状态（模块级：商品页加购 → 购物车页查看，跨路由保留）
export interface CartLine {
  id: string;
  name: string;
  price: number; // 单位：分
  stock: number; // -1 不限，0 售罄
  image: string; // 哈希或完整 URL
  spec: string;
  menu_id?: string; // 所属分类（侧边栏角标统计用；旧本地数据可能没有）
  qty: number;
}

// 本地持久化：关闭应用/刷新后购物车依旧保留（localStorage）
const CART_KEY = "lfdd_cart";

function loadCart(): CartLine[] {
  try {
    const raw = localStorage.getItem(CART_KEY);
    const v: unknown = raw ? JSON.parse(raw) : [];
    if (!Array.isArray(v)) return [];
    // 丢弃结构损坏的行（localStorage 属外部输入）
    return (v as CartLine[]).filter(
      l =>
        l && typeof l.id === "string" && typeof l.qty === "number" && l.qty > 0
    );
  } catch {
    return [];
  }
}

export const cart = ref<CartLine[]>(loadCart());

// 任何变动即时写入本地（增删、改数量、清空都走 ref 变更）
watch(
  cart,
  v => {
    try {
      localStorage.setItem(CART_KEY, JSON.stringify(v));
    } catch {
      // 存储被禁用/超限时忽略，不影响功能
    }
  },
  { deep: true }
);

export const cartCount = computed(() =>
  cart.value.reduce((s, l) => s + l.qty, 0)
);
export const cartTotal = computed(() =>
  cart.value.reduce((s, l) => s + l.price * l.qty, 0)
);

export function lineQty(id: string): number {
  return cart.value.find(l => l.id === id)?.qty ?? 0;
}

// 某分类在购物车里的件数（侧边栏右上角角标）
export function menuQty(menuId: string): number {
  return cart.value.reduce((s, l) => (l.menu_id === menuId ? s + l.qty : s), 0);
}

// 增减数量：delta=1 加，-1 减；受库存限制（stock=-1 不限），减到 0 自动移出行
export function changeQty(
  item: {
    id: string;
    name: string;
    price: number;
    stock: number;
    image?: string;
    spec?: string;
    menu_id?: string;
  },
  delta: number
) {
  const idx = cart.value.findIndex(l => l.id === item.id);
  if (delta > 0) {
    if (idx >= 0) {
      const line = cart.value[idx]!;
      if (item.stock > 0 && line.qty >= item.stock) return; // 超库存忽略
      line.qty += 1;
    } else {
      cart.value.push({
        ...item,
        image: item.image ?? "",
        spec: item.spec ?? "",
        menu_id: item.menu_id ?? "",
        qty: 1
      });
    }
  } else if (idx >= 0) {
    const line = cart.value[idx]!;
    line.qty -= 1;
    if (line.qty <= 0) cart.value.splice(idx, 1);
  }
}

export function clearCart() {
  cart.value = [];
}

// 一键删除整行（数量行的垃圾桶按钮）
export function removeLine(id: string) {
  const idx = cart.value.findIndex(l => l.id === id);
  if (idx >= 0) cart.value.splice(idx, 1);
}

// 复制订单明细进购物车（与已有行合并数量），用于重新提交订单。
// 订单里没有库存信息，stock 取 -1（不限）——下单时后端仍会校验库存。
export function copyOrderToCart(
  items: {
    goods_id: string;
    menu_id?: string;
    name: string;
    price: number;
    spec: string;
    image: string;
    qty: number;
  }[]
) {
  for (const it of items) {
    if (!it.goods_id || !(it.qty > 0)) continue;
    const idx = cart.value.findIndex(l => l.id === it.goods_id);
    if (idx >= 0) {
      cart.value[idx]!.qty += it.qty;
      // 合并时补上分类（旧行可能没有）
      if (!cart.value[idx]!.menu_id && it.menu_id)
        cart.value[idx]!.menu_id = it.menu_id;
    } else {
      cart.value.push({
        id: it.goods_id,
        name: it.name,
        price: it.price,
        stock: -1,
        image: it.image ?? "",
        spec: it.spec ?? "",
        menu_id: it.menu_id ?? "",
        qty: it.qty
      });
    }
  }
}

// 直接设置数量（数字输入框用）：0/非法 → 移除行，超出库存封顶到库存（stock=-1 不限）
export function setQty(
  item: {
    id: string;
    name: string;
    price: number;
    stock: number;
    image?: string;
    spec?: string;
    menu_id?: string;
  },
  qty: number
) {
  const idx = cart.value.findIndex(l => l.id === item.id);
  const n = Math.floor(qty);
  if (!Number.isFinite(n) || n <= 0) {
    if (idx >= 0) cart.value.splice(idx, 1);
    return;
  }
  const capped = item.stock > 0 && n > item.stock ? item.stock : n;
  if (idx >= 0) {
    cart.value[idx]!.qty = capped;
  } else {
    cart.value.push({
      ...item,
      image: item.image ?? "",
      spec: item.spec ?? "",
      menu_id: item.menu_id ?? "",
      qty: capped
    });
  }
}

export function fmtPrice(price: number): string {
  return (price / 100).toFixed(2);
}

// image 字段只存哈希，展示时拼成 /api/upload/{hash}（兼容旧数据里的完整 URL/路径）
export function imgSrc(v: string): string {
  if (!v) return "";
  if (/^https?:\/\//.test(v)) return v;
  if (v.startsWith("/")) return apiPath(v);
  return apiPath(`/api/upload/${v}`);
}
