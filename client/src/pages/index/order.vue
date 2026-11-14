<template>
  <q-page class="order-page column" :style-fn="pageStyle">
    <!-- 标题栏：未结单 / 已结单 两个页签 -->
    <div class="row items-center justify-between q-px-md q-py-sm">
      <q-tabs
        v-model="statusTab"
        dense
        active-color="primary"
        class="col"
        @update:model-value="onTabChange"
      >
        <q-tab name="0" label="未结单" />
        <q-tab name="1" label="已结单" />
      </q-tabs>
      <q-btn flat dense size="sm" color="primary" icon="refresh" label="刷新" @click="loadOrders(true)" />
    </div>

    <!-- 空态 -->
    <div v-if="orders.length === 0 && !loading" class="column items-center justify-center empty-box">
      <q-icon name="receipt_long" size="64px" color="grey-4" />
      <div class="text-body1 text-grey-7 q-mt-md">
        {{ statusTab === "0" ? "没有未结单的订单" : "没有已结单的订单" }}
      </div>
      <q-btn v-if="statusTab === '0'" class="q-mt-md" color="primary" label="去点单" to="/" />
    </div>

    <div v-else class="order-list">
      <div v-if="loading" class="state-box text-grey-6">加载中…</div>
      <div v-if="errorMsg" class="state-box text-negative">{{ errorMsg }}</div>

      <!-- 订单卡片 -->
      <div v-for="o in orders" :key="o.id" class="order-card">
        <div class="text-body2 text-weight-medium">{{ o.order_no }}</div>
        <div class="text-caption text-grey-6">{{ o.created_at }}</div>

        <!-- 操作按钮单独一行，避免挤压订单号 -->
        <div class="row items-center q-gutter-xs q-mt-sm order-actions">
          <q-btn
            v-if="o.status === 0"
            color="orange"
            icon="payments"
            label="结单"
            unelevated
            size="sm"
            :loading="finishingId === o.id"
            @click="finishOrder(o)"
          />
          <q-btn
            color="primary"
            icon="shopping_cart"
            label="复制"
            unelevated
            size="sm"
            @click="copyToCart(o)"
          />
          <q-btn
            color="primary"
            icon="print"
            label="打印"
            unelevated
            size="sm"
            :loading="printingId === o.id"
            @click="openPrint(o)"
          />
          <q-btn
            v-if="o.status === 0"
            color="teal"
            icon="price_change"
            label="改价"
            unelevated
            size="sm"
            @click="openPrice(o)"
          />
          <q-btn
            v-if="o.status === 0"
            color="negative"
            icon="block"
            label="废单"
            unelevated
            size="sm"
            :loading="voidingId === o.id"
            @click="confirmVoid(o)"
          />
          <q-btn
            v-if="o.status === 1"
            color="negative"
            icon="delete"
            label="删除"
            unelevated
            size="sm"
            :loading="deletingId === o.id"
            @click="confirmDelete(o)"
          />
        </div>

        <div class="order-items q-mt-sm">
          <div v-for="(it, idx) in o.items" :key="idx" class="order-item row items-center">
            <div class="item-thumb">
              <img v-if="it.image" :src="imgSrc(it.image)" :alt="it.name" />
              <q-icon v-else name="shopping_bag" size="20px" color="grey-5" />
            </div>
            <div class="col">
              <div class="text-caption ellipsis">{{ it.name }}</div>
              <div v-if="it.spec" class="text-caption text-grey-6 ellipsis">{{ it.spec }}</div>
            </div>
            <div class="text-caption text-grey-7">x{{ it.qty }}</div>
            <div class="column items-end" style="min-width: 96px">
              <!-- 单价：改过价显示原价划线 + 改后价 -->
              <div class="text-caption text-grey-7">
                <template v-if="it.edited === 1">
                  <span class="text-grey-5" style="text-decoration: line-through">
                    ¥{{ fmtPrice(it.orig_price) }}
                  </span>
                  <span class="text-weight-medium text-negative q-ml-xs">
                    ¥{{ fmtPrice(it.price) }}
                  </span>
                </template>
                <template v-else>¥{{ fmtPrice(it.price) }}</template>
              </div>
              <div class="text-caption text-weight-medium">¥{{ fmtPrice(it.price * it.qty) }}</div>
            </div>
            <q-chip
              v-if="it.edited === 1"
              dense
              size="xs"
              flat
              color="deep-orange"
              text-color="white"
              class="q-ma-none"
            >
              改价
            </q-chip>
          </div>
        </div>

        <div class="row items-center justify-end q-gutter-sm q-mt-sm">
          <span class="text-caption text-grey-7">共 {{ o.count }} 件</span>
          <span class="text-caption text-weight-bold text-negative">合计 ¥{{ fmtPrice(o.total) }}</span>
          <q-chip v-if="o.status === 0" color="orange" text-color="white" dense size="sm" class="q-ma-none">
            待结单
          </q-chip>
        </div>
      </div>

      <div class="text-center q-pb-md">
        <q-btn
          v-if="hasMore"
          flat
          dense
          color="primary"
          :loading="loadingMore"
          label="加载更多"
          @click="loadOrders(false)"
        />
        <span v-else-if="orders.length > 0" class="text-caption text-grey-6">— 没有更多了 —</span>
      </div>
    </div>

    <!-- 打印小票对话框 -->
    <q-dialog v-model="printDlg">
      <q-card class="print-card">
        <q-card-section class="row items-center justify-between">
          <span class="text-body2 text-weight-medium">打印小票</span>
          <q-icon name="print" color="primary" />
        </q-card-section>

        <q-card-section class="q-pt-none">
          <div class="text-caption text-grey-7">订单 {{ printOrder?.order_no }}</div>

          <!-- 已添加设备：点选即打印（编码按各设备设置） -->
          <q-list v-if="printers.length > 0" bordered separator class="q-mt-sm">
            <q-item
              v-for="p in printers"
              :key="p.key"
              clickable
              v-ripple
              :disable="printing"
              @click="printWith(p)"
            >
              <q-item-section avatar>
                <q-icon
                  :name="p.kind === 'label' ? 'label' : p.conn === 'spp' ? 'settings_input_antenna' : 'bluetooth'"
                  :color="p.kind === 'label' ? 'deep-purple' : p.conn === 'spp' ? 'teal' : 'primary'"
                />
              </q-item-section>
              <q-item-section>
                <q-item-label>{{ p.name }}</q-item-label>
                <q-item-label caption>
                  {{ p.kind === "label" ? "标签打印机" : "小票机" }} ·
                  {{ p.transport === "wifi" ? "WiFi" : p.conn === "spp" ? "传统蓝牙" : "低功耗" }}
                  · {{ p.encoding === "utf8" ? "UTF-8" : "GBK" }}
                </q-item-label>
              </q-item-section>
              <q-item-section side><q-icon name="print" color="grey-7" /></q-item-section>
            </q-item>
          </q-list>

          <!-- 没有设备：引导去管理页添加 -->
          <div v-else class="column items-center q-mt-md">
            <q-icon name="printer" size="48px" color="grey-4" />
            <div class="text-body2 text-grey-7 q-mt-sm">还没有添加打印机</div>
            <q-btn
              class="q-mt-sm"
              color="primary"
              label="去添加设备"
              unelevated
              @click="goManagePrinter"
            />
          </div>

          <div
            v-if="printStatus"
            class="q-mt-sm text-caption"
            :class="printStatusOk ? 'text-positive' : 'text-negative'"
          >
            {{ printStatus }}
          </div>
        </q-card-section>

        <q-card-actions align="right">
          <q-btn flat label="关闭" v-close-popup />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- 批量改价对话框（未结单：整单所有商品改单价，改回原价即重置） -->
    <q-dialog v-model="priceDlg">
      <q-card class="print-card price-card">
        <q-card-section class="row items-center justify-between">
          <span class="text-body2 text-weight-medium">批量改价</span>
          <q-icon name="price_change" color="primary" />
        </q-card-section>

        <q-card-section class="q-pt-none">
          <div class="text-caption text-grey-7">订单 {{ priceOrder?.order_no }}</div>

          <div class="q-mt-sm price-rows">
            <div v-for="r in priceRows" :key="r.goods_id" class="price-row row items-center">
              <div class="col">
                <div class="text-caption ellipsis">{{ r.name }}</div>
                <div class="text-caption text-grey-6">
                  原价 ¥{{ fmtPrice(r.orig) }}
                  <span v-if="r.edited" class="text-weight-medium text-negative q-ml-xs">
                    现改价 ¥{{ fmtPrice(r.initial) }}
                  </span>
                </div>
              </div>
              <q-input
                v-model="r.val"
                type="number"
                dense
                outlined
                class="price-input"
                :aria-label="`${r.name}新单价`"
              />
              <q-btn
                flat
                dense
                round
                color="grey-7"
                icon="undo"
                size="sm"
                :aria-label="`重置${r.name}单价`"
                @click="resetRow(r)"
              />
            </div>
          </div>
        </q-card-section>

        <q-card-actions align="right">
          <q-btn flat label="全部重置" color="grey-8" icon="restart_alt" @click="resetAll" />
          <q-btn flat label="取消" v-close-popup />
          <q-btn color="primary" label="保存" unelevated :loading="savingPrice" @click="savePrice" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useQuasar } from "quasar";
import { useRouter } from "vue-router";
import { copyOrderToCart, fmtPrice, imgSrc } from "../../utils/cart";
import {
  bleInit,
  blePrintTo,
  buildOrderPayload,
  describeError,
  isNative,
  sppPrint,
  webBluetoothPrintTo,
  type ReceiptOrder
} from "../../utils/printer";
import { printers, type SavedPrinter } from "../../utils/printerStore";

const $q = useQuasar();
const router = useRouter();

interface OrderItem {
  goods_id: string;
  name: string;
  price: number; // 单价，分（未改价的待结单单为商品当前价；改价/结单后为锁定价）
  edited: number; // 是否改过价（改价标记）
  orig_price: number; // 原价，分（未结单=商品现价；已结单=结单时商品价）
  spec: string;
  qty: number;
  image: string;
}
interface Order {
  id: string;
  order_no: string;
  total: number; // 合计，分（待结单时为 0）
  count: number;
  status: number; // 0 待结单、1 已结单
  created_at: string;
  items: OrderItem[];
}

const PAGE_SIZE = 30;
const statusTab = ref("0"); // "0" 未结单、"1" 已结单
const orders = ref<Order[]>([]);
const total = ref(0);
const page = ref(1);
const loading = ref(false);
const loadingMore = ref(false);
const errorMsg = ref("");

const hasMore = computed(() => orders.value.length < total.value);

// 切页签：重新拉对应状态的第一页
function onTabChange() {
  void loadOrders(true);
}

// QPage styleFn：header/footer 占位后剩余高度，让页面撑满、列表区各自滚动
function pageStyle(offset: number, height: number) {
  if (height > 0) {
    return { height: `${height - offset}px` };
  }
  return { height: offset !== 0 ? `calc(100vh - ${offset}px)` : "100vh" };
}

function describeErrorLocal(e: unknown) {
  return e instanceof Error ? e.message : String(e);
}

// 订单分页（reset=true 重新第一页，false 加载更多）
async function loadOrders(reset: boolean) {
  if (reset) {
    page.value = 1;
    errorMsg.value = "";
    loading.value = true;
  } else {
    loadingMore.value = true;
  }
  try {
    const res = await fetch(
      `/api/order/page?page=${page.value}&page_size=${PAGE_SIZE}&status=${statusTab.value}`
    );
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.message || `请求失败（${res.status}）`);

    const list: Order[] = Array.isArray(data.list) ? data.list : [];
    orders.value = reset ? list : [...orders.value, ...list];
    total.value = typeof data.total === "number" ? data.total : 0;
    if (list.length > 0) page.value += 1;
  } catch (e) {
    if (reset) {
      orders.value = [];
      total.value = 0;
    }
    errorMsg.value = `加载失败：${describeErrorLocal(e)}，请确认后端服务已启动`;
  } finally {
    loading.value = false;
    loadingMore.value = false;
  }
}

const finishingId = ref<string | null>(null); // 结单中的订单

// 结单：按当前价锁定本单价格，成功后刷新列表（锁价结果从后端重取）
async function finishOrder(o: Order) {
  if (finishingId.value) return;
  finishingId.value = o.id;
  try {
    const res = await fetch("/api/order/finish", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id: o.id })
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.message || `结单失败（${res.status}）`);
    $q.notify({ message: `结单成功，合计 ¥${fmtPrice(data.total)}`, color: "positive" });
    await loadOrders(true);
  } catch (e) {
    $q.notify({ message: describeErrorLocal(e), color: "negative" });
  } finally {
    finishingId.value = null;
  }
}

const deletingId = ref<string | null>(null); // 删除中的订单

// 删除已结单订单：二次确认防手误
function confirmDelete(o: Order) {
  $q.dialog({
    title: "删除订单",
    message: `确定删除订单 ${o.order_no} 吗？删除后不可恢复`,
    cancel: true,
    persistent: true
  }).onOk(() => {
    void deleteOrder(o);
  });
}

async function deleteOrder(o: Order) {
  if (deletingId.value) return;
  deletingId.value = o.id;
  try {
    const res = await fetch("/api/order/delete", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id: o.id })
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.message || `删除失败（${res.status}）`);
    $q.notify({ message: "已删除", color: "positive" });
    await loadOrders(true);
  } catch (e) {
    $q.notify({ message: describeErrorLocal(e), color: "negative" });
  } finally {
    deletingId.value = null;
  }
}

// ================= 复制到购物车 / 废单 =================

// 商品 ID → 分类 ID 映射（现有商品接口反推，供复制后侧边栏角标统计）。
// 分页拉全量（每页 100，上限 20 页），已下架/已删商品查不到则不计角标。
async function fetchMenuMap(): Promise<Record<string, string>> {
  const map: Record<string, string> = {};
  try {
    let got = 0;
    let total = Infinity;
    for (let page = 1; page <= 20 && got < total; page++) {
      const res = await fetch(`/api/goods/page?menu_id=0&page=${page}&page_size=100`);
      const data = await res.json().catch(() => ({}));
      if (!res.ok) break;
      const list: { id: string; menu_id: string }[] = Array.isArray(data.list) ? data.list : [];
      for (const g of list) map[g.id] = g.menu_id;
      got += list.length;
      total = typeof data.total === "number" ? data.total : got;
      if (list.length === 0) break;
    }
  } catch {
    // 拿不到映射就只影响角标，不阻断复制
  }
  return map;
}

// 复制订单明细进购物车，去购物车改完重新提交订单
async function copyToCart(o: Order) {
  const menuMap = await fetchMenuMap();
  copyOrderToCart(
    o.items.map((it) => ({
      goods_id: it.goods_id,
      menu_id: menuMap[it.goods_id] ?? "",
      name: it.name,
      price: it.price,
      spec: it.spec,
      image: it.image,
      qty: it.qty
    }))
  );
  $q.notify({ message: `已复制 ${o.count} 件到购物车`, color: "positive" });
  void router.push("/cart");
}

// ================= 批量改价 =================
const priceDlg = ref(false);
const priceOrder = ref<Order | null>(null);
const savingPrice = ref(false);

// 改价行：orig=原价（商品现价，分）、initial=打开时生效价（分）、val=输入框（元）
interface PriceRow {
  goods_id: string;
  name: string;
  orig: number;
  initial: number;
  edited: number;
  val: string;
}
const priceRows = ref<PriceRow[]>([]);

function openPrice(o: Order) {
  priceOrder.value = o;
  priceRows.value = o.items.map((it) => ({
    goods_id: it.goods_id,
    name: it.name,
    orig: it.orig_price || it.price,
    initial: it.price,
    edited: it.edited,
    val: fmtPrice(it.price)
  }));
  priceDlg.value = true;
}

// 单行重置：填回原价（保存时与商品现价一致会自动取消改价标记）
function resetRow(r: PriceRow) {
  r.val = fmtPrice(r.orig);
}

// 全部重置：所有行填回原价
function resetAll() {
  for (const r of priceRows.value) r.val = fmtPrice(r.orig);
}

// 保存：只提交有变动的行（元 → 分），成功后刷新（合计/标记由后端算）
async function savePrice() {
  if (!priceOrder.value || savingPrice.value) return;
  const changed: { goods_id: string; price: number }[] = [];
  for (const r of priceRows.value) {
    const yuan = Number(r.val);
    if (!Number.isFinite(yuan) || yuan < 0) {
      $q.notify({ message: `「${r.name}」请输入有效价格`, color: "orange" });
      return;
    }
    const fen = Math.round(yuan * 100);
    if (fen !== r.initial) changed.push({ goods_id: r.goods_id, price: fen });
  }
  if (changed.length === 0) {
    $q.notify({ message: "价格没有变动", color: "grey-7" });
    priceDlg.value = false;
    return;
  }
  savingPrice.value = true;
  try {
    await Promise.all(
      changed.map((c) =>
        fetch("/api/order/price", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ id: priceOrder.value!.id, ...c })
        }).then(async (res) => {
          const data = await res.json().catch(() => ({}));
          if (!res.ok) throw new Error(data.message || `改价失败（${res.status}）`);
        })
      )
    );
    $q.notify({ message: `已改价 ${changed.length} 项`, color: "positive" });
    priceDlg.value = false;
    await loadOrders(true);
  } catch (e) {
    $q.notify({ message: describeErrorLocal(e), color: "negative" });
  } finally {
    savingPrice.value = false;
  }
}

const voidingId = ref<string | null>(null); // 废单中的订单
// 废单：作废未结单订单（回补库存并删除），二次确认防手误
function confirmVoid(o: Order) {
  $q.dialog({
    title: "废单",
    message: `确定作废订单 ${o.order_no} 吗？库存会退回，订单删除且不可恢复`,
    cancel: true,
    persistent: true
  }).onOk(() => {
    void voidOrder(o);
  });
}

async function voidOrder(o: Order) {
  if (voidingId.value) return;
  voidingId.value = o.id;
  try {
    const res = await fetch("/api/order/void", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id: o.id })
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.message || `废单失败（${res.status}）`);
    $q.notify({ message: "已废单，库存已退回", color: "positive" });
    await loadOrders(true);
  } catch (e) {
    $q.notify({ message: describeErrorLocal(e), color: "negative" });
  } finally {
    voidingId.value = null;
  }
}

// ================= 打印小票 =================
const printDlg = ref(false);
const printOrder = ref<Order | null>(null);
const printing = ref(false);
const printStatus = ref("");
const printStatusOk = ref(false);
const printingId = ref<string | null>(null);

function openPrint(o: Order) {
  printOrder.value = o;
  printStatus.value = "";
  printStatusOk.value = false;
  printDlg.value = true;
}

// 没有已添加设备时跳设备管理页
function goManagePrinter() {
  printDlg.value = false;
  void router.push("/admin");
}

function toReceipt(o: Order): ReceiptOrder {
  return {
    orderNo: o.order_no,
    createdAt: o.created_at,
    count: o.count,
    total: o.total,
    items: o.items.map((it) => ({ name: it.name, spec: it.spec, qty: it.qty, price: it.price }))
  };
}

// 选中已添加设备打印：按设备连接方式（BLE/SPP）与编码格式出票
// WiFi / 标签为占位项，暂不支持
async function printWith(p: SavedPrinter) {
  if (!printOrder.value || printing.value) return;
  if (p.transport === "wifi") {
    printStatusOk.value = false;
    printStatus.value = "WiFi 打印即将支持，暂无法使用该设备";
    return;
  }
  if (p.kind === "label") {
    printStatusOk.value = false;
    printStatus.value = "标签打印即将支持，暂无法使用该设备；小票打印请在管理页把设备类型改为小票机";
    return;
  }
  printing.value = true;
  printingId.value = printOrder.value.id;
  try {
    const payload = buildOrderPayload(toReceipt(printOrder.value), p.encoding);
    let name: string;
    if (p.conn === "spp") {
      if (!isNative()) throw new Error("网页端不支持传统蓝牙打印，请在 App 中打印");
      name = await sppPrint({ address: p.id, name: p.name }, payload);
    } else if (isNative()) {
      try {
        await bleInit(); // 初始化（内部触发 Android「附近设备」权限）
      } catch {
        // 已初始化过则忽略（权限被拒时由下面 blePrintTo 统一报错）
      }
      name = await blePrintTo(p.id, payload, p.name);
    } else {
      await webBluetoothPrintTo(p.id, payload);
      name = p.name;
    }
    printStatusOk.value = true;
    printStatus.value = `已发送到「${name}」，请查看打印机出纸`;
  } catch (e) {
    printStatusOk.value = false;
    printStatus.value = describeError(e);
  } finally {
    printing.value = false;
    printingId.value = null;
  }
}

onMounted(() => {
  void loadOrders(true);
});
</script>

<style scoped lang="scss">
.order-page {
  overflow: hidden;
  background: #fff;
}

.empty-box {
  flex: 1;
}

.order-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 0 16px;
}

.state-box {
  padding: 24px 0;
  text-align: center;
}

.order-card {
  padding: 12px;
  margin-bottom: 12px;
  border: 1px solid #ececec;
  border-radius: 10px;
  background: #fff;
}

.order-items {
  border-top: 1px solid #f5f5f5;
}

.order-item {
  gap: 10px;
  padding: 8px 0;
  border-bottom: 1px solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }
}

.item-thumb {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  border-radius: 6px;
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

.print-card {
  min-width: 340px;
  max-width: 90vw;
}

/* 操作按钮行：窄屏自动折行，不与订单号抢空间 */
.order-actions {
  flex-wrap: wrap;
}

/* 批量改价对话框 */
.price-card {
  width: 420px;
}

.price-rows {
  max-height: 50vh;
  overflow-y: auto;
}

.price-row {
  gap: 8px;
  padding: 8px 0;
  border-bottom: 1px solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }

  .price-input {
    width: 110px;
    flex-shrink: 0;
  }
}
</style>
