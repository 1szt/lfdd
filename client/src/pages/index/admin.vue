<template>
  <q-page class="q-pa-md">
    <!-- ================= 未登录：登录 / 注册 ================= -->
    <div v-if="!token" class="flex flex-center" style="min-height: 60vh">
      <q-card style="width: 360px; max-width: 95vw">
        <q-tabs
          v-model="authTab"
          dense
          class="text-grey-7"
          active-color="primary"
          indicator-color="primary"
          align="justify"
          narrow-indicator
        >
          <q-tab name="login" label="登录" />
          <q-tab name="register" label="注册" />
        </q-tabs>
        <q-separator />

        <q-card-section class="column q-gutter-sm">
          <q-input
            v-if="authTab === 'register'"
            v-model="regEmail"
            label="邮箱"
            type="email"
            outlined
            dense
            autofocus
          />
          <q-input
            v-else
            v-model="account"
            label="账号（用户名 / 邮箱）"
            outlined
            dense
            autofocus
          />
          <q-input
            v-model="password"
            label="密码"
            type="password"
            outlined
            dense
            :hint="authTab === 'register' ? '至少 8 位' : undefined"
            @keyup.enter="submitAuth"
          />
          <div v-if="authError" class="text-negative text-caption">{{ authError }}</div>
          <q-btn
            color="primary"
            :label="authTab === 'login' ? '登录' : '注册并登录'"
            :loading="authBusy"
            @click="submitAuth"
          />
        </q-card-section>
      </q-card>
    </div>

    <!-- ================= 已登录：管理界面 ================= -->
    <div v-else class="admin-wrap">
      <div class="row items-center q-mb-sm">
        <div class="text-h6">管理</div>
        <q-space />
        <q-btn flat dense color="negative" icon="logout" label="退出登录" @click="logout" />
      </div>

      <q-tabs
        v-model="tab"
        dense
        class="text-grey-7"
        active-color="primary"
        indicator-color="primary"
        align="justify"
        narrow-indicator
      >
        <q-tab name="goods" label="商品管理" icon="inventory_2" />
        <q-tab name="menus" label="分类管理" icon="category" />
        <q-tab name="print" label="设备管理" icon="print" />
      </q-tabs>
      <q-separator />

      <q-tab-panels v-model="tab" animated class="bg-transparent">
        <!-- ============ 商品管理 ============ -->
        <q-tab-panel name="goods" class="q-pa-none q-pt-sm">
          <!-- 分类筛选：横向点选，替代下拉框 -->
          <div class="row no-wrap q-gutter-xs q-px-none" style="overflow-x: auto">
            <q-btn
              v-for="opt in menuFilterOptions"
              :key="opt.value"
              dense
              unelevated
              size="sm"
              :label="opt.label"
              :color="String(filterMenu) === String(opt.value) ? 'primary' : 'grey-3'"
              :text-color="String(filterMenu) === String(opt.value) ? 'white' : 'grey-8'"
              @click="setFilter(opt.value)"
            />
          </div>
          <q-toolbar class="q-px-none">
            <div class="text-grey-7 text-caption">拖动商品行调整顺序</div>
            <q-space />
            <q-btn
              color="primary"
              icon="add"
              label="新建商品"
              :disable="menus.length === 0"
              @click="openGoods()"
            />
          </q-toolbar>

          <div v-if="menus.length === 0" class="text-grey-6 q-pa-md">
            还没有分类，先到「分类管理」创建一个分类。
          </div>

          <div v-else-if="goodsLoading" class="text-grey-6 q-pa-md">加载中…</div>

          <div v-else-if="goodsList.length === 0" class="text-grey-6 q-pa-md">
            该分类下暂无商品，点右上角「新建商品」添加。
          </div>

          <q-list v-else separator>
            <q-item
              v-for="(g, gIdx) in goodsList"
              :key="g.id"
              class="menu-row"
              draggable="true"
              @dragstart="onGoodsDragStart(gIdx)"
              @dragover.prevent
              @drop="onGoodsDrop(gIdx)"
              @dragend="dragGoodsIdx = null"
            >
              <q-item-section avatar>
                <div class="goods-thumb">
                  <img v-if="g.image" :src="imgSrc(g.image)" :alt="g.name" />
                  <q-icon v-else name="shopping_bag" size="22px" color="grey-5" />
                </div>
              </q-item-section>
              <q-item-section>
                <q-item-label class="text-weight-medium">{{ g.name }}</q-item-label>
                <q-item-label caption>
                  {{ menuName(g.menu_id) }}
                  <span v-if="g.spec"> · {{ g.spec }}</span>
                  · 库存 {{ g.stock === -1 ? "不限" : g.stock }}
                </q-item-label>
              </q-item-section>
              <q-item-section side class="text-negative text-weight-bold">
                ¥{{ (g.price / 100).toFixed(2) }}
              </q-item-section>
              <q-item-section side>
                <q-toggle
                  :model-value="g.status === 1"
                  :label="g.status === 1 ? '上架' : '下架'"
                  dense
                  color="positive"
                  @update:model-value="(v: boolean) => toggleStatus(g, v)"
                />
              </q-item-section>
              <q-item-section side>
                <div class="row no-wrap">
                  <q-btn dense flat round icon="edit" @click="openGoods(g)" />
                  <q-btn dense flat round icon="delete" color="negative" @click="removeGoods(g)" />
                </div>
              </q-item-section>
            </q-item>
          </q-list>
        </q-tab-panel>

        <!-- ============ 分类管理 ============ -->
        <q-tab-panel name="menus" class="q-pa-none q-pt-sm">
          <q-toolbar class="q-px-none">
            <div class="text-grey-7 text-caption">拖动分类行调整前台侧边栏顺序</div>
            <q-space />
            <q-btn color="primary" icon="add" label="新建分类" @click="openMenu()" />
          </q-toolbar>

          <div v-if="menus.length === 0" class="text-grey-6 q-pa-md">暂无分类。</div>

          <q-list v-else separator>
            <q-item
              v-for="(m, idx) in menus"
              :key="m.id"
              class="menu-row"
              draggable="true"
              @dragstart="onMenuDragStart(idx)"
              @dragover.prevent
              @drop="onMenuDrop(idx)"
              @dragend="dragMenuIdx = null"
            >
              <q-item-section avatar>
                <q-icon name="drag_indicator" color="grey-5" />
              </q-item-section>
              <q-item-section>
                <q-item-label>{{ m.name }}</q-item-label>
                <q-item-label caption>ID {{ m.id }}</q-item-label>
              </q-item-section>
              <q-item-section side>
                <div class="row no-wrap">
                  <q-btn dense flat round icon="edit" @click="openMenu(m)" />
                  <q-btn dense flat round icon="delete" color="negative" @click="removeMenu(m)" />
                </div>
              </q-item-section>
            </q-item>
          </q-list>
        </q-tab-panel>

        <!-- ============ 蓝牙搜索（双列：低功耗 / 传统） ============ -->
        <q-tab-panel name="print" class="q-pa-none q-pt-sm">
          <q-toolbar class="q-px-none">
            <q-btn
              color="primary"
              icon="bluetooth_searching"
              :label="scanning ? '搜索中…' : '搜索蓝牙'"
              :loading="scanning"
              @click="startScan"
            />
            <q-btn
              v-if="scanning"
              class="q-ml-sm"
              flat
              label="停止"
              @click="stopScan"
            />
            <q-space />
            <span class="text-caption text-grey-6">搜索后点击设备添加；点已保存设备发测试小票</span>
          </q-toolbar>

          <q-banner
            v-if="status"
            dense
            rounded
            class="q-mb-sm"
            :class="statusOk ? 'bg-green-1 text-green-9' : 'bg-red-1 text-red-9'"
          >
            {{ status }}
          </q-banner>

          <!-- 已保存设备：其他页面打印时直接从这里选 -->
          <q-list bordered separator class="q-mb-sm">
            <q-item-label header class="bg-grey-2 text-grey-8">
              已添加打印机（{{ printers.length }}）
            </q-item-label>
            <q-item v-if="printers.length === 0">
              <q-item-section class="text-grey-6">
                还没有添加打印机，先搜索并点击下方设备添加
              </q-item-section>
            </q-item>
            <q-item v-for="p in printers" :key="p.conn + p.id">
              <q-item-section avatar>
                <q-icon
                  :name="p.conn === 'spp' ? 'settings_input_antenna' : 'bluetooth'"
                  :color="p.conn === 'spp' ? 'teal' : 'primary'"
                />
              </q-item-section>
              <q-item-section>
                <q-item-label>{{ p.name }}</q-item-label>
                <q-item-label caption>
                  {{ p.conn === "spp" ? "传统" : "低功耗" }} · {{ p.id }} ·
                  {{ p.encoding === "utf8" ? "UTF-8" : "GBK" }}
                </q-item-label>
              </q-item-section>
              <q-item-section side>
                <div class="row no-wrap q-gutter-xs">
                  <q-btn dense flat round icon="edit" aria-label="编辑打印机" @click="openEdit(p)" />
                  <q-btn
                    dense
                    flat
                    round
                    icon="print"
                    color="primary"
                    aria-label="测试打印"
                    @click="testPrinter(p)"
                  />
                  <q-btn
                    dense
                    flat
                    round
                    icon="delete"
                    color="negative"
                    aria-label="删除打印机"
                    @click="confirmRemovePrinter(p)"
                  />
                </div>
              </q-item-section>
            </q-item>
          </q-list>

          <div class="row q-col-gutter-sm full-height">
            <!-- 低功耗蓝牙 -->
            <div class="col-6">
              <q-list bordered separator class="scan-list">
                <q-item-label header class="bg-blue-1 text-blue-8">
                  低功耗蓝牙（BLE）
                </q-item-label>
                <q-item v-if="bleDevices.length === 0">
                  <q-item-section class="text-grey-6">
                    {{ scanning ? "搜索中，等待附近设备广播…" : "未搜索到设备" }}
                  </q-item-section>
                </q-item>
                <q-item
                  v-for="d in bleDevices"
                  :key="d.deviceId"
                  clickable
                  v-ripple
                  @click="openAdd(d, 'ble')"
                >
                  <q-item-section avatar>
                    <q-icon name="bluetooth" color="primary" />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label>{{ d.name || "未命名设备" }}</q-item-label>
                    <q-item-label caption>{{ d.deviceId }}</q-item-label>
                  </q-item-section>
                  <q-item-section v-if="d.rssi != null" side caption>
                    {{ d.rssi }} dBm
                  </q-item-section>
                </q-item>
              </q-list>
            </div>

            <!-- 传统蓝牙 -->
            <div class="col-6">
              <q-list bordered separator class="scan-list">
                <q-item-label header class="bg-teal-1 text-teal-8">
                  传统蓝牙（SPP）
                </q-item-label>
                <q-item v-if="sppDevices.length === 0">
                  <q-item-section class="text-grey-6">
                    {{
                      !isNative()
                        ? "传统蓝牙仅 App 内可用"
                        : scanning
                          ? "搜索中，等待设备进入配对模式…"
                          : "未搜索到设备"
                    }}
                  </q-item-section>
                </q-item>
                <q-item
                  v-for="d in sppDevices"
                  :key="d.address"
                  clickable
                  v-ripple
                  @click="openAdd(d, 'spp')"
                >
                  <q-item-section avatar>
                    <q-icon name="settings_input_antenna" color="teal" />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label>{{ d.name || "未命名设备" }}</q-item-label>
                    <q-item-label caption>{{ d.address }}</q-item-label>
                  </q-item-section>
                </q-item>
              </q-list>
            </div>
          </div>
        </q-tab-panel>
      </q-tab-panels>

      <!-- ============ 添加 / 编辑打印机 ============ -->
      <q-dialog v-model="devDlg" persistent>
        <q-card style="min-width: 320px; max-width: 95vw">
          <q-card-section class="text-subtitle1">
            {{ devForm.editing ? "编辑打印机" : "添加打印机" }}
          </q-card-section>
          <q-card-section class="column q-gutter-sm">
            <q-input v-model="devForm.name" label="设备名称" outlined dense autofocus />
            <div class="text-caption text-grey-6">
              地址 {{ devForm.id }} · {{ devForm.conn === "spp" ? "传统蓝牙" : "低功耗蓝牙" }}
            </div>
            <q-select
              v-model="devForm.encoding"
              :options="encOptions"
              label="编码格式"
              outlined
              dense
              emit-value
              map-options
            />
            <div class="text-caption text-grey-6">
              国产热敏机一般选 GBK；打印中文乱码时切换另一种再试
            </div>
          </q-card-section>
          <q-card-actions align="right">
            <q-btn flat label="取消" v-close-popup />
            <q-btn color="primary" label="保存" unelevated @click="saveDev" />
          </q-card-actions>
        </q-card>
      </q-dialog>

      <!-- ============ 新建 / 编辑商品 ============ -->
      <q-dialog v-model="goodsDlg">
        <q-card style="min-width: 360px; max-width: 95vw">
          <q-card-section class="text-subtitle1">
            {{ goodsForm.id ? "编辑商品" : "新建商品" }}
          </q-card-section>

          <q-card-section class="column q-gutter-sm">
            <q-select
              v-model="goodsForm.menu_id"
              :options="menuOptions"
              emit-value
              map-options
              label="所属分类"
              outlined
              dense
            />
            <q-input v-model="goodsForm.name" label="名称" outlined dense autofocus />
            <q-input
              v-model="goodsForm.priceYuan"
              label="价格（元）"
              type="number"
              step="0.01"
              min="0"
              outlined
              dense
            />
            <q-input v-model="goodsForm.spec" label="规格（如：大杯 / 500g，可空）" outlined dense />
            <q-input
              v-model="goodsForm.stock"
              label="库存"
              type="number"
              outlined
              dense
              hint="-1 = 不限库存，0 = 售罄"
            />
            <!-- 图片：选择文件即上传（哈希命名），image 字段只存哈希 -->
            <q-file
              v-model="goodsImageFile"
              outlined
              dense
              label="上传图片（可空，jpg/png/gif/webp，≤5MB）"
              accept="image/*"
              :loading="imgUploading"
              @update:model-value="uploadGoodsImage"
            >
              <template #prepend>
                <q-icon name="upload" />
              </template>
            </q-file>
            <div v-if="goodsForm.image" class="row items-center q-gutter-sm">
              <img
                :src="imgSrc(goodsForm.image)"
                class="goods-thumb"
                style="width: 48px; height: 48px"
                alt=""
              />
              <span class="text-caption text-grey-6">预览</span>
              <q-btn
                flat
                dense
                size="sm"
                color="negative"
                icon="close"
                label="移除"
                @click="goodsForm.image = ''"
              />
            </div>
            <q-toggle v-model="goodsForm.statusOn" label="上架销售" color="positive" />
          </q-card-section>

          <q-card-actions align="right">
            <q-btn flat label="取消" v-close-popup />
            <q-btn color="primary" label="保存" :loading="goodsSaving" @click="saveGoods" />
          </q-card-actions>
        </q-card>
      </q-dialog>

      <!-- ============ 新建 / 编辑分类 ============ -->
      <q-dialog v-model="menuDlg">
        <q-card style="min-width: 300px; max-width: 95vw">
          <q-card-section class="text-subtitle1">
            {{ menuForm.id ? "编辑分类" : "新建分类" }}
          </q-card-section>
          <q-card-section>
            <q-input
              v-model="menuForm.name"
              label="分类名称"
              outlined
              dense
              autofocus
              @keyup.enter="saveMenu"
            />
          </q-card-section>
          <q-card-actions align="right">
            <q-btn flat label="取消" v-close-popup />
            <q-btn color="primary" label="保存" :loading="menuSaving" @click="saveMenu" />
          </q-card-actions>
        </q-card>
      </q-dialog>

    </div>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useQuasar } from "quasar";
import {
  bleInit,
  blePrintTo,
  bleScanStart,
  encodeText,
  isNative,
  sppDiscover,
  sppPrint,
  webBluetoothPick,
  webBluetoothPrintTo,
  type BleScanDevice,
  type PrinterConn,
  type PrinterEncoding,
  type SppDevice
} from "@/utils/printer";
import {
  printers,
  removePrinter,
  savePrinter,
  updatePrinter,
  type SavedPrinter
} from "@/utils/printerStore";

const $q = useQuasar();

const TOKEN_KEY = "lfdd_token";

// ================= 类型 =================
interface SidebarMenu {
  id: string;
  name: string;
}

interface AdminGoods {
  id: string;
  menu_id: string;
  name: string;
  price: number; // 分
  spec: string;
  stock: number;
  image: string;
  sort: number;
  status: number;
}

// ================= 通用 =================
function describeError(e: unknown) {
  return e instanceof Error ? e.message : String(e);
}

// ================= 认证 =================
const token = ref(localStorage.getItem(TOKEN_KEY) ?? "");
const authTab = ref("login");
const account = ref("");
const regEmail = ref("");
const password = ref("");
const authBusy = ref(false);
const authError = ref("");

function setToken(t: string) {
  token.value = t;
  localStorage.setItem(TOKEN_KEY, t);
}

// 统一请求：自动带 Bearer token，401 时清理登录态
// eslint-disable-next-line @typescript-eslint/no-explicit-any
async function api(path: string, opts: { method?: string; body?: unknown } = {}): Promise<any> {
  const res = await fetch(path, {
    method: opts.method ?? "GET",
    headers: {
      "Content-Type": "application/json",
      ...(token.value ? { Authorization: `Bearer ${token.value}` } : {})
    },
    ...(opts.body !== undefined ? { body: JSON.stringify(opts.body) } : {})
  });

  if (res.status === 401) {
    token.value = "";
    localStorage.removeItem(TOKEN_KEY);
    throw new Error("登录已过期，请重新登录");
  }

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.message || `请求失败（${res.status}）`);
  }
  return data;
}

async function submitAuth() {
  authError.value = "";
  authBusy.value = true;
  try {
    const isRegister = authTab.value === "register";
    const data = await api(isRegister ? "/api/auth/register" : "/api/auth/login", {
      method: "POST",
      body: isRegister
        ? { email: regEmail.value.trim(), password: password.value }
        : { account: account.value.trim(), password: password.value }
    });
    setToken(data.token);
    password.value = "";
    await loadMenus();
    await loadGoods();
  } catch (e) {
    authError.value = describeError(e);
  } finally {
    authBusy.value = false;
  }
}

async function logout() {
  try {
    await api("/api/auth/logout", { method: "POST" });
  } catch {
    // token 已失效也照常退出
  }
  token.value = "";
  localStorage.removeItem(TOKEN_KEY);
  menus.value = [];
  goodsList.value = [];
}

// ================= 分类 =================
const menus = ref<SidebarMenu[]>([]);
const menuDlg = ref(false);
const menuSaving = ref(false);
const menuForm = reactive({ id: "", name: "" });

const menuOptions = computed(() => menus.value.map((m) => ({ label: m.name, value: m.id })));
const menuFilterOptions = computed(() => [
  { label: "全部分类", value: "0" },
  ...menuOptions.value
]);

function menuName(id: string) {
  return menus.value.find((m) => m.id === id)?.name ?? "-";
}

async function loadMenus() {
  const data = await api("/api/goods/sidebar");
  menus.value = Array.isArray(data.menus) ? data.menus : [];
}

function openMenu(m?: SidebarMenu) {
  menuForm.id = m?.id ?? "";
  menuForm.name = m?.name ?? "";
  menuDlg.value = true;
}

async function saveMenu() {
  const name = menuForm.name.trim();
  if (!name) {
    $q.notify({ message: "请输入分类名称", color: "negative" });
    return;
  }
  menuSaving.value = true;
  try {
    if (menuForm.id) {
      await api("/api/menu/update", { method: "POST", body: { id: menuForm.id, name } });
    } else {
      await api("/api/menu/create", { method: "POST", body: { name, sort: menus.value.length } });
    }
    $q.notify({ message: "保存成功", color: "positive" });
    menuDlg.value = false;
    await loadMenus();
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    menuSaving.value = false;
  }
}

function removeMenu(m: SidebarMenu) {
  $q.dialog({
    title: "删除分类",
    message: `确定删除「${m.name}」？分类下有商品时会被后端拒绝。`,
    cancel: true
  }).onOk(async () => {
    try {
      await api("/api/menu/delete", { method: "POST", body: { id: m.id } });
      $q.notify({ message: "已删除", color: "positive" });
      await loadMenus();
      if (filterMenu.value === m.id) filterMenu.value = "0";
      await loadGoods();
    } catch (e) {
      $q.notify({ message: describeError(e), color: "negative" });
    }
  });
}

// ================= 分类拖拽排序 =================
const dragMenuIdx = ref<number | null>(null);

function onMenuDragStart(idx: number) {
  dragMenuIdx.value = idx;
}

async function onMenuDrop(idx: number) {
  const from = dragMenuIdx.value;
  dragMenuIdx.value = null;
  if (from === null || from === idx) return;
  const ids = menus.value.map((m) => m.id);
  const moved = ids.splice(from, 1)[0]!;
  ids.splice(idx, 0, moved);
  try {
    await api("/api/menu/sort", { method: "POST", body: { ids } });
    await loadMenus();
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  }
}

// ================= 商品 =================
const tab = ref("goods");
const filterMenu = ref("0");
const goodsList = ref<AdminGoods[]>([]);
const goodsLoading = ref(false);
const goodsDlg = ref(false);
const goodsSaving = ref(false);
const goodsForm = reactive({
  id: "",
  menu_id: "",
  name: "",
  priceYuan: "0",
  spec: "",
  stock: "-1",
  image: "",
  statusOn: true,
  sort: 0
});

// 图片上传：选择文件即传给后端（内容哈希命名），成功后把返回哈希填进 goodsForm.image
const goodsImageFile = ref<File | null>(null);
const imgUploading = ref(false);

// image 字段只存哈希，展示时拼成 /api/upload/{hash}（兼容旧数据里的完整 URL/路径）
function imgSrc(v: string): string {
  if (!v) return "";
  if (v.startsWith("/") || /^https?:\/\//.test(v)) return v;
  return `/api/upload/${v}`;
}

async function uploadGoodsImage(f: unknown) {
  if (!f) return;
  imgUploading.value = true;
  try {
    const fd = new FormData();
    fd.append("file", f as File);
    const res = await fetch("/api/goods/upload", {
      method: "POST",
      ...(token.value ? { headers: { Authorization: `Bearer ${token.value}` } } : {}),
      body: fd
    });
    if (res.status === 401) {
      token.value = "";
      localStorage.removeItem(TOKEN_KEY);
      throw new Error("登录已过期，请重新登录");
    }
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.message || `上传失败（${res.status}）`);
    goodsForm.image = data.hash;
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    imgUploading.value = false;
    goodsImageFile.value = null; // 清掉以便再次选同一文件也能触发
  }
}

async function loadGoods() {
  goodsLoading.value = true;
  try {
    const params = new URLSearchParams({ all: "1", page: "1", page_size: "100" });
    if (filterMenu.value !== "0") params.set("menu_id", filterMenu.value);
    const data = await api(`/api/goods/page?${params.toString()}`);
    goodsList.value = Array.isArray(data.list) ? data.list : [];
  } finally {
    goodsLoading.value = false;
  }
}

// 切换分类筛选
async function setFilter(v: string | number) {
  const val = String(v);
  if (String(filterMenu.value) === val) return;
  filterMenu.value = val;
  await loadGoods();
}

// ================= 商品拖拽排序 =================
// 后端 sort 是全局序列（列表按 sort ASC 排序），拖拽只提交当前可见列表的 id 顺序，
// 后端按下标 0..n 重写；分类筛选时只影响本分类，「全部分类」时整体重排
const dragGoodsIdx = ref<number | null>(null);

function onGoodsDragStart(idx: number) {
  dragGoodsIdx.value = idx;
}

async function onGoodsDrop(idx: number) {
  const from = dragGoodsIdx.value;
  dragGoodsIdx.value = null;
  if (from === null || from === idx) return;
  const ids = goodsList.value.map((g) => g.id);
  const moved = ids.splice(from, 1)[0]!;
  ids.splice(idx, 0, moved);
  // 乐观更新：先本地换序，保存成功后刷新
  const list = [...goodsList.value];
  const g = list.splice(from, 1)[0]!;
  list.splice(idx, 0, g);
  goodsList.value = list;
  try {
    await api("/api/goods/sort", { method: "POST", body: { ids } });
    await loadGoods();
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
    await loadGoods();
  }
}

function openGoods(g?: AdminGoods) {
  if (menus.value.length === 0) {
    $q.notify({ message: "请先创建分类", color: "negative" });
    return;
  }
  if (g) {
    goodsForm.id = g.id;
    goodsForm.menu_id = g.menu_id;
    goodsForm.name = g.name;
    goodsForm.priceYuan = (g.price / 100).toFixed(2);
    goodsForm.spec = g.spec;
    goodsForm.stock = String(g.stock);
    goodsForm.image = g.image;
    goodsForm.statusOn = g.status === 1;
    goodsForm.sort = g.sort;
  } else {
    goodsForm.id = "";
    goodsForm.menu_id = filterMenu.value !== "0" ? filterMenu.value : menus.value[0]!.id;
    goodsForm.name = "";
    goodsForm.priceYuan = "0";
    goodsForm.spec = "";
    goodsForm.stock = "-1";
    goodsForm.image = "";
    goodsForm.statusOn = true;
    goodsForm.sort = 0;
  }
  goodsImageFile.value = null;
  goodsDlg.value = true;
}

async function saveGoods() {
  const name = goodsForm.name.trim();
  const price = Math.round(parseFloat(goodsForm.priceYuan || "0") * 100);
  const stock = parseInt(goodsForm.stock, 10);

  if (!name) {
    $q.notify({ message: "请输入商品名称", color: "negative" });
    return;
  }
  if (!goodsForm.menu_id) {
    $q.notify({ message: "请选择所属分类", color: "negative" });
    return;
  }
  if (!Number.isFinite(price) || price < 0) {
    $q.notify({ message: "价格无效", color: "negative" });
    return;
  }
  if (!Number.isInteger(stock)) {
    $q.notify({ message: "库存需为整数（-1 = 不限）", color: "negative" });
    return;
  }

  goodsSaving.value = true;
  try {
    const body = {
      menu_id: goodsForm.menu_id,
      name,
      price,
      spec: goodsForm.spec.trim(),
      stock,
      image: goodsForm.image.trim(),
      status: goodsForm.statusOn ? 1 : 0,
      sort: goodsForm.sort
    };
    if (goodsForm.id) {
      await api("/api/goods/update", { method: "POST", body: { id: goodsForm.id, ...body } });
    } else {
      await api("/api/goods/create", { method: "POST", body });
    }
    $q.notify({ message: "保存成功", color: "positive" });
    goodsDlg.value = false;
    await loadGoods();
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    goodsSaving.value = false;
  }
}

async function toggleStatus(g: AdminGoods, on: boolean) {
  try {
    await api("/api/goods/update", {
      method: "POST",
      body: {
        id: g.id,
        menu_id: g.menu_id,
        name: g.name,
        price: g.price,
        spec: g.spec,
        stock: g.stock,
        image: g.image,
        status: on ? 1 : 0,
        sort: g.sort
      }
    });
    g.status = on ? 1 : 0;
    $q.notify({ message: on ? "已上架" : "已下架", color: "positive" });
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
    await loadGoods();
  }
}

function removeGoods(g: AdminGoods) {
  $q.dialog({
    title: "删除商品",
    message: `确定删除「${g.name}」？`,
    cancel: true
  }).onOk(async () => {
    try {
      await api("/api/goods/delete", { method: "POST", body: { id: g.id } });
      $q.notify({ message: "已删除", color: "positive" });
      await loadGoods();
    } catch (e) {
      $q.notify({ message: describeError(e), color: "negative" });
    }
  });
}

// ================= 蓝牙搜索（双列：低功耗 / 传统） =================
const scanning = ref(false);
const status = ref("");
const statusOk = ref(false);
const bleDevices = ref<BleScanDevice[]>([]);
const sppDevices = ref<SppDevice[]>([]);
let stopBle: (() => Promise<void>) | null = null;

// ================= 打印机设备管理 =================
const encOptions = [
  { label: "GBK（推荐）", value: "gbk" },
  { label: "UTF-8", value: "utf8" }
];
const devDlg = ref(false);
const devForm = reactive({
  id: "",
  name: "",
  conn: "ble" as PrinterConn,
  encoding: "gbk" as PrinterEncoding,
  editing: false
});

// 组装测试小票（按设备编码格式）
function buildTestPayload(enc: PrinterEncoding = "gbk") {
  const now = new Date();
  const pad = (n: number) => String(n).padStart(2, "0");
  const lines = [
    "--------------------------------",
    "你好 蓝牙小票机",
    `当前时间：${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ` +
      `${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`,
    "--------------------------------"
  ];
  // ESC/POS：初始化 + 居中放大 + 文本 + 走纸切纸
  const parts: Uint8Array[] = [
    new Uint8Array([0x1b, 0x40]), // ESC @ 初始化
    new Uint8Array([0x1b, 0x61, 0x01]), // ESC a 1 居中
    new Uint8Array([0x1d, 0x21, 0x11]), // GS ! 双倍宽高
    encodeText(lines.join("\n"), enc),
    new Uint8Array([0x1b, 0x61, 0x00]), // ESC a 0 左对齐
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

// 扫描列表点设备 → 弹添加对话框（名称预填，可选编码）
function openAdd(d: BleScanDevice | SppDevice, conn: PrinterConn) {
  devForm.editing = false;
  devForm.conn = conn;
  devForm.id = conn === "ble" ? (d as BleScanDevice).deviceId : (d as SppDevice).address;
  devForm.name = d.name || "小票机";
  devForm.encoding = "gbk";
  devDlg.value = true;
}

// 编辑已保存设备（改名 / 换编码）
function openEdit(p: SavedPrinter) {
  devForm.editing = true;
  devForm.id = p.id;
  devForm.name = p.name;
  devForm.conn = p.conn;
  devForm.encoding = p.encoding;
  devDlg.value = true;
}

function saveDev() {
  const name = devForm.name.trim() || "小票机";
  if (devForm.editing) {
    updatePrinter(devForm.id, devForm.conn, { name, encoding: devForm.encoding });
  } else {
    savePrinter({ id: devForm.id, name, conn: devForm.conn, encoding: devForm.encoding });
  }
  devDlg.value = false;
  $q.notify({ message: "已保存打印机", color: "positive" });
}

function confirmRemovePrinter(p: SavedPrinter) {
  $q.dialog({
    title: "删除打印机",
    message: `确定删除「${p.name}」？`,
    cancel: true
  }).onOk(() => removePrinter(p.id, p.conn));
}

// 测试打印：按设备连接方式与编码发测试小票
async function testPrinter(p: SavedPrinter) {
  try {
    const payload = buildTestPayload(p.encoding);
    const name =
      p.conn === "spp"
        ? await sppPrint({ address: p.id, name: p.name }, payload)
        : isNative()
          ? await blePrintTo(p.id, payload, p.name)
          : await webBluetoothPrintTo(p.id, payload);
    statusOk.value = true;
    status.value = `已连接「${name}」，测试小票已发送，请查看打印机出纸`;
  } catch (e) {
    statusOk.value = false;
    status.value = describeError(e);
  }
}

// 开始搜索：
// - 原生：BLE 实时扫描 + 传统蓝牙 discoverUnpaired（仅未配对设备，不返回已保存的）
// - 网页：Web Bluetooth 没有被动扫描 API，弹系统设备选择框（勾选即"搜索"）
async function startScan() {
  status.value = "";
  bleDevices.value = [];
  sppDevices.value = [];
  scanning.value = true;
  try {
    if (isNative()) {
      // 初始化 BLE（触发 Android「附近设备」运行时权限，扫描也依赖它）
      try {
        await bleInit();
      } catch {
        // 已初始化过则忽略
      }
      stopBle = await bleScanStart((d) => {
        if (!bleDevices.value.some((x) => x.deviceId === d.deviceId)) {
          bleDevices.value.push(d);
        }
      });

      // 传统蓝牙扫描（约 12 秒，结束后 resolve）
      // 单独 catch：权限被拒等错误只影响传统蓝牙一列，不打断 BLE 搜索
      try {
        const found = await sppDiscover((d) => {
          if (!sppDevices.value.some((x) => x.address === d.address)) {
            sppDevices.value.push(d);
          }
        });
        if (found.length === 0 && bleDevices.value.length === 0 && status.value === "") {
          statusOk.value = false;
          status.value = "未搜索到设备，请确认打印机已开机并进入配对/广播模式";
        }
      } catch (e) {
        const msg = describeError(e);
        statusOk.value = false;
        status.value = msg.includes("Location permission")
          ? "搜索传统蓝牙需要位置权限：请在系统设置 → 应用 → 权限中允许「位置」后重试（低功耗蓝牙不受影响）"
          : `传统蓝牙搜索失败：${msg}`;
      }
    } else {
      // 网页环境：Web Bluetooth 设备选择框（列出附近可见的 BLE 设备）
      try {
        const d = await webBluetoothPick();
        if (d.deviceId && !bleDevices.value.some((x) => x.deviceId === d.deviceId)) {
          bleDevices.value.push(d);
        }
        statusOk.value = true;
        status.value = `已选择「${d.name || "未命名设备"}」，点击列表中的设备添加保存`;
      } catch (e) {
        statusOk.value = false;
        status.value = describeError(e);
      }
    }
  } catch (e) {
    statusOk.value = false;
    status.value = describeError(e);
  } finally {
    scanning.value = false;
    // BLE 扫描由插件持续进行，直到点「停止」或离开页面
  }
}

// 停止搜索
async function stopScan() {
  scanning.value = false;
  const fn = stopBle;
  stopBle = null;
  if (fn) await fn().catch(() => undefined);
}

// ================= 初始化 =================
onMounted(async () => {
  if (!token.value) return;
  try {
    await loadMenus();
    await loadGoods();
  } catch (e) {
    // 401 时 token 已被清理，静默回到登录页；其他错误提示
    if (token.value) {
      $q.notify({ message: describeError(e), color: "negative" });
    }
  }
});
</script>

<style scoped lang="scss">
.admin-wrap {
  max-width: 960px;
  margin: 0 auto;
}

.goods-thumb {
  width: 40px;
  height: 40px;
  border-radius: 6px;
  background: #f5f5f5;
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

.scan-list {
  max-height: calc(100vh - 320px);
  overflow-y: auto;
}

.menu-row {
  cursor: grab;
  &:active {
    cursor: grabbing;
  }
}
</style>
