// 商品本地模糊搜索：中文、拼音全拼、拼音首字母，并容忍漏字/错字
import { pinyin } from "pinyin-pro";

export interface SearchableGoods {
  name: string;
  alias?: string;
  spec?: string;
}

const CJK = /[\u4e00-\u9fa5]/;
// 非字母/数字/汉字一律当分隔符，避免空格和标点影响匹配
const NON_WORD = /[^a-z0-9\u4e00-\u9fa5]+/g;

// 文本检索变体缓存：商品数据稳定，避免每次输入都重复转拼音
const variantCache = new Map<string, string[]>();

function norm(s: string): string {
  return s.toLowerCase().replace(NON_WORD, "");
}

// 一个文本的检索变体：原文、拼音全拼、拼音首字母
function variants(text: string): string[] {
  const cached = variantCache.get(text);
  if (cached) return cached;

  const raw = norm(text);
  const out = new Set<string>();
  if (raw) {
    out.add(raw);
    if (CJK.test(raw)) {
      out.add(norm(pinyin(raw, { toneType: "none" })));
      out.add(norm(pinyin(raw, { pattern: "first", toneType: "none" })));
    }
  }
  const list = [...out];
  variantCache.set(text, list);
  return list;
}

// q 是否为 t 的子序列（容忍输入时漏字）
function isSubsequence(q: string, t: string): boolean {
  let i = 0;
  for (const ch of t) {
    if (ch === q[i]) i += 1;
    if (i === q.length) return true;
  }
  return false;
}

// 标准编辑距离（两行滚动数组）
function levenshtein(a: string, b: string): number {
  if (a === b) return 0;
  if (a.length === 0) return b.length;
  if (b.length === 0) return a.length;

  let prev = Array.from({ length: b.length + 1 }, (_, j) => j);
  for (let i = 1; i <= a.length; i += 1) {
    const cur = [i];
    for (let j = 1; j <= b.length; j += 1) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1;
      const del = (prev[j] ?? 0) + 1;
      const ins = (cur[j - 1] ?? 0) + 1;
      const sub = (prev[j - 1] ?? 0) + cost;
      cur[j] = Math.min(del, ins, sub);
    }
    prev = cur;
  }
  return prev[b.length] ?? a.length;
}

// 查询串与文本任意等长（±1）子串的最小编辑距离，容忍错字/多字/少字
function minWindowDistance(q: string, t: string): number {
  let best = Infinity;
  for (const len of [q.length - 1, q.length, q.length + 1]) {
    if (len <= 0 || len > t.length) continue;
    for (let i = 0; i + len <= t.length; i += 1) {
      const d = levenshtein(q, t.slice(i, i + len));
      if (d < best) best = d;
      if (best === 0) return 0;
    }
  }
  return best;
}

// 单项打分：完全相等 > 前缀 > 包含 > 子序列（漏字）> 编辑距离（错字）
function scoreOne(q: string, t: string): number {
  if (!q || !t) return 0;
  if (t === q) return 100;
  if (t.startsWith(q)) return 90;
  if (t.includes(q)) return 80;
  if (q.length >= 2 && isSubsequence(q, t)) return 60;
  if (q.length >= 2 && minWindowDistance(q, t) <= (q.length <= 4 ? 1 : 2)) return 40;
  return 0;
}

// 字段得分：取「查询变体 × 文本变体」的最高分
function fieldScore(qs: string[], text: string): number {
  let best = 0;
  const ts = variants(text);
  for (const q of qs) {
    for (const t of ts) {
      const s = scoreOne(q, t);
      if (s > best) best = s;
    }
  }
  return best;
}

/**
 * 本地模糊搜索：返回按匹配度降序过滤后的商品（原顺序稳定）
 * 匹配范围：名称、别名（逗号分隔的每个别名）、规格
 */
export function searchGoods<T extends SearchableGoods>(list: T[], query: string): T[] {
  const q = query.trim();
  if (!q) return list;

  const qs = variants(q);
  if (qs.length === 0) return [];

  const scored: { item: T; score: number }[] = [];
  for (const item of list) {
    const fields: { text: string; weight: number }[] = [{ text: item.name, weight: 1 }];
    if (item.alias) {
      for (const a of item.alias.split(/[,，、\s]+/)) {
        if (a) fields.push({ text: a, weight: 0.95 });
      }
    }
    if (item.spec) fields.push({ text: item.spec, weight: 0.9 });

    let best = 0;
    for (const f of fields) {
      const s = fieldScore(qs, f.text) * f.weight;
      if (s > best) best = s;
    }
    if (best > 0) scored.push({ item, score: best });
  }

  scored.sort((a, b) => b.score - a.score);
  return scored.map((s) => s.item);
}
