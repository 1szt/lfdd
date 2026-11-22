<template>
  <q-page class="q-pa-md">
    <!-- ================= 未登录：登录（注册已关闭，初始账号见服务端启动日志） ================= -->
    <div v-if="!token" class="flex flex-center" style="min-height: 60vh">
      <q-card style="width: 360px; max-width: 95vw">
        <q-card-section class="text-h6">登录</q-card-section>
        <q-card-section class="column q-gutter-sm q-pt-none">
          <q-input
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
            @keyup.enter="submitAuth"
          />
          <div v-if="authError" class="text-negative text-caption">{{
            authError
          }}</div>
          <q-btn
            color="primary"
            label="登录"
            :loading="authBusy"
            @click="submitAuth"
          />
          <div class="text-caption text-grey-6">
            初始账号与密码见服务端启动日志
          </div>
        </q-card-section>
      </q-card>
    </div>

    <!-- ================= 已登录：管理界面 ================= -->
    <div v-else class="admin-wrap">
      <div class="row items-center q-mb-sm">
        <div class="text-h6">管理</div>
        <q-space />
        <q-btn
          flat
          dense
          color="negative"
          icon="logout"
          label="退出登录"
          @click="logout"
        />
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
        <q-tab
          v-if="isAdmin"
          name="goods"
          label="商品管理"
          icon="inventory_2"
        />
        <q-tab v-if="isAdmin" name="menus" label="分类管理" icon="category" />
        <q-tab name="print" label="设备管理" icon="print" />
        <q-tab v-if="isAdmin" name="users" label="用户管理" icon="people" />
        <q-tab v-if="isAdmin" name="ai" label="AI 设置" icon="smart_toy" />
      </q-tabs>
      <q-separator />

      <q-tab-panels v-model="tab" animated class="bg-transparent">
        <!-- ============ 商品管理 ============ -->
        <q-tab-panel name="goods" class="q-pa-none q-pt-sm">
          <!-- 分类筛选：横向点选，替代下拉框 -->
          <div
            class="row no-wrap q-gutter-xs q-px-none"
            style="overflow-x: auto"
          >
            <q-btn
              v-for="opt in menuFilterOptions"
              :key="opt.value"
              dense
              unelevated
              size="sm"
              :label="opt.label"
              :color="
                String(filterMenu) === String(opt.value) ? 'primary' : 'grey-3'
              "
              :text-color="
                String(filterMenu) === String(opt.value) ? 'white' : 'grey-8'
              "
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

          <div v-else-if="goodsLoading" class="text-grey-6 q-pa-md"
            >加载中…</div
          >

          <div v-else-if="goodsList.length === 0" class="text-grey-6 q-pa-md">
            该分类下暂无商品，点右上角「新建商品」添加。
          </div>

          <q-list v-else separator>
            <q-item
              v-for="(g, gIdx) in goodsList"
              :key="g.id"
              class="menu-row goods-row"
              draggable="true"
              @dragstart="onGoodsDragStart(gIdx)"
              @dragover.prevent
              @drop="onGoodsDrop(gIdx)"
              @dragend="dragGoodsIdx = null"
            >
              <q-item-section avatar class="goods-photo">
                <div class="goods-thumb">
                  <img v-if="g.image" :src="imgSrc(g.image)" :alt="g.name" />
                  <q-icon
                    v-else
                    name="shopping_bag"
                    size="22px"
                    color="grey-5"
                  />
                </div>
              </q-item-section>
              <q-item-section class="goods-body">
                <!-- 上：规格信息 -->
                <div class="goods-spec">
                  <div class="goods-head">
                    <div class="goods-name text-weight-medium">{{
                      g.name
                    }}</div>
                    <div class="goods-price text-negative text-weight-bold">
                      ¥{{ (g.price / 100).toFixed(2) }}
                    </div>
                  </div>
                  <div class="goods-meta text-caption text-grey-7">
                    {{ menuName(g.menu_id) }}
                    <span v-if="g.spec"> · {{ g.spec }}</span>
                    <span v-if="g.alias"> · 别名：{{ g.alias }}</span>
                    · 库存 {{ g.stock === -1 ? "不限" : g.stock }}
                  </div>
                </div>
                <!-- 下：操作 -->
                <div class="goods-actions">
                  <q-toggle
                    :model-value="g.status === 1"
                    :label="g.status === 1 ? '上架' : '下架'"
                    dense
                    color="positive"
                    @update:model-value="(v: boolean) => toggleStatus(g, v)"
                  />
                  <q-space />
                  <q-btn dense flat round icon="edit" @click="openGoods(g)" />
                  <q-btn
                    dense
                    flat
                    round
                    icon="delete"
                    color="negative"
                    @click="removeGoods(g)"
                  />
                </div>
              </q-item-section>
            </q-item>
          </q-list>
        </q-tab-panel>

        <!-- ============ 分类管理 ============ -->
        <q-tab-panel name="menus" class="q-pa-none q-pt-sm">
          <q-toolbar class="q-px-none">
            <div class="text-grey-7 text-caption"
              >拖动分类行调整前台侧边栏顺序</div
            >
            <q-space />
            <q-btn
              color="primary"
              icon="add"
              label="新建分类"
              @click="openMenu()"
            />
          </q-toolbar>

          <div v-if="menus.length === 0" class="text-grey-6 q-pa-md"
            >暂无分类。</div
          >

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
                  <q-btn
                    dense
                    flat
                    round
                    icon="delete"
                    color="negative"
                    @click="removeMenu(m)"
                  />
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
              icon="add"
              label="添加设备"
              @click="openAddManual"
            />
            <q-btn
              class="q-ml-sm"
              outline
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
            <span class="text-caption text-grey-6"
              >点已保存设备可编辑 / 测试 / 删除</span
            >
          </q-toolbar>

          <q-banner
            v-if="status"
            dense
            rounded
            class="q-mb-sm"
            :class="
              statusOk ? 'bg-green-1 text-green-9' : 'bg-red-1 text-red-9'
            "
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
                还没有添加打印机，点上方「添加设备」或搜索后点击下方设备添加
              </q-item-section>
            </q-item>
            <q-item v-for="p in printers" :key="p.key">
              <q-item-section avatar>
                <q-icon
                  :name="
                    p.kind === 'label'
                      ? 'label'
                      : p.conn === 'spp'
                        ? 'settings_input_antenna'
                        : 'bluetooth'
                  "
                  :color="
                    p.kind === 'label'
                      ? 'deep-purple'
                      : p.conn === 'spp'
                        ? 'teal'
                        : 'primary'
                  "
                />
              </q-item-section>
              <q-item-section>
                <q-item-label>{{ p.name }}</q-item-label>
                <q-item-label caption>
                  {{ p.kind === "label" ? "标签打印机" : "小票机" }} ·
                  {{
                    p.transport === "wifi"
                      ? "WiFi"
                      : p.conn === "spp"
                        ? "传统蓝牙"
                        : "低功耗"
                  }}
                  · {{ p.id }} · {{ p.encoding === "utf8" ? "UTF-8" : "GBK" }}
                </q-item-label>
              </q-item-section>
              <q-item-section side>
                <div class="row no-wrap q-gutter-xs">
                  <q-btn
                    dense
                    flat
                    round
                    icon="edit"
                    aria-label="编辑打印机"
                    @click="openEdit(p)"
                  />
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
                    {{
                      scanning ? "搜索中，等待附近设备广播…" : "未搜索到设备"
                    }}
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

        <!-- ============ 用户管理 ============ -->
        <q-tab-panel name="users" class="q-pa-none q-pt-sm">
          <!-- 工具栏：手机上搜索框独占一行，按钮换行到第二行 -->
          <div class="row q-gutter-sm items-center q-mb-xs full-width">
            <q-input
              v-model="userKeyword"
              dense
              outlined
              class="col-12 col-sm-5"
              placeholder="搜索用户名 / 邮箱 / 姓名"
              clearable
              @keyup.enter="loadUsers(1)"
            >
              <template #append>
                <q-icon
                  name="search"
                  class="cursor-pointer"
                  @click="loadUsers(1)"
                />
              </template>
            </q-input>
            <div class="row no-wrap q-gutter-sm items-center">
              <q-btn
                flat
                dense
                icon="refresh"
                label="刷新"
                @click="loadUsers()"
              />
              <q-btn
                color="primary"
                dense
                icon="person_add"
                label="添加用户"
                unelevated
                @click="openAddUser"
              />
              <span class="text-caption text-grey-6 gt-xs"
                >共 {{ userTotal }} 位用户</span
              >
            </div>
          </div>

          <div v-if="usersLoading" class="text-grey-6 q-pa-md">加载中…</div>

          <div v-else-if="users.length === 0" class="text-grey-6 q-pa-md"
            >没有匹配的用户。</div
          >

          <q-list v-else separator>
            <q-item v-for="u in users" :key="u.id">
              <q-item-section avatar>
                <q-avatar color="primary" text-color="white" size="38px">
                  {{ (u.name || u.username).charAt(0).toUpperCase() }}
                </q-avatar>
              </q-item-section>
              <q-item-section>
                <q-item-label class="text-weight-medium ellipsis">
                  {{ u.name || u.username }}
                  <q-badge
                    v-if="u.id === meId"
                    color="orange"
                    label="我"
                    class="q-ml-xs"
                  />
                  <q-badge
                    :color="u.role === 2 ? 'purple' : 'grey-5'"
                    :label="u.role === 2 ? '管理员' : '普通用户'"
                    class="q-ml-xs"
                  />
                </q-item-label>
                <q-item-label caption class="ellipsis">
                  {{ u.username }} · {{ u.email }}
                  <span v-if="u.last_login_at" class="gt-xs">
                    · 最近登录 {{ fmtTime(u.last_login_at) }}
                  </span>
                </q-item-label>
              </q-item-section>
              <!-- 操作区：竖排两行图标，避免手机上一行排不下 -->
              <q-item-section side>
                <div class="column items-end q-gutter-xs">
                  <div class="row no-wrap items-center q-gutter-xs">
                    <q-toggle
                      :model-value="u.status === 1"
                      dense
                      :color="u.status === 1 ? 'positive' : 'negative'"
                      :title="u.status === 1 ? '点击禁用' : '点击启用'"
                      @update:model-value="
                        (v: boolean) => toggleUserStatus(u, v)
                      "
                    />
                    <q-btn
                      dense
                      flat
                      round
                      icon="admin_panel_settings"
                      :color="u.role === 2 ? 'purple' : 'grey-6'"
                      :title="
                        u.role === 2
                          ? '当前是管理员，点击设为普通'
                          : '点击设为管理员'
                      "
                      @click="toggleUserRole(u)"
                    />
                  </div>
                  <div class="row no-wrap items-center q-gutter-xs">
                    <q-btn
                      dense
                      flat
                      round
                      icon="edit"
                      title="编辑用户名 / 密码"
                      @click="openEditUser(u)"
                    />
                    <q-btn
                      dense
                      flat
                      round
                      icon="devices"
                      title="设备管理"
                      @click="openUserDevices(u)"
                    />
                    <q-btn
                      dense
                      flat
                      round
                      icon="delete"
                      color="negative"
                      title="删除用户"
                      :disable="u.id === meId"
                      @click="removeUser(u)"
                    />
                  </div>
                </div>
              </q-item-section>
            </q-item>
          </q-list>

          <!-- 分页 -->
          <div
            v-if="userTotal > userPageSize"
            class="row justify-center q-gutter-sm q-mt-sm"
          >
            <q-btn
              dense
              flat
              icon="chevron_left"
              :disable="userPage <= 1"
              @click="loadUsers(userPage - 1)"
            />
            <span class="text-caption self-center">
              {{ userPage }} / {{ Math.ceil(userTotal / userPageSize) }}
            </span>
            <q-btn
              dense
              flat
              icon="chevron_right"
              :disable="userPage >= Math.ceil(userTotal / userPageSize)"
              @click="loadUsers(userPage + 1)"
            />
          </div>

          <!-- ============ 添加用户 ============ -->
          <q-dialog v-model="userAddDlg">
            <q-card style="width: 420px; max-width: 95vw">
              <q-card-section class="text-h6">添加用户</q-card-section>
              <q-card-section class="column q-gutter-sm q-pt-none">
                <q-input
                  v-model="addForm.username"
                  outlined
                  dense
                  label="用户名"
                  autofocus
                />
                <q-input
                  v-model="addForm.email"
                  outlined
                  dense
                  label="邮箱"
                  type="email"
                />
                <q-input
                  v-model="addForm.password"
                  outlined
                  dense
                  label="密码"
                  type="password"
                  hint="至少 8 位"
                />
                <q-select
                  v-model="addForm.role"
                  :options="roleOptions"
                  outlined
                  dense
                  label="角色"
                  emit-value
                  map-options
                />
              </q-card-section>
              <q-card-actions align="right">
                <q-btn flat label="取消" v-close-popup />
                <q-btn
                  color="primary"
                  label="创建"
                  :loading="userAddBusy"
                  @click="saveAddUser"
                />
              </q-card-actions>
            </q-card>
          </q-dialog>

          <!-- ============ 编辑用户（用户名 / 邮箱 / 密码） ============ -->
          <q-dialog v-model="userEditDlg">
            <q-card style="width: 420px; max-width: 95vw">
              <q-card-section class="text-h6">编辑用户</q-card-section>
              <q-card-section class="column q-gutter-sm q-pt-none">
                <q-input
                  v-model="editForm.username"
                  outlined
                  dense
                  label="用户名"
                  autofocus
                />
                <q-input
                  v-model="editForm.email"
                  outlined
                  dense
                  type="email"
                  label="邮箱"
                  hint="邮箱也可用于登录"
                />
                <q-input
                  v-model="editForm.password"
                  outlined
                  dense
                  label="新密码"
                  type="password"
                  hint="留空则不修改密码；修改后该用户的其他登录会失效"
                />
              </q-card-section>
              <q-card-actions align="right">
                <q-btn flat label="取消" v-close-popup />
                <q-btn
                  color="primary"
                  label="保存"
                  :loading="userEditBusy"
                  @click="saveEditUser"
                />
              </q-card-actions>
            </q-card>
          </q-dialog>

          <!-- ============ 设备管理（登录会话） ============ -->
          <q-dialog v-model="userDevDlg">
            <q-card style="width: 520px; max-width: 95vw">
              <q-card-section class="text-h6">
                设备管理 · {{ devUser ? devUser.username : "" }}
              </q-card-section>
              <q-card-section class="q-pt-none">
                <div v-if="devLoading" class="text-grey-6">加载中…</div>
                <div v-else-if="devSessions.length === 0" class="text-grey-6">
                  该用户暂无登录会话。
                </div>
                <q-list v-else separator>
                  <q-item v-for="s in devSessions" :key="s.token">
                    <q-item-section>
                      <q-item-label>
                        {{ s.device || "未知设备" }}
                        <q-badge
                          v-if="s.current"
                          color="orange"
                          label="当前会话"
                          class="q-ml-xs"
                        />
                        <q-badge
                          v-if="s.expired"
                          color="grey-6"
                          label="已过期"
                          class="q-ml-xs"
                        />
                      </q-item-label>
                      <q-item-label caption>
                        IP {{ s.ip || "未知" }} · 登录
                        {{ fmtTime(s.created_at) }} · 过期
                        {{ fmtTime(s.expires_at) }}
                      </q-item-label>
                    </q-item-section>
                    <q-item-section side>
                      <q-btn
                        dense
                        flat
                        color="negative"
                        label="踢下线"
                        :disable="s.current || s.expired"
                        @click="kickSession(s)"
                      />
                    </q-item-section>
                  </q-item>
                </q-list>
              </q-card-section>
              <q-card-actions align="right">
                <q-btn flat label="关闭" v-close-popup />
              </q-card-actions>
            </q-card>
          </q-dialog>
        </q-tab-panel>

        <!-- ============ AI 设置 ============ -->
        <q-tab-panel name="ai" class="q-pa-none q-pt-sm">
          <div class="column q-gutter-md" style="max-width: 640px">
            <!-- 模型 API -->
            <q-card flat bordered>
              <q-card-section class="text-subtitle1">模型 API</q-card-section>
              <q-card-section class="column q-gutter-sm q-pt-none">
                <q-input
                  v-model="aiForm.base_url"
                  label="接口地址（OpenAI 兼容，不带 /chat/completions）"
                  outlined
                  dense
                  placeholder="https://api.deepseek.com/v1"
                />
                <q-input
                  v-model="aiForm.api_key"
                  label="API Key"
                  outlined
                  dense
                  type="password"
                  clearable
                  hint="请保密，不要泄露给他人"
                />
                <q-input
                  v-model="aiForm.model"
                  label="模型名"
                  outlined
                  dense
                  placeholder="deepseek-chat"
                />
                <!-- 思考模式：请求以 thinking.type = enabled/disabled 控制深度思考 -->
                <div class="row items-center q-gutter-sm q-pt-xs">
                  <q-toggle
                    v-model="aiForm.thinking"
                    label="思考模式"
                    color="primary"
                    dense
                  />
                  <q-select
                    v-model="aiForm.reasoning_effort"
                    :options="effortOptions"
                    outlined
                    dense
                    label="思考强度"
                    emit-value
                    map-options
                    :disable="!aiForm.thinking"
                    style="min-width: 150px"
                  />
                  <span class="text-caption text-grey-7">
                    开启时请求带 thinking.type=enabled 与
                    reasoning_effort；关闭时为 disabled
                  </span>
                </div>
                <div class="row q-gutter-sm">
                  <q-btn
                    color="primary"
                    label="保存"
                    unelevated
                    :loading="aiSaving"
                    @click="saveAi"
                  />
                  <q-btn
                    outline
                    color="primary"
                    label="测试连接"
                    :loading="aiTesting"
                    @click="testAi"
                  />
                </div>
                <div
                  v-if="aiTestResult"
                  class="text-caption"
                  :class="aiTestOk ? 'text-green-8' : 'text-negative'"
                >
                  {{ aiTestResult }}
                </div>
              </q-card-section>
            </q-card>

            <!-- 系统提示词 -->
            <q-card flat bordered>
              <q-card-section class="text-subtitle1"
                >AI 商品匹配 · 系统提示词</q-card-section
              >
              <q-card-section class="q-pt-none">
                <q-input
                  v-model="aiForm.system_prompt"
                  type="textarea"
                  outlined
                  autogrow
                  label="系统提示词"
                  hint='匹配点单文字时发给大模型的指令；要求只输出 JSON 数组 [{"id":"商品id","qty":数量}]'
                />

                <!-- 拼接后的完整提示词展示：system 提示词 + 每次点单拼接的用户消息 -->
                <div class="text-caption text-grey-7 q-mt-md">
                  拼接后的完整提示词（每次点单实际发送给大模型的内容，随上方编辑实时更新，保存后生效）
                </div>
                <q-input
                  :model-value="aiPromptPreview"
                  type="textarea"
                  readonly
                  outlined
                  autogrow
                  :input-style="{ fontFamily: 'monospace', fontSize: '13px' }"
                  class="q-mt-xs"
                />

                <q-btn
                  class="q-mt-sm"
                  color="primary"
                  label="保存提示词"
                  unelevated
                  :loading="aiSaving"
                  @click="saveAi"
                />
              </q-card-section>
            </q-card>
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
            <!-- 设备类型：标签打印占位 -->
            <q-select
              v-model="devForm.kind"
              :options="kindOptions"
              label="设备类型"
              outlined
              dense
              emit-value
              map-options
            />
            <!-- 连接方式：WiFi 打印占位 -->
            <q-select
              v-model="devForm.transport"
              :options="transportOptions"
              label="连接方式"
              outlined
              dense
              emit-value
              map-options
            />
            <!-- 蓝牙子类型（仅连接方式=蓝牙） -->
            <q-select
              v-if="devForm.transport === 'bt'"
              v-model="devForm.conn"
              :options="connOptions"
              label="蓝牙方式"
              outlined
              dense
              emit-value
              map-options
            />
            <q-input
              v-model="devForm.id"
              :label="
                devForm.transport === 'wifi'
                  ? '打印机地址（IP:端口）'
                  : '蓝牙地址'
              "
              outlined
              dense
              :placeholder="
                devForm.transport === 'wifi' ? '192.168.1.88:9100' : undefined
              "
              :hint="
                devForm.transport === 'wifi'
                  ? 'WiFi 打印即将支持，可先保存'
                  : undefined
              "
            />
            <q-input
              v-model="devForm.name"
              label="设备名称"
              outlined
              dense
              autofocus
            />
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
            <q-input
              v-model="goodsForm.name"
              label="名称"
              outlined
              dense
              autofocus
            />
            <q-input
              v-model="goodsForm.alias"
              label="别名（多个用逗号分隔，可空）"
              outlined
              dense
              hint="AI 点单匹配时名称与别名都会识别"
            />
            <q-input
              v-model="goodsForm.priceYuan"
              label="价格（元）"
              type="number"
              step="0.01"
              min="0"
              outlined
              dense
            />
            <q-input
              v-model="goodsForm.costYuan"
              label="进价（元）"
              type="number"
              step="0.01"
              min="0"
              outlined
              dense
            />
            <q-input
              v-model="goodsForm.spec"
              label="规格（如：大杯 / 500g，可空）"
              outlined
              dense
            />
            <q-input
              v-model="goodsForm.stock"
              label="库存"
              type="number"
              outlined
              dense
              hint="-1 = 不限库存，0 = 售罄"
            />
            <!-- 媒体：上传（哈希存储）与第三方直链（只存链接）并排，左右与上方字段对齐 -->
            <div class="row items-center">
              <q-file
                v-model="goodsMediaFiles"
                outlined
                dense
                multiple
                class="col"
                label="上传图片/视频（压缩为 WebP ≤1MB，视频原样保存）"
                accept="image/*,video/*"
                :loading="imgUploading"
                @update:model-value="uploadGoodsMedia"
              >
                <template #prepend>
                  <q-icon name="upload" />
                </template>
              </q-file>
              <q-btn
                dense
                outline
                color="primary"
                icon="photo_camera"
                label="拍照"
                no-caps
                class="media-link-btn q-ml-sm"
                :loading="imgUploading"
                @click="captureMedia"
              />
              <q-btn
                dense
                outline
                color="primary"
                icon="link"
                label="添加第三方链接"
                no-caps
                class="media-link-btn q-ml-sm"
                @click="promptMediaUrl"
              />
            </div>
            <div v-if="goodsForm.media.length" class="media-grid">
              <div
                v-for="(m, i) in goodsForm.media"
                :key="mediaKey(m)"
                class="media-item"
              >
                <img v-if="m.type === 'image'" :src="mediaSrc(m)" alt="" />
                <video
                  v-else-if="m.type === 'video'"
                  :src="mediaSrc(m)"
                  muted
                  playsinline
                  preload="metadata"
                />
                <div v-else class="media-web">
                  <q-icon name="language" size="26px" />
                </div>
                <q-btn
                  round
                  dense
                  size="xs"
                  color="negative"
                  icon="close"
                  class="media-remove"
                  @click="goodsForm.media.splice(i, 1)"
                />
                <div
                  v-if="m.type === 'image' && mediaKey(m) === firstMediaImage()"
                  class="media-tag"
                  >首页</div
                >
              </div>
            </div>
            <q-toggle
              v-model="goodsForm.statusOn"
              label="上架销售"
              color="positive"
            />
          </q-card-section>

          <q-card-actions align="right">
            <q-btn flat label="取消" v-close-popup />
            <q-btn
              color="primary"
              label="保存"
              :loading="goodsSaving"
              @click="saveGoods"
            />
          </q-card-actions>
        </q-card>
      </q-dialog>

      <!-- 第三方链接：手动粘贴或扫一扫二维码 -->
      <link-scan-dialog v-model="linkDlg" @ok="addMediaUrl" />

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
            <q-btn
              color="primary"
              label="保存"
              :loading="menuSaving"
              @click="saveMenu"
            />
          </q-card-actions>
        </q-card>
      </q-dialog>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useQuasar } from "quasar";
import { apiPath } from "@/utils/api";
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
  type PrinterKind,
  type PrinterTransport,
  type SavedPrinter
} from "@/utils/printerStore";

import {
  clearGuestIdentity,
  markLoggedIn,
  markLoggedOut,
  TOKEN_KEY
} from "@/utils/guest";
import LinkScanDialog from "@/components/LinkScanDialog.vue";

const $q = useQuasar();

// ================= 类型 =================
interface SidebarMenu {
  id: string;
  name: string;
}

interface AdminGoods {
  id: string;
  menu_id: string;
  name: string;
  alias: string; // 别名（多个用逗号分隔），AI 匹配时与名称一起识别
  price: number; // 分
  cost?: number; // 进价，分（仅管理员接口返回）
  spec: string;
  stock: number;
  image: string;
  media?: GoodsMedia[]; // 图片/视频列表
  sort: number;
  status: number;
}

interface GoodsMedia {
  hash?: string; // 本站上传：内容哈希
  url?: string; // 第三方：图片/视频直链或网页地址
  type: "image" | "video" | "web";
}

// ================= 通用 =================
function describeError(e: unknown) {
  return e instanceof Error ? e.message : String(e);
}

// ================= 认证 =================
const token = ref(localStorage.getItem(TOKEN_KEY) ?? "");
const account = ref("");
const password = ref("");
const authBusy = ref(false);
const authError = ref("");

function setToken(t: string) {
  token.value = t;
  localStorage.setItem(TOKEN_KEY, t);
  // 登录后忽略并清掉游客身份（后端登录时已作废该身份）
  clearGuestIdentity();
  markLoggedIn();
}

// 统一请求：自动带 Bearer token，401 时清理登录态
// eslint-disable-next-line @typescript-eslint/no-explicit-any
async function api(
  path: string,
  opts: { method?: string; body?: unknown } = {}
): Promise<any> {
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
    markLoggedOut();
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
    const data = await api("/api/auth/login", {
      method: "POST",
      body: { account: account.value.trim(), password: password.value }
    });
    setToken(data.token);
    password.value = "";
    await loadMe();
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
  markLoggedOut();
  menus.value = [];
  goodsList.value = [];
  isAdmin.value = false;
  meId.value = "";
  tab.value = "goods";
}

// ================= 分类 =================
const menus = ref<SidebarMenu[]>([]);
const menuDlg = ref(false);
const menuSaving = ref(false);
const menuForm = reactive({ id: "", name: "" });

const menuOptions = computed(() =>
  menus.value.map(m => ({ label: m.name, value: m.id }))
);
const menuFilterOptions = computed(() => [
  { label: "全部分类", value: "0" },
  ...menuOptions.value
]);

function menuName(id: string) {
  return menus.value.find(m => m.id === id)?.name ?? "-";
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
      await api("/api/menu/update", {
        method: "POST",
        body: { id: menuForm.id, name }
      });
    } else {
      await api("/api/menu/create", {
        method: "POST",
        body: { name, sort: menus.value.length }
      });
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
  const ids = menus.value.map(m => m.id);
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
  alias: "",
  priceYuan: "0",
  costYuan: "0",
  spec: "",
  stock: "-1",
  media: [] as GoodsMedia[],
  statusOn: true,
  sort: 0
});

// 媒体上传：选择文件即传给后端（内容哈希命名），成功后追加到 goodsForm.media
const goodsMediaFiles = ref<File[] | null>(null);
const imgUploading = ref(false);

// image 字段只存哈希，展示时拼成 /api/upload/{hash}（兼容旧数据里的完整 URL/路径）
function imgSrc(v: string): string {
  if (!v) return "";
  if (/^https?:\/\//.test(v)) return v;
  if (v.startsWith("/")) return apiPath(v);
  return apiPath(`/api/upload/${v}`);
}

// 媒体标识：本站哈希或第三方直链
function mediaKey(m: GoodsMedia): string {
  return m.hash || m.url || "";
}

// 媒体地址：本站哈希拼成 /api/upload/{hash}，第三方直链原样使用
function mediaSrc(m: GoodsMedia): string {
  return imgSrc(mediaKey(m));
}

// 第一张图片：首页展示用（哈希或直链）
function firstMediaImage(): string {
  const m = goodsForm.media.find(x => x.type === "image");
  return m ? mediaKey(m) : "";
}

// 拍照：唤起系统相机，拍完直接按上传流程压缩入库（无相机的桌面端退化为选择文件）
function captureMedia() {
  if (imgUploading.value) return;
  const input = document.createElement("input");
  input.type = "file";
  input.accept = "image/*";
  input.setAttribute("capture", "environment");
  input.style.display = "none";
  input.addEventListener("change", () => {
    const f = input.files?.[0];
    input.remove();
    if (f) uploadGoodsMedia([f]);
  });
  document.body.appendChild(input);
  input.click();
}

// ================= 第三方媒体链接 =================
// 链接弹窗（手动粘贴 / 扫一扫）：独立组件，避免在商品弹窗里再加输入控件
const linkDlg = ref(false);
function promptMediaUrl() {
  linkDlg.value = true;
}

// 第三方链接：只存链接，不下载；按扩展名判断图片/视频，其余按网页嵌入
function addMediaUrl(u: string) {
  if (!/^https?:\/\//i.test(u)) {
    $q.notify({ message: "请输入 http/https 开头的完整链接", color: "negative" });
    return;
  }
  const path = u.split(/[?#]/)[0]!.toLowerCase();
  const type: GoodsMedia["type"] = /\.(mp4|m4v|mov|webm)$/.test(path)
    ? "video"
    : /\.(jpe?g|png|gif|webp|avif|bmp)$/.test(path)
      ? "image"
      : "web";
  goodsForm.media.push({ url: u, type });
}

// 批量上传：逐张/逐段传到后端，按返回哈希追加进媒体列表
async function uploadGoodsMedia(files: unknown) {
  const targets = (Array.isArray(files) ? files : files ? [files] : []).filter(
    (f): f is File => f instanceof File
  );
  if (targets.length === 0) return;
  imgUploading.value = true;
  try {
    for (const f of targets) {
      const fd = new FormData();
      fd.append("file", f);
      const res = await fetch("/api/goods/upload", {
        method: "POST",
        ...(token.value
          ? { headers: { Authorization: `Bearer ${token.value}` } }
          : {}),
        body: fd
      });
      if (res.status === 401) {
        token.value = "";
        localStorage.removeItem(TOKEN_KEY);
        markLoggedOut();
        throw new Error("登录已过期，请重新登录");
      }
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.message || `上传失败（${res.status}）`);
      goodsForm.media.push({
        hash: data.hash,
        type: f.type.startsWith("video/") ? "video" : "image"
      });
    }
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    imgUploading.value = false;
    goodsMediaFiles.value = null; // 清掉以便再次选同一文件也能触发
  }
}

async function loadGoods() {
  goodsLoading.value = true;
  try {
    const params = new URLSearchParams({
      all: "1",
      page: "1",
      page_size: "100"
    });
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
  const ids = goodsList.value.map(g => g.id);
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
    goodsForm.alias = g.alias || "";
    goodsForm.priceYuan = (g.price / 100).toFixed(2);
    goodsForm.costYuan = ((g.cost ?? 0) / 100).toFixed(2);
    goodsForm.spec = g.spec;
    goodsForm.stock = String(g.stock);
    goodsForm.media =
      g.media && g.media.length
        ? g.media.map(m => ({ ...m }))
        : g.image
          ? [{ hash: g.image, type: "image" }]
          : [];
    goodsForm.statusOn = g.status === 1;
    goodsForm.sort = g.sort;
  } else {
    goodsForm.id = "";
    goodsForm.menu_id =
      filterMenu.value !== "0" ? filterMenu.value : menus.value[0]!.id;
    goodsForm.name = "";
    goodsForm.alias = "";
    goodsForm.priceYuan = "0";
    goodsForm.costYuan = "0";
    goodsForm.spec = "";
    goodsForm.stock = "-1";
    goodsForm.media = [];
    goodsForm.statusOn = true;
    goodsForm.sort = 0;
  }
  goodsMediaFiles.value = null;
  goodsDlg.value = true;
}

async function saveGoods() {
  const name = goodsForm.name.trim();
  const price = Math.round(parseFloat(goodsForm.priceYuan || "0") * 100);
  const cost = Math.round(parseFloat(goodsForm.costYuan || "0") * 100);
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
  if (!Number.isFinite(cost) || cost < 0) {
    $q.notify({ message: "进价无效", color: "negative" });
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
      alias: goodsForm.alias.trim(),
      price,
      cost,
      spec: goodsForm.spec.trim(),
      stock,
      media: goodsForm.media.map(m => ({ ...m })),
      image: firstMediaImage(),
      status: goodsForm.statusOn ? 1 : 0,
      sort: goodsForm.sort
    };
    if (goodsForm.id) {
      await api("/api/goods/update", {
        method: "POST",
        body: { id: goodsForm.id, ...body }
      });
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
        alias: g.alias,
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

// ================= 打印机设备管理（设备类型 / 连接方式表单，不去重） =================
const encOptions = [
  { label: "GBK（推荐）", value: "gbk" },
  { label: "UTF-8", value: "utf8" }
];
const kindOptions = [
  { label: "小票机", value: "receipt" },
  { label: "标签打印机（即将支持）", value: "label" }
];
const transportOptions = [
  { label: "蓝牙", value: "bt" },
  { label: "WiFi（即将支持）", value: "wifi" }
];
const connOptions = [
  { label: "低功耗蓝牙", value: "ble" },
  { label: "传统蓝牙", value: "spp" }
];
const devDlg = ref(false);
const devForm = reactive({
  id: "",
  name: "",
  conn: "ble" as PrinterConn,
  transport: "bt" as PrinterTransport,
  kind: "receipt" as PrinterKind,
  encoding: "gbk" as PrinterEncoding,
  editing: false,
  key: "" // 编辑目标（不去重，按 key 操作）
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

// 手动添加（扫描不到时手填地址）
function openAddManual() {
  devForm.editing = false;
  devForm.key = "";
  devForm.id = "";
  devForm.name = "";
  devForm.conn = "ble";
  devForm.transport = "bt";
  devForm.kind = "receipt";
  devForm.encoding = "gbk";
  devDlg.value = true;
}

// 扫描列表点设备 → 弹添加对话框（预填地址/名称/连接方式，每次保存都是新条目，不去重）
function openAdd(d: BleScanDevice | SppDevice, conn: PrinterConn) {
  devForm.editing = false;
  devForm.key = "";
  devForm.conn = conn;
  devForm.transport = "bt";
  devForm.kind = "receipt";
  devForm.id =
    conn === "ble" ? (d as BleScanDevice).deviceId : (d as SppDevice).address;
  devForm.name = d.name || "小票机";
  devForm.encoding = "gbk";
  devDlg.value = true;
}

// 编辑已保存设备（改名 / 换类型 / 换连接方式 / 换编码）
function openEdit(p: SavedPrinter) {
  devForm.editing = true;
  devForm.key = p.key;
  devForm.id = p.id;
  devForm.name = p.name;
  devForm.conn = p.conn;
  devForm.transport = p.transport;
  devForm.kind = p.kind;
  devForm.encoding = p.encoding;
  devDlg.value = true;
}

function saveDev() {
  const id = devForm.id.trim();
  if (!id) {
    $q.notify({ message: "请填写打印机地址", color: "negative" });
    return;
  }
  const name = devForm.name.trim() || "小票机";
  if (devForm.editing) {
    updatePrinter(devForm.key, {
      id,
      name,
      conn: devForm.conn,
      transport: devForm.transport,
      kind: devForm.kind,
      encoding: devForm.encoding
    });
  } else {
    savePrinter({
      id,
      name,
      conn: devForm.conn,
      transport: devForm.transport,
      kind: devForm.kind,
      encoding: devForm.encoding
    });
  }
  devDlg.value = false;
  $q.notify({ message: "已保存打印机", color: "positive" });
}

function confirmRemovePrinter(p: SavedPrinter) {
  $q.dialog({
    title: "删除打印机",
    message: `确定删除「${p.name}」？`,
    cancel: true
  }).onOk(() => removePrinter(p.key));
}

// 测试打印：按设备连接方式与编码发测试小票（WiFi / 标签为占位项，暂不支持）
async function testPrinter(p: SavedPrinter) {
  if (p.transport === "wifi") {
    statusOk.value = false;
    status.value = "WiFi 打印即将支持，暂无法测试";
    return;
  }
  if (p.kind === "label") {
    statusOk.value = false;
    status.value =
      "标签打印即将支持，暂无法测试；小票打印请把设备类型改为小票机";
    return;
  }
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
      stopBle = await bleScanStart(d => {
        if (!bleDevices.value.some(x => x.deviceId === d.deviceId)) {
          bleDevices.value.push(d);
        }
      });

      // 传统蓝牙扫描（约 12 秒，结束后 resolve）
      // 单独 catch：权限被拒等错误只影响传统蓝牙一列，不打断 BLE 搜索
      try {
        const found = await sppDiscover(d => {
          if (!sppDevices.value.some(x => x.address === d.address)) {
            sppDevices.value.push(d);
          }
        });
        if (
          found.length === 0 &&
          bleDevices.value.length === 0 &&
          status.value === ""
        ) {
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
        if (
          d.deviceId &&
          !bleDevices.value.some(x => x.deviceId === d.deviceId)
        ) {
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

// ================= 当前用户（角色决定是否展示管理页签） =================
interface MeInfo {
  id: string;
  username: string;
  name: string;
  role: number; // 1 普通 2 管理员
  status: number;
}

const meId = ref("");
const isAdmin = ref(false);

async function loadMe() {
  try {
    const d: MeInfo = await api("/api/user/me");
    meId.value = d.id;
    isAdmin.value = d.role === 2;
  } catch {
    isAdmin.value = false;
  }
  // 非管理员看不到商品/分类/用户/AI 页签，若正停在这些页签则落到「设备管理」
  if (!isAdmin.value && ["goods", "menus", "users", "ai"].includes(tab.value)) {
    tab.value = "print";
  }
}

// ================= 用户管理 =================
interface UserItem {
  id: string;
  username: string;
  email: string;
  name: string;
  status: number; // 1 正常 0 禁用
  role: number; // 1 普通 2 管理员
  last_login_at?: string;
}

const users = ref<UserItem[]>([]);
const usersLoading = ref(false);
const userKeyword = ref("");
const userPage = ref(1);
const userPageSize = 20;
const userTotal = ref(0);

async function loadUsers(page = userPage.value) {
  usersLoading.value = true;
  try {
    const params = new URLSearchParams({
      page: String(page),
      page_size: String(userPageSize)
    });
    const kw = userKeyword.value.trim();
    if (kw) params.set("keyword", kw);
    const d = await api(`/api/user/list?${params.toString()}`);
    users.value = Array.isArray(d.list) ? d.list : [];
    userTotal.value = d.total ?? 0;
    userPage.value = d.page ?? page;
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    usersLoading.value = false;
  }
}

// 角色/状态更新：全量回传 {id, role, status}
async function patchUser(
  u: UserItem,
  patch: Partial<Pick<UserItem, "role" | "status">>
) {
  const body = { id: u.id, role: u.role, status: u.status, ...patch };
  await api("/api/user/update", { method: "POST", body });
  Object.assign(u, patch);
}

async function toggleUserStatus(u: UserItem, on: boolean) {
  try {
    await patchUser(u, { status: on ? 1 : 0 });
    $q.notify({ message: on ? "已启用" : "已禁用", color: "positive" });
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
    await loadUsers();
  }
}

async function toggleUserRole(u: UserItem) {
  try {
    await patchUser(u, { role: u.role === 2 ? 1 : 2 });
    $q.notify({ message: "角色已更新", color: "positive" });
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
    await loadUsers();
  }
}

// ---- 添加用户 ----
const userAddDlg = ref(false);
const userAddBusy = ref(false);
const addForm = reactive({ username: "", email: "", password: "", role: 1 });
const roleOptions = [
  { label: "普通用户", value: 1 },
  { label: "管理员", value: 2 }
];

function openAddUser() {
  addForm.username = "";
  addForm.email = "";
  addForm.password = "";
  addForm.role = 1;
  userAddDlg.value = true;
}

async function saveAddUser() {
  const username = addForm.username.trim();
  const email = addForm.email.trim();
  if (!username) {
    $q.notify({ message: "用户名不能为空", color: "negative" });
    return;
  }
  if (!email) {
    $q.notify({ message: "邮箱不能为空", color: "negative" });
    return;
  }
  if (addForm.password.length < 8) {
    $q.notify({ message: "密码长度至少 8 位", color: "negative" });
    return;
  }
  userAddBusy.value = true;
  try {
    await api("/api/user/create", {
      method: "POST",
      body: {
        username,
        email,
        password: addForm.password,
        role: addForm.role
      }
    });
    $q.notify({ message: "用户已创建", color: "positive" });
    userAddDlg.value = false;
    await loadUsers(1);
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    userAddBusy.value = false;
  }
}

// ---- 编辑用户（修改用户名 / 邮箱 / 密码） ----
const userEditDlg = ref(false);
const userEditBusy = ref(false);
const editForm = reactive({ id: "", username: "", email: "", password: "" });

function openEditUser(u: UserItem) {
  editForm.id = u.id;
  editForm.username = u.username;
  editForm.email = u.email;
  editForm.password = "";
  userEditDlg.value = true;
}

async function saveEditUser() {
  const username = editForm.username.trim();
  if (!username) {
    $q.notify({ message: "用户名不能为空", color: "negative" });
    return;
  }
  const email = editForm.email.trim();
  if (!email || !email.includes("@")) {
    $q.notify({ message: "请输入正确的邮箱", color: "negative" });
    return;
  }
  if (editForm.password && editForm.password.length < 8) {
    $q.notify({ message: "密码长度至少 8 位", color: "negative" });
    return;
  }
  const body: Record<string, unknown> = { id: editForm.id, username, email };
  if (editForm.password) body.password = editForm.password;

  userEditBusy.value = true;
  try {
    await api("/api/user/update-info", { method: "POST", body });
    $q.notify({
      message: editForm.password ? "已保存，该用户的旧登录已失效" : "已保存",
      color: "positive"
    });
    userEditDlg.value = false;
    await loadUsers();
    await loadMe(); // 改的是自己时同步刷新页面上显示的用户名/邮箱
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    userEditBusy.value = false;
  }
}

// ---- 设备管理（登录会话） ----
interface UserSessionItem {
  token: string;
  ip: string;
  device: string;
  created_at: string;
  expires_at: string;
  expired: boolean;
  current: boolean;
}

const userDevDlg = ref(false);
const devLoading = ref(false);
const devUser = ref<UserItem | null>(null);
const devSessions = ref<UserSessionItem[]>([]);

async function openUserDevices(u: UserItem) {
  devUser.value = u;
  devSessions.value = [];
  userDevDlg.value = true;
  devLoading.value = true;
  try {
    const d = await api(
      `/api/user/sessions?user_id=${encodeURIComponent(u.id)}`
    );
    devSessions.value = Array.isArray(d.list) ? d.list : [];
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    devLoading.value = false;
  }
}

function kickSession(s: UserSessionItem) {
  if (!devUser.value) return;
  $q.dialog({
    title: "踢下线",
    message: `确定让「${s.device || "未知设备"}」退出登录？`,
    cancel: true
  }).onOk(async () => {
    try {
      await api("/api/user/kick", {
        method: "POST",
        body: { user_id: devUser.value!.id, token: s.token }
      });
      $q.notify({ message: "已踢下线", color: "positive" });
      if (devUser.value) await openUserDevices(devUser.value);
    } catch (e) {
      $q.notify({ message: describeError(e), color: "negative" });
    }
  });
}

function removeUser(u: UserItem) {
  $q.dialog({
    title: "删除用户",
    message: `确定删除「${u.username}」？其登录会话会立即失效。`,
    cancel: true
  }).onOk(async () => {
    try {
      await api("/api/user/delete", { method: "POST", body: { id: u.id } });
      $q.notify({ message: "已删除", color: "positive" });
      await loadUsers();
    } catch (e) {
      $q.notify({ message: describeError(e), color: "negative" });
    }
  });
}

function fmtTime(v: string) {
  const d = new Date(v);
  if (Number.isNaN(d.getTime())) return v;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getMonth() + 1}-${d.getDate()} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

// ================= AI 设置 =================
const aiForm = reactive({
  base_url: "",
  api_key: "",
  model: "",
  system_prompt: "",
  thinking: false,
  reasoning_effort: "medium"
});
// 思考强度档位（对应 OpenAI reasoning_effort，通用三档）
const effortOptions = [
  { label: "低（快）", value: "low" },
  { label: "中（默认）", value: "medium" },
  { label: "高（深思）", value: "high" }
];
const aiSaving = ref(false);
const aiTesting = ref(false);
const aiTestResult = ref("");
const aiTestOk = ref(false);
// 后端按当前上架商品真实拼接的 user 消息模板（切到本页签时刷新）
const aiUserTemplate = ref("");

async function loadAi() {
  try {
    const d = await api("/api/ai/settings");
    aiForm.base_url = d.base_url ?? "";
    aiForm.api_key = d.api_key ?? "";
    aiForm.model = d.model ?? "";
    aiForm.system_prompt = d.system_prompt ?? "";
    aiForm.thinking = d.thinking === true;
    aiForm.reasoning_effort = d.reasoning_effort || "medium";
    aiUserTemplate.value = d.user_prompt_template ?? "";
    aiTestResult.value = "";
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  }
}

async function saveAi() {
  if (!aiForm.base_url.trim() || !aiForm.model.trim()) {
    $q.notify({ message: "接口地址和模型名不能为空", color: "negative" });
    return;
  }
  if (!aiForm.api_key.trim()) {
    $q.notify({ message: "API Key 不能为空", color: "negative" });
    return;
  }
  if (!aiForm.system_prompt.trim()) {
    $q.notify({ message: "系统提示词不能为空", color: "negative" });
    return;
  }
  aiSaving.value = true;
  try {
    await api("/api/ai/save", {
      method: "POST",
      body: {
        base_url: aiForm.base_url.trim(),
        api_key: aiForm.api_key.trim(),
        model: aiForm.model.trim(),
        system_prompt: aiForm.system_prompt.trim(),
        thinking: aiForm.thinking,
        reasoning_effort: aiForm.reasoning_effort
      }
    });
    $q.notify({ message: "已保存，立即生效", color: "positive" });
  } catch (e) {
    $q.notify({ message: describeError(e), color: "negative" });
  } finally {
    aiSaving.value = false;
  }
}

async function testAi() {
  aiTesting.value = true;
  aiTestResult.value = "";
  try {
    const d = await api("/api/ai/test", {
      method: "POST",
      body: {
        base_url: aiForm.base_url.trim(),
        api_key: aiForm.api_key.trim(),
        model: aiForm.model.trim()
      }
    });
    aiTestOk.value = true;
    aiTestResult.value = `连通成功，模型回复：${d.reply}`;
  } catch (e) {
    aiTestOk.value = false;
    aiTestResult.value = describeError(e);
  } finally {
    aiTesting.value = false;
  }
}

// 拼接后的完整提示词：system 提示词（输入框实时值）+ 后端真实拼接的 user 消息
// （user 部分由 /api/ai/settings 返回，商品列表取当前上架商品，与 MatchGoods 实际发送一致）
const aiPromptPreview = computed(() => {
  const sys = aiForm.system_prompt.trim() || "（系统提示词为空）";
  const user = aiUserTemplate.value || "（商品列表加载中…）";
  return "【system 系统提示词】\n" + sys + "\n\n【user 用户消息】\n" + user;
});

// 切到用户管理 / AI 设置页签时懒加载数据
watch(tab, v => {
  if (v === "users") void loadUsers(1);
  else if (v === "ai") void loadAi();
});

// ================= 初始化 =================
onMounted(async () => {
  if (!token.value) return;
  try {
    await loadMe();
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

/* 商品媒体：多图/视频缩略图网格 */
.media-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(72px, 1fr));
  gap: 8px;
}

/* 网页嵌套项：缩略图占位 */
.media-web {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #eef1f5;
  color: #5f6b7a;
}

/* 添加第三方链接按钮：与左侧 dense 上传框同高（40px）、同圆角 */
.media-link-btn {
  height: 40px;
  border-radius: 4px;
}

.media-item {
  position: relative;
  aspect-ratio: 1;
  border-radius: 6px;
  overflow: hidden;
  background: #f5f5f5;

  img,
  video {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
}

.media-remove {
  position: absolute;
  top: 2px;
  right: 2px;
  z-index: 1;
}

.media-tag {
  position: absolute;
  left: 2px;
  bottom: 2px;
  padding: 0 4px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  font-size: 10px;
  line-height: 16px;
}

.menu-row {
  cursor: grab;
  &:active {
    cursor: grabbing;
  }
}

/* 商品行：左图 + 右侧上下两段（上规格 / 下操作） */
.goods-row {
  align-items: flex-start;
  padding-top: 8px;
  padding-bottom: 8px;
}

.goods-photo {
  padding-top: 2px;
}

.goods-body {
  min-width: 0;
}

.goods-spec,
.goods-actions {
  width: 100%;
  min-width: 0;
}

.goods-head {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}

.goods-name {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.goods-price {
  flex: 0 0 auto;
  white-space: nowrap;
}

.goods-meta {
  margin-top: 2px;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.goods-actions {
  display: flex;
  align-items: center;
  margin-top: 6px;
}

@media (max-width: 599px) {
  .goods-thumb {
    width: 34px;
    height: 34px;
  }

  .goods-actions {
    margin-top: 4px;
  }
}
</style>
