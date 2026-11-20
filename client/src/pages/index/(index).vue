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
          <q-badge
            v-if="cartCount > 0"
            floating
            color="red"
            :label="cartCount"
          />
        </div>
        <div
          v-for="m in menus"
          :key="m.id"
          class="sidebar-item"
          :class="{ active: activeMenuId === m.id }"
          @click="selectMenu(m.id)"
        >
          {{ m.name }}
          <q-badge
            v-if="menuQty(m.id) > 0"
            floating
            color="red"
            :label="menuQty(m.id)"
          />
        </div>
        <div
          v-if="menus.length === 0"
          class="text-caption text-grey-6 q-pa-sm text-center"
        >
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

        <div v-if="loadingGoods || searchLoading" class="state-box text-grey-6"
          >加载中…</div
        >
        <div
          v-else-if="displayGoods.length === 0 && !errorMsg"
          class="state-box text-grey-6"
        >
          {{ searchMode ? "没有匹配的商品，换个关键词试试" : "暂无商品" }}
        </div>

        <template v-else>
          <div v-for="(g, idx) in displayGoods" :key="g.id" class="goods-card">
            <div class="goods-img">
              <img
                v-if="g.image"
                :src="imgSrc(g.image)"
                :alt="g.name"
                class="goods-img-clickable"
                @click="openViewer(g)"
              />
              <q-icon v-else name="shopping_bag" size="32px" color="grey-5" />
              <div v-if="g.stock === 0" class="sold-out">售罄</div>
            </div>
            <div class="col">
              <div class="row items-start no-wrap goods-name-row">
                <div class="goods-name text-weight-medium col">{{ g.name }}</div>
                <!-- 管理员：第一行小眼睛开启进价，点进价关闭并复现小眼睛 -->
                <div
                  v-if="isAdmin"
                  class="goods-cost-toggle"
                  :class="{ 'cursor-pointer': idx === 0 }"
                  @click="idx === 0 && (showCost = !showCost)"
                >
                  <q-icon
                    v-if="idx === 0 && !showCost"
                    name="visibility"
                    size="16px"
                    class="cost-eye"
                  />
                  <span
                    v-else-if="showCost"
                    class="goods-cost text-caption"
                    >进 ¥{{ fmtPrice(g.cost ?? 0) }}</span
                  >
                </div>
              </div>
              <div v-if="g.spec" class="text-caption text-grey-7 ellipsis">{{
                g.spec
              }}</div>
              <div class="goods-foot row items-center justify-between q-mt-xs">
                <div class="goods-price-row row items-baseline q-gutter-sm">
                  <div class="text-negative text-weight-bold"
                    >¥{{ fmtPrice(g.price) }}</div
                  >
                  <span class="text-caption text-grey-6">
                    {{
                      g.stock === 0
                        ? "已售罄"
                        : g.stock === -1
                          ? ""
                          : `库存 ${g.stock}`
                    }}
                  </span>
                </div>
                <!-- 加减按钮（大号）：数量即时算进购物车 -->
                <div v-if="g.stock !== 0" class="qty-ctrl">
                  <q-btn
                    round
                    dense
                    color="primary"
                    class="sub-btn"
                    :class="{ 'qty-placeholder': lineQty(g.id) <= 0 }"
                    :tabindex="lineQty(g.id) > 0 ? 0 : -1"
                    aria-label="减少"
                    @click="changeQty(g, -1)"
                  >
                    <q-icon name="remove" size="18px" />
                  </q-btn>
                  <q-input
                    :model-value="lineQty(g.id)"
                    type="number"
                    dense
                    borderless
                    class="qty-input"
                    :class="{ 'qty-placeholder': lineQty(g.id) <= 0 }"
                    :tabindex="lineQty(g.id) > 0 ? 0 : -1"
                    :aria-hidden="lineQty(g.id) <= 0"
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
            <span v-else-if="!searchMode" class="text-caption text-grey-6"
              >— 没有更多了 —</span
            >
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
          <div class="text-weight-bold" style="line-height: 1.2"
            >¥{{ fmtPrice(cartTotal) }}</div
          >
        </div>
        <q-space />
        <q-icon name="chevron_right" size="22px" />
      </div>
    </div>
    <!-- 商品图片/视频查看：自制全屏预览，支持左右滑动切换 -->
    <div v-if="viewerOpen" class="viewer-wrap">
      <q-btn
        round
        dense
        flat
        icon="close"
        color="white"
        class="viewer-close"
        aria-label="关闭"
        @click="requestCloseViewer"
      />
      <div
        ref="viewerScroller"
        class="viewer-scroller"
        @scroll.passive="onViewerScroll"
        @pointerdown="onViewerPointerDown"
        @pointermove="onViewerPointerMove"
        @pointerup="onViewerPointerUp"
        @pointerleave="onViewerPointerUp"
      >
        <div v-for="m in viewerMedia" :key="m.hash" class="viewer-slide">
          <img
            v-if="m.type === 'image'"
            :src="imgSrc(m.hash)"
            alt=""
            class="viewer-media"
            draggable="false"
          />
          <video
            v-else
            :src="imgSrc(m.hash)"
            controls
            playsinline
            preload="metadata"
            class="viewer-media"
          />
        </div>
      </div>
      <template v-if="viewerMedia.length > 1">
        <q-btn
          round
          dense
          flat
          icon="chevron_left"
          color="white"
          class="viewer-arrow viewer-arrow-left"
          aria-label="上一张"
          @click="goToMedia(viewerIndex - 1)"
        />
        <q-btn
          round
          dense
          flat
          icon="chevron_right"
          color="white"
          class="viewer-arrow viewer-arrow-right"
          aria-label="下一张"
          @click="goToMedia(viewerIndex + 1)"
        />
        <div class="viewer-dots">
          <span
            v-for="(m, i) in viewerMedia"
            :key="m.hash"
            class="viewer-dot"
            :class="{ active: i === viewerIndex }"
            @click="goToMedia(i)"
          />
        </div>
      </template>
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
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch
} from "vue";
import { useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { searchGoods } from "../../utils/search";
import { hasToken } from "../../utils/guest";
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
import { aiLoading, aiOrder, aiText } from "../../utils/aiOrder";
// 购物车数据在 utils/cart.ts（模块级共享，购物车页同用一份）

// 与后端 service 对齐的接口返回结构
interface SidebarMenu {
  id: string; // `,string` tag，如 "101"
  name: string;
}

interface GoodsMedia {
  hash: string;
  type: "image" | "video";
}

interface GoodsItem {
  id: string;
  menu_id: string;
  name: string;
  alias?: string; // 别名（多个用逗号分隔），本地搜索时与名称一起匹配
  price: number; // 单位：分（1850 = 18.50 元）
  cost?: number; // 进价，单位：分（仅管理员接口返回）
  spec: string;
  stock: number; // -1 = 不限库存，0 = 售罄
  image: string; // 第一张图片哈希（首页展示）
  media?: GoodsMedia[]; // 图片/视频列表
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
  searchMode.value
    ? searchGoods(allGoods.value, searchText.value ?? "")
    : goods.value
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
    $q.notify({
      message: `搜索数据加载失败：${describeError(e)}`,
      color: "negative"
    });
  } finally {
    searchLoading.value = false;
  }
}

// 输入关键词进入搜索态；清空后恢复当前分类的分页列表
watch(searchText, async v => {
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

// ================= 进价（仅管理员可见，第一行小眼睛切换） =================
const isAdmin = ref(false);
const showCost = ref(false);

async function loadMe() {
  if (!hasToken()) return;
  try {
    const data = await fetchJson("/api/user/me");
    isAdmin.value = Number(data?.role) === 2;
  } catch {
    isAdmin.value = false;
  }
}

// ================= 图片/视频查看器（自制全屏预览，左右滑动切换） =================
const viewerOpen = ref(false);
const viewerIndex = ref(0);
const viewerMedia = ref<GoodsMedia[]>([]);
const viewerScroller = ref<HTMLElement | null>(null);

function goodsMediaOf(g: GoodsItem): GoodsMedia[] {
  if (g.media && g.media.length) return g.media;
  return g.image ? [{ hash: g.image, type: "image" }] : [];
}

function openViewer(g: GoodsItem) {
  const list = goodsMediaOf(g);
  if (list.length === 0) return;
  viewerMedia.value = list;
  // 首页展示第一张图片，打开时定位到它（无图片则从第一项开始）
  const firstImageIdx = list.findIndex(m => m.type === "image");
  viewerIndex.value = firstImageIdx >= 0 ? firstImageIdx : 0;
  viewerOpen.value = true;
  // 压入一条历史记录：手机返回键/侧滑返回先关预览，而不是退出页面或 App
  history.pushState({ lfddViewer: true }, "");
  // 打开后立即滚到目标项（等 DOM 渲染完）
  nextTick(() => {
    const el = viewerScroller.value;
    if (el) el.scrollLeft = viewerIndex.value * el.clientWidth;
  });
}

function closeViewer() {
  viewerOpen.value = false;
  pauseViewerVideos();
}

// 关闭按钮 / Esc：有自身历史记录时先回退，由 popstate 统一关闭
function requestCloseViewer() {
  if (history.state?.lfddViewer) history.back();
  else closeViewer();
}

// 返回键/侧滑返回：只关闭预览，不发其他导航
function onViewerPopState() {
  if (!viewerOpen.value) return;
  closeViewer();
}

// 离开当前项时暂停视频，避免后台继续出声
function pauseViewerVideos() {
  viewerScroller.value?.querySelectorAll("video").forEach(v => v.pause());
}

// 滚动结束按位置换算当前索引（用于圆点高亮）
function onViewerScroll() {
  const el = viewerScroller.value;
  if (!el || el.clientWidth === 0) return;
  const idx = Math.round(el.scrollLeft / el.clientWidth);
  if (idx !== viewerIndex.value) viewerIndex.value = idx;
}

// 箭头/圆点切换：平滑滚动到指定项
function goToMedia(i: number) {
  const n = viewerMedia.value.length;
  if (n === 0) return;
  const idx = Math.min(Math.max(i, 0), n - 1);
  viewerIndex.value = idx;
  const el = viewerScroller.value;
  if (el) el.scrollTo({ left: idx * el.clientWidth, behavior: "smooth" });
}

// 切到别的图片/视频时，暂停原来正在播放的视频
watch(viewerIndex, () => pauseViewerVideos());

// 桌面鼠标拖拽滑动（触屏走浏览器原生横向滚动，无需额外处理）
let mouseDrag = false;
let mouseStartX = 0;
let mouseStartLeft = 0;

function onViewerPointerDown(e: PointerEvent) {
  if (e.pointerType !== "mouse") return;
  if ((e.target as HTMLElement).closest("video")) return; // 不抢视频控件
  const el = viewerScroller.value;
  if (!el || viewerMedia.value.length < 2) return;
  mouseDrag = true;
  mouseStartX = e.clientX;
  mouseStartLeft = el.scrollLeft;
  el.style.scrollSnapType = "none"; // 拖动时关闭吸附，避免吸附抖动
}

function onViewerPointerMove(e: PointerEvent) {
  if (!mouseDrag) return;
  const el = viewerScroller.value;
  if (!el) return;
  el.scrollLeft = mouseStartLeft - (e.clientX - mouseStartX);
}

function onViewerPointerUp() {
  if (!mouseDrag) return;
  mouseDrag = false;
  const el = viewerScroller.value;
  if (!el) return;
  el.style.scrollSnapType = "";
  goToMedia(Math.round(el.scrollLeft / el.clientWidth));
}

// 键盘：Esc 关闭，左右方向键切换
function onViewerKey(e: KeyboardEvent) {
  if (!viewerOpen.value) return;
  if (e.key === "Escape") {
    requestCloseViewer();
  } else if (e.key === "ArrowLeft") {
    goToMedia(viewerIndex.value - 1);
  } else if (e.key === "ArrowRight") {
    goToMedia(viewerIndex.value + 1);
  }
}

function goCart() {
  void router.push("/cart");
}

// 数量输入框直接改数字：0/非法 → 移除该商品，超库存由 setQty 封顶
function onQtyInput(g: GoodsItem, v: string | number) {
  const n = Math.floor(Number(v));
  setQty(g, Number.isFinite(n) ? n : 0);
}

// ================= AI 点单（状态与购物车页共享，见 utils/aiOrder） =================
async function aiMatch() {
  if (!aiText.value.trim() || aiLoading.value) return;
  try {
    const count = await aiOrder();
    if (count === 0) {
      $q.notify({ message: "没有匹配到商品，换个说法试试", color: "orange" });
      return;
    }
    $q.notify({
      message: `已加入 ${count} 种商品`,
      color: "positive"
    });
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
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
  window.addEventListener("keydown", onViewerKey);
  window.addEventListener("popstate", onViewerPopState);
  await Promise.all([loadMenus(), loadGoods(true), loadMe()]);
  // 数据渲染后恢复上次的滚动位置
  await nextTick();
  if (sideRef.value) sideRef.value.scrollTop = pagePos.side;
  if (mainRef.value) mainRef.value.scrollTop = pagePos.main;
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", onViewerKey);
  window.removeEventListener("popstate", onViewerPopState);
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
  /* 顶部不留内边距：吸顶栏直接贴住滚动区顶部，避免滚动时上方露出内容 */
  padding: 0 12px 8px;
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

/* 点击商品图放大查看 */
.goods-img-clickable {
  cursor: zoom-in;
}

/* 全屏图片/视频查看器 */
.viewer-wrap {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: #000;
  overflow: hidden;
}

.viewer-close {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 3;
}

/* 原生横向滚动 + 吸附：触屏可直接左右滑，鼠标拖动见脚本 */
.viewer-scroller {
  display: flex;
  width: 100%;
  height: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  overscroll-behavior-x: contain;
  scroll-snap-type: x mandatory;
  scrollbar-width: none;
  cursor: grab;

  &::-webkit-scrollbar {
    display: none;
  }
}

.viewer-slide {
  flex: 0 0 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  scroll-snap-align: center;
  scroll-snap-stop: always;
}

.viewer-media {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  user-select: none;
  -webkit-user-drag: none;
}

.viewer-arrow {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  z-index: 2;
  background: rgba(0, 0, 0, 0.35);
}

.viewer-arrow-left {
  left: 8px;
}

.viewer-arrow-right {
  right: 8px;
}

.viewer-dots {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 16px;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.viewer-dot {
  width: 6px;
  height: 6px;
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.4);
  cursor: pointer;
  transition: width 0.2s, background 0.2s;

  &.active {
    width: 16px;
    background: #fff;
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

/* 管理员：进价展示与第一行小眼睛开关 */
.goods-name-row {
  gap: 8px;
}

.goods-cost-toggle {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  line-height: 1.2;
  /* 再上移一点，贴近卡片顶部 */
  margin-top: -7px;
}

.cost-eye {
  color: #999;
}

.goods-cost {
  flex: 0 0 auto;
  color: #999;
  white-space: nowrap;
  line-height: 1.2;
}

/* 价格行 + 加减区：窄屏换行时加减区依然固定靠右 */
.goods-foot {
  row-gap: 6px;
}

.goods-price-row {
  min-width: 0;
}

/* 吸顶顶部栏：模糊搜索 + AI 点单 */
.goods-topbar {
  position: sticky;
  top: 0;
  z-index: 6;
  /* 左右抵消 .goods-main 内边距，吸顶时铺满整宽并盖住滚动内容 */
  margin: 0 -12px;
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
  justify-content: flex-end;
  gap: 8px;
  /* 单独换行时也贴右侧，不与价格行错位 */
  margin-left: auto;
  flex: 0 0 auto;

  .add-btn {
    width: 36px;
    height: 36px;
  }

  .sub-btn {
    width: 32px;
    height: 32px;
  }
}

/* 未加入购物车时「−」和数量隐藏但保留占位，浮现时不会引起位移 */
.qty-placeholder {
  visibility: hidden;
  pointer-events: none;
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
    appearance: textfield;
    -moz-appearance: textfield;
  }

  :deep(input::-webkit-outer-spin-button),
  :deep(input::-webkit-inner-spin-button) {
    appearance: none;
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
