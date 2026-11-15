<template>
  <q-page class="goods-page" :style-fn="pageStyle">
    <div class="row no-wrap goods-body">
      <!-- 左侧分类栏（蜜雪冰城/肯德基式窄边栏） -->
      <div ref="sideRef" class="goods-sidebar">
        <div
          class="sidebar-item"
          :class="{ active: activeMenuId === '0' }"
          @click="selectMenu('0')"
        >
          全部
          <q-badge v-if="cartCount > 0" floating color="red" :label="cartCount" />
        </div>
        <div
          v-for="m in menus"
          :key="m.id"
          class="sidebar-item"
          :class="{ active: activeMenuId === m.id }"
          @click="selectMenu(m.id)"
        >
          {{ m.name }}
          <q-badge v-if="menuQty(m.id) > 0" floating color="red" :label="menuQty(m.id)" />
        </div>
        <div v-if="menus.length === 0" class="text-caption text-grey-6 q-pa-sm text-center">
          暂无分类
        </div>
      </div>

      <!-- 右侧商品列表 -->
      <div ref="mainRef" class="goods-main">
        <!-- 吸顶：本地模糊搜索（名称/别名/规格，支持拼音与容错）+ AI 点单 -->
        <div class="goods-topbar">
          <q-input
            v-model="searchText"
            outlined
            dense
            clearable
            class="search-input"
            placeholder="搜索商品：名称/别名/拼音，如 nmh"
            :loading="searchLoading"
          >
            <template #prepend>
              <q-icon name="search" color="grey-7" />
            </template>
          </q-input>

          <!-- AI 点单：自然语言描述 → 大模型匹配商品 → 自动加入购物车 -->
          <div class="ai-bar row items-center q-gutter-sm">
            <q-input
              v-model="aiText"
              outlined
              dense
              class="col"
              placeholder="AI 点单：如「两杯柠檬水和一个冰淇淋」"
              :disable="aiLoading"
              @keyup.enter="aiMatch"
            >
              <template #prepend>
                <q-icon name="auto_awesome" color="primary" />
              </template>
            </q-input>
            <q-btn
              color="primary"
              label="点单"
              unelevated
              :loading="aiLoading"
              @click="aiMatch"
            />
          </div>
        </div>

        <div v-if="searchMode" class="search-hint text-caption text-grey-7">
          搜索「{{ searchText?.trim() }}」
          {{ searchLoading ? "加载中…" : `找到 ${displayGoods.length} 个商品` }}
        </div>

        <div v-if="errorMsg" class="error-bar text-negative text-caption">
          {{ errorMsg }}
        </div>

        <div v-if="loadingGoods || searchLoading" class="state-box text-grey-6">加载中…</div>
        <div v-else-if="displayGoods.length === 0 && !errorMsg" class="state-box text-grey-6">
          {{ searchMode ? "没有匹配的商品，换个关键词试试" : "暂无商品" }}
        </div>

        <template v-else>
          <div v-for="g in displayGoods" :key="g.id" class="goods-card">
            <div class="goods-img">
              <img v-if="g.image" :src="imgSrc(g.image)" :alt="g.name" />
              <q-icon v-else name="shopping_bag" size="32px" color="grey-5" />
              <div v-if="g.stock === 0" class="sold-out">售罄</div>
            </div>
            <div class="col">
              <div class="goods-name text-weight-medium">{{ g.name }}</div>
              <div v-if="g.spec" class="text-caption text-grey-7 ellipsis">{{ g.spec }}</div>
              <div class="row items-center justify-between q-mt-xs">
                <div class="row items-baseline q-gutter-sm">
                  <div class="text-negative text-weight-bold">¥{{ fmtPrice(g.price) }}</div>
                  <span class="text-caption text-grey-6">
                    {{ g.stock === 0 ? '已售罄' : g.stock === -1 ? '' : `库存 ${g.stock}` }}
                  </span>
                </div>
                <!-- 加减按钮（大号）：数量即时算进购物车 -->
                <div v-if="g.stock !== 0" class="qty-ctrl">
                  <q-btn
                    v-if="lineQty(g.id) > 0"
                    round
                    dense
                    color="primary"
                    class="sub-btn"
                    aria-label="减少"
                    @click="changeQty(g, -1)"
                  >
                    <q-icon name="remove" size="18px" />
                  </q-btn>
                  <q-input
                    v-if="lineQty(g.id) > 0"
                    :model-value="lineQty(g.id)"
                    type="number"
                    dense
                    borderless
                    class="qty-input"
                    aria-label="修改数量"
                    @change="onQtyInput(g, $event)"
                  />
                  <q-btn
                    round
                    dense
                    color="primary"
                    class="add-btn"
                    aria-label="增加"
                    @click="changeQty(g, 1)"
                  >
                    <q-icon name="add" size="22px" />
                  </q-btn>
                </div>
              </div>
            </div>
          </div>

          <div class="text-center q-pb-md">
            <q-btn
              v-if="!searchMode && hasMore"
              flat
              dense
              color="primary"
              :loading="loadingMore"
              label="加载更多"
              @click="loadGoods(false)"
            />
            <span v-else-if="!searchMode" class="text-caption text-grey-6">— 没有更多了 —</span>
          </div>
        </template>
      </div>
    </div>

    <!-- 底部悬浮购物车入口：合计实时计算，点击进购物车页看明细/图片 -->
    <div v-if="cartCount > 0" class="cart-dock">
      <div class="cart-bar row items-center" @click="goCart">
        <div class="relative-position">
          <q-icon name="shopping_cart" size="26px" />
          <q-badge floating color="red" :label="cartCount" />
        </div>
        <div class="col q-ml-sm">
          <div class="text-caption" style="line-height: 1">合计</div>
          <div class="text-weight-bold" style="line-height: 1.2">¥{{ fmtPrice(cartTotal) }}</div>
        </div>
        <q-space />
        <q-icon name="chevron_right" size="22px" />
      </div>
    </div>
  </q-page>
</template>

<script lang="ts">
// 模块级：跨路由切换保留的商品页位置（组件卸载后仍然存在）
// 切到购物车/订单/后台再切回来时，恢复分类和两栏滚动位置
const pagePos = {
  activeMenuId: "0",
  side: 0,
  main: 0
};
</script>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { searchGoods } from "../../utils/search";
import {
  cartCount,
  cartTotal,
  changeQty,
  fmtPrice,
  imgSrc,
  lineQty,
  menuQty,
  setQty
} from "../../utils/cart";
// 购物车数据在 utils/cart.ts（模块级共享，购物车页同用一份）

// 与后端 service 对齐的接口返回结构
interface SidebarMenu {
  id: string; // `,string` tag，如 "101"
  name: string;
}

interface GoodsItem {
  id: string;
  menu_id: string;
  name: string;
  alias?: string; // 别名（多个用逗号分隔），本地搜索时与名称一起匹配
  price: number; // 单位：分（1850 = 18.50 元）
  spec: string;
  stock: number; // -1 = 不限库存，0 = 售罄
  image: string;
  status: number;
}

const PAGE_SIZE = 20;

const menus = ref<SidebarMenu[]>([]);
const activeMenuId = ref(pagePos.activeMenuId); // 恢复上次选中的分类
const goods = ref<GoodsItem[]>([]);
const total = ref(0);
const page = ref(1);
const loadingGoods = ref(false);
const loadingMore = ref(false);
const errorMsg = ref("");

const sideRef = ref<HTMLElement | null>(null);
const mainRef = ref<HTMLElement | null>(null);

let reqSeq = 0; // 递增序号，丢弃切换分类时已过期的响应

const hasMore = computed(() => goods.value.length < total.value);

// ================= 本地模糊搜索（一次性拉全量上架商品，输入即时过滤） =================
// q-input 的 clearable 清空时会把值置为 null，这里兼容 null
const searchText = ref<string | null>("");
const searchLoading = ref(false);
const allGoods = ref<GoodsItem[]>([]);
const allLoaded = ref(false);

const searchMode = computed(() => (searchText.value ?? "").trim() !== "");
const displayGoods = computed(() =>
  searchMode.value ? searchGoods(allGoods.value, searchText.value ?? "") : goods.value
);

// 全量上架商品只拉一次；搜索时自动切到「全部」分类，实现跨分类检索
async function loadAllGoods() {
  if (allLoaded.value || searchLoading.value) return;
  searchLoading.value = true;
  try {
    const data = await fetchJson("/api/goods/all");
    allGoods.value = Array.isArray(data.list) ? data.list : [];
    allLoaded.value = true;
  } catch (e) {
    $q.notify({ message: `搜索数据加载失败：${describeError(e)}`, color: "negative" });
  } finally {
    searchLoading.value = false;
  }
}

// 输入关键词进入搜索态；清空后恢复当前分类的分页列表
watch(searchText, async (v) => {
  if ((v ?? "").trim()) {
    if (activeMenuId.value !== "0") {
      activeMenuId.value = "0";
      pagePos.activeMenuId = "0";
    }
    if (mainRef.value) mainRef.value.scrollTop = 0;
    await loadAllGoods();
  } else {
    await loadGoods(true);
  }
});

// QPage styleFn：header/footer 占位后剩余高度，让页面撑满且两栏各自滚动
function pageStyle(offset: number, height: number) {
  if (height > 0) {
    return { height: `${height - offset}px` };
  }
  return { height: offset !== 0 ? `calc(100vh - ${offset}px)` : "100vh" };
}

function describeError(e: unknown) {
  return e instanceof Error ? e.message : String(e);
}

const $q = useQuasar();
const router = useRouter();

function goCart() {
  void router.push("/cart");
}

// 数量输入框直接改数字：0/非法 → 移除该商品，超库存由 setQty 封顶
function onQtyInput(g: GoodsItem, v: string | number) {
  const n = Math.floor(Number(v));
  setQty(g, Number.isFinite(n) ? n : 0);
}

// ================= AI 点单（后端大模型匹配商品） =================
const aiText = ref("");
const aiLoading = ref(false);

interface AiMatchedGoods {
  id: string;
  name: string;
  price: number;
  spec: string;
  stock: number;
  image: string;
  qty: number;
}

async function aiMatch() {
  const text = aiText.value.trim();
  if (!text || aiLoading.value) return;
  aiLoading.value = true;
  try {
    const res = await fetch("/api/goods/match", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ text })
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.message || `匹配失败（${res.status}）`);

    const list: AiMatchedGoods[] = Array.isArray(data.list) ? data.list : [];
    if (list.length === 0) {
      $q.notify({ message: "没有匹配到商品，换个说法试试", color: "orange" });
      return;
    }
    // 按返回的数量逐个加入（changeQty 内部会按库存封顶）
    for (const item of list) {
      const { qty, ...goods } = item;
      for (let n = 0; n < qty; n++) {
        changeQty(goods, 1);
      }
    }
    aiText.value = "";
    $q.notify({
      message: `已加入 ${list.length} 种商品`,
      color: "positive"
    });
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    aiLoading.value = false;
  }
}

async function fetchJson(url: string) {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`请求失败（${res.status}）`);
  }
  return res.json();
}

// 左侧分类列表
async function loadMenus() {
  try {
    const data = await fetchJson("/api/goods/sidebar");
    menus.value = Array.isArray(data.menus) ? data.menus : [];
  } catch (e) {
    menus.value = [];
    // 分类失败不阻断商品加载（menu_id=0 仍可看全部）
    console.error("加载侧边分类失败:", e);
  }
}

// 商品列表（reset=true 重新第一页，false 加载更多）
async function loadGoods(reset: boolean) {
  const seq = ++reqSeq;
  if (reset) {
    page.value = 1;
    errorMsg.value = "";
    loadingGoods.value = true;
  } else {
    loadingMore.value = true;
  }

  try {
    const data = await fetchJson(
      `/api/goods/page?menu_id=${activeMenuId.value}&page=${page.value}&page_size=${PAGE_SIZE}`
    );
    if (seq !== reqSeq) return; // 已有更新的请求，丢弃本次结果

    const list: GoodsItem[] = Array.isArray(data.list) ? data.list : [];
    goods.value = reset ? list : [...goods.value, ...list];
    total.value = typeof data.total === "number" ? data.total : 0;
    if (list.length > 0) {
      page.value += 1;
    }
  } catch (e) {
    if (seq !== reqSeq) return;
    if (reset) {
      goods.value = [];
      total.value = 0;
    }
    errorMsg.value = `加载失败：${describeError(e)}，请确认后端服务已启动`;
  } finally {
    if (seq === reqSeq) {
      loadingGoods.value = false;
      loadingMore.value = false;
    }
  }
}

function selectMenu(id: string) {
  activeMenuId.value = id;
  pagePos.activeMenuId = id;
  if (searchText.value) {
    searchText.value = ""; // 退出搜索态，watch 会按新分类重新加载
  } else {
    void loadGoods(true);
  }
  // 切分类后商品区回到顶部
  if (mainRef.value) mainRef.value.scrollTop = 0;
}

onMounted(async () => {
  await Promise.all([loadMenus(), loadGoods(true)]);
  // 数据渲染后恢复上次的滚动位置
  await nextTick();
  if (sideRef.value) sideRef.value.scrollTop = pagePos.side;
  if (mainRef.value) mainRef.value.scrollTop = pagePos.main;
});

onBeforeUnmount(() => {
  pagePos.activeMenuId = activeMenuId.value;
  pagePos.side = sideRef.value?.scrollTop ?? 0;
  pagePos.main = mainRef.value?.scrollTop ?? 0;
});
</script>

<style scoped lang="scss">
.goods-page {
  overflow: hidden;
}

.goods-body {
  height: 100%;
}

/* 左侧分类栏 */
.goods-sidebar {
  width: 88px;
  flex-shrink: 0;
  overflow-y: auto;
  background: #f5f5f5;
  /* 顶部/右侧留空：避免第一项「全部」的悬浮角标被 overflow 裁切 */
  padding-top: 8px;
  padding-right: 6px;
}

.sidebar-item {
  position: relative;
  padding: 15px 6px;
  text-align: center;
  font-size: 13px;
  line-height: 1.3;
  color: #666;
  cursor: pointer;
  word-break: break-all;

  &.active {
    background: #fff;
    color: $primary;
    font-weight: 600;

    &::before {
      content: "";
      position: absolute;
      left: 0;
      top: 20%;
      bottom: 20%;
      width: 3px;
      border-radius: 0 3px 3px 0;
      background: $primary;
    }
  }
}

/* 右侧商品区 */
.goods-main {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  padding: 8px 12px;
}

.error-bar {
  padding: 8px 0 4px;
}

.state-box {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
}

.goods-card {
  display: flex;
  gap: 10px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;

  &:last-of-type {
    border-bottom: none;
  }
}

.goods-img {
  position: relative;
  width: 72px;
  height: 72px;
  flex-shrink: 0;
  border-radius: 8px;
  background: #fafafa;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.sold-out {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  font-size: 13px;
}

.goods-name {
  font-size: 14px;
}

/* 吸顶顶部栏：模糊搜索 + AI 点单 */
.goods-topbar {
  position: sticky;
  top: 0;
  z-index: 6;
  /* 抵消 .goods-main 内边距，吸顶时铺满整宽并盖住滚动内容 */
  margin: -8px -12px 0;
  padding: 8px 12px 4px;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
}

.search-input {
  margin-bottom: 8px;
}

/* AI 点单输入条 */
.ai-bar {
  margin-bottom: 8px;
}

.search-hint {
  padding: 6px 0 2px;
}

/* 加减按钮组（大号，触屏好点） */
.qty-ctrl {
  display: flex;
  align-items: center;
  gap: 8px;

  .add-btn {
    width: 36px;
    height: 36px;
  }

  .sub-btn {
    width: 32px;
    height: 32px;
  }
}

/* 数量输入框：直接改数字（0 = 移除），隐藏数字框 spinner */
.qty-input {
  width: 48px;

  :deep(input) {
    text-align: center;
    font-size: 15px;
    font-weight: 700;
    color: var(--q-primary);
    padding: 0;
    -moz-appearance: textfield;
  }

  :deep(input::-webkit-outer-spin-button),
  :deep(input::-webkit-inner-spin-button) {
    -webkit-appearance: none;
    margin: 0;
  }
}

/* 底部悬浮购物车入口（明细在 /cart 页面看） */
.cart-dock {
  position: fixed;
  right: 16px;
  bottom: 72px; /* 让开底部导航栏 */
  z-index: 1000;
}

.cart-bar {
  position: relative;
  min-width: 180px;
  padding: 10px 16px;
  border-radius: 28px;
  background: $primary;
  color: #fff;
  cursor: pointer;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
  user-select: none;
}
</style>
