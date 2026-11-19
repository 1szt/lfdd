import { ref } from "vue";
import { changeQty } from "./cart";

// AI 点单共享状态（模块级：首页与购物车页共用同一份输入内容，切换路由不丢失）
export const aiText = ref("");
export const aiLoading = ref(false);

export interface AiMatchedGoods {
  id: string;
  name: string;
  price: number;
  spec: string;
  stock: number;
  image: string;
  menu_id?: string;
}

// 自然语言描述 → 后端大模型匹配商品 → 逐个加入购物车
// 返回加入的商品种类数（无匹配/输入为空时返回 0）
export async function aiOrder(): Promise<number> {
  const text = aiText.value.trim();
  if (!text || aiLoading.value) return 0;
  aiLoading.value = true;
  try {
    const res = await fetch("/api/goods/match", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ text })
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.message || `匹配失败（${res.status}）`);

    const list: (AiMatchedGoods & { qty: number })[] = Array.isArray(data.list)
      ? data.list
      : [];
    // 按返回的数量逐个加入（changeQty 内部会按库存封顶）
    for (const item of list) {
      const { qty, ...goods } = item;
      for (let n = 0; n < qty; n++) {
        changeQty(goods, 1);
      }
    }
    if (list.length > 0) aiText.value = "";
    return list.length;
  } finally {
    aiLoading.value = false;
  }
}
