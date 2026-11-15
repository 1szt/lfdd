<template>
  <!-- 应用布局容器 排序：顶部导航栏、左侧抽屉、页面内容、底部导航栏 -->
  <q-layout view="lHh Lpr lFf">
    <!-- 顶部导航栏 -->
    <q-header elevated>
      <!-- 顶部导航栏菜单 -->
    </q-header>

    <q-page-container>
      <router-view />
    </q-page-container>

    <!-- 底部导航栏 -->
    <q-footer elevated>
      <q-tabs
        v-model="tab"
        dense
        class="bg-white text-grey-7"
        active-color="primary"
        indicator-color="primary"
        align="justify"
        narrow-indicator
      >
        <q-route-tab to="/" name="menu" icon="shopping_bag" label="商品" />
        <q-route-tab
          to="/cart"
          name="cart"
          icon="shopping_cart"
          label="购物车"
        />
        <q-route-tab
          to="/order"
          name="order"
          icon="receipt_long"
          label="订单"
        />
        <!-- 管理入口：App 未登录也保留（原生端无法手动输 URL）；浏览器仅登录后显示，游客仍可直接访问 #/admin -->
        <q-route-tab
          v-if="loggedIn || nativeApp"
          to="/admin"
          name="admin"
          icon="admin_panel_settings"
          label="管理"
        />
      </q-tabs>
    </q-footer>
  </q-layout>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { loggedIn } from "../utils/guest";

// 底部导航当前选中项
const tab = ref("menu");

// 是否运行在原生 App（Capacitor）内：原生端没法手动输入 URL，管理入口需始终保留
const nativeApp = Boolean(
  (
    window as unknown as {
      Capacitor?: { isNativePlatform?: () => boolean };
    }
  ).Capacitor?.isNativePlatform?.()
);
</script>
