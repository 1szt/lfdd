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
        <div class="row items-center justify-between">
          <div class="col">
            <div class="text-body2 text-weight-medium">{{ o.order_no }}</div>
            <div class="text-caption text-grey-6">{{ o.created_at }}</div>
          </div>
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
            v-if="o.status === 0"
            color="teal"
            icon="edit"
            label="编辑"
            unelevated
            size="sm"
            @click="openEdit(o)"
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
            <div
              class="text-caption text-weight-medium"
              style="min-width: 64px; text-align: right"
            >
              ¥{{ fmtPrice(it.price * it.qty) }}
            </div>
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

          <div class="row q-gutter-sm q-mt-md">
            <q-btn
              color="primary"
              label="选择蓝牙设备"
              unelevated
              icon="bluetooth"
              :loading="printing"
              @click="printBle"
            />
            <q-btn
              outline
              color="primary"
              label="已配对设备"
              icon="list"
              :loading="loadingSpp"
              @click="loadSppDevices"
            />
          </div>

          <q-list v-if="sppDevices.length > 0" bordered separator class="q-mt-sm">
            <q-item v-for="d in sppDevices" :key="d.address" clickable @click="printSpp(d)">
              <q-item-section avatar><q-icon name="bluetooth" color="primary" /></q-item-section>
              <q-item-section>{{ d.name || d.address }}</q-item-section>
              <q-item-section side><q-icon name="print" color="grey-7" /></q-item-section>
            </q-item>
          </q-list>

          <div
            v-if="printStatus"
            class="q-mt-sm text-caption"
            :class="printStatusOk ? 'text-positive' : 'text-negative'"
          >
            {{ printStatus }}
          </div>
          <div v-else class="q-mt-sm text-caption text-grey-6">
            「选择蓝牙设备」弹系统选择框（BLE）；打印机是经典蓝牙时用「已配对设备」列表。
          </div>
        </q-card-section>

        <q-card-actions align="right">
          <q-btn flat label="关闭" v-close-popup />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- 编辑订单对话框（未结单：改数量 / 删行，保存时后端按差额调库存） -->
    <q-dialog v-model="editDlg">
      <q-card class="print-card">
        <q-card-section class="row items-center justify-between">
          <span class="text-body2 text-weight-medium">编辑订单</span>
          <q-icon name="edit" color="teal" />
        </q-card-section>

        <q-card-section class="q-pt-none">
          <div class="text-caption text-grey-7">订单 {{ editOrder?.order_no }}</div>

          <div v-for="(it, idx) in editItems" :key="it.goods_id" class="row items-center q-gutter-sm q-mt-sm">
            <div class="col">
              <div class="text-caption ellipsis">{{ it.name }}</div>
              <div v-if="it.spec" class="text-caption text-grey-6 ellipsis">{{ it.spec }}</div>
            </div>
            <q-input
              v-model="it.qty"
              type="number"
              dense
              outlined
              class="edit-qty"
              aria-label="修改数量"
              @change="onEditQty(idx)"
            />
            <q-btn
              flat
              dense
              round
              color="negative"
              icon="delete"
              size="sm"
              aria-label="移除该商品"
              @click="editItems.splice(idx, 1)"
            />
          </div>

          <div v-if="editItems.length === 0" class="text-caption text-grey-6 q-mt-sm">
            没有商品了，至少保留一件
          </div>
        </q-card-section>

        <q-card-actions align="right">
          <q-btn flat label="取消" v-close-popup />
          <q-btn
            color="teal"
            label="保存"
            unelevated
            :disable="editItems.length === 0"
            :loading="savingEdit"
            @click="saveEdit"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useQuasar } from "quasar";
import { fmtPrice, imgSrc } from "../../utils/cart";
import {
  bleInit,
  blePrint,
  buildOrderPayload,
  describeError,
  isNative,
  sppList,
  sppPrint,
  webBluetoothPrint,
  type ReceiptOrder,
  type SppDevice
} from "../../utils/printer";

const $q = useQuasar();

interface OrderItem {
  goods_id: string;
  name: string;
  price: number; // 单价，分
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

// ================= 编辑订单（未结单） =================
interface EditItem {
  goods_id: string;
  name: string;
  spec: string;
  qty: number;
}
const editDlg = ref(false);
const editOrder = ref<Order | null>(null);
const editItems = ref<EditItem[]>([]);
const savingEdit = ref(false);

function openEdit(o: Order) {
  editOrder.value = o;
  editItems.value = o.items.map((it) => ({
    goods_id: it.goods_id,
    name: it.name,
    spec: it.spec,
    qty: it.qty
  }));
  editDlg.value = true;
}

// 数量输入：0/非法 → 移除该行
function onEditQty(idx: number) {
  const n = Math.floor(Number(editItems.value[idx]?.qty));
  if (!Number.isFinite(n) || n <= 0) {
    editItems.value.splice(idx, 1);
  } else {
    editItems.value[idx]!.qty = n;
  }
}

// 保存：整体提交新明细，后端按差额调库存并重算件数
async function saveEdit() {
  if (!editOrder.value || savingEdit.value) return;
  const items = editItems.value
    .map((it) => ({ id: it.goods_id, qty: Math.floor(Number(it.qty)) }))
    .filter((it) => Number.isFinite(it.qty) && it.qty > 0);
  if (items.length === 0) {
    $q.notify({ message: "至少保留一件商品", color: "orange" });
    return;
  }
  savingEdit.value = true;
  try {
    const res = await fetch("/api/order/update", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id: editOrder.value.id, items })
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.message || `保存失败（${res.status}）`);
    $q.notify({ message: `已保存，共 ${data.count} 件`, color: "positive" });
    editDlg.value = false;
    await loadOrders(true);
  } catch (e) {
    $q.notify({ message: describeErrorLocal(e), color: "negative" });
  } finally {
    savingEdit.value = false;
  }
}

// ================= 打印小票 =================
const printDlg = ref(false);
const printOrder = ref<Order | null>(null);
const printing = ref(false);
const loadingSpp = ref(false);
const sppDevices = ref<SppDevice[]>([]);
const printStatus = ref("");
const printStatusOk = ref(false);
const printingId = ref<string | null>(null);

function openPrint(o: Order) {
  printOrder.value = o;
  printStatus.value = "";
  sppDevices.value = [];
  printDlg.value = true;
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

// BLE / 网页：弹系统设备选择框，选中即打印
async function printBle() {
  if (!printOrder.value || printing.value) return;
  printing.value = true;
  printingId.value = printOrder.value.id;
  try {
    if (isNative()) {
      try {
        await bleInit(); // 初始化（触发 Android「附近设备」权限）
      } catch {
        // 已初始化过则忽略
      }
    }
    const payload = buildOrderPayload(toReceipt(printOrder.value));
    const name = isNative() ? await blePrint(payload) : await webBluetoothPrint(payload);
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

// 经典蓝牙：读取系统已配对设备，点击即打印
async function loadSppDevices() {
  loadingSpp.value = true;
  try {
    sppDevices.value = await sppList();
    if (sppDevices.value.length === 0) {
      $q.notify({ message: "没有已配对的蓝牙设备，请先在系统设置中配对打印机", color: "orange" });
    }
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    loadingSpp.value = false;
  }
}

async function printSpp(d: SppDevice) {
  if (!printOrder.value || printing.value) return;
  printing.value = true;
  printingId.value = printOrder.value.id;
  try {
    const payload = buildOrderPayload(toReceipt(printOrder.value));
    const name = await sppPrint(d, payload);
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

/* 编辑对话框数量输入 */
.edit-qty {
  width: 76px;

  :deep(input) {
    text-align: center;
    font-weight: 700;
  }
}
</style>
