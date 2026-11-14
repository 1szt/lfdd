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
      <q-tabs v-model="tab" dense class="bg-white text-grey-7" active-color="primary" indicator-color="primary"
        align="justify" narrow-indicator>
        <q-route-tab to="/" name="menu" icon="shopping_bag" label="商品" />
        <q-route-tab to="/cart" name="cart" icon="shopping_cart" label="购物车" />
        <q-route-tab to="/order" name="order" icon="receipt_long" label="订单" />
        <q-route-tab to="/admin" name="admin" icon="admin_panel_settings" label="管理" />
      </q-tabs>
    </q-footer>
  </q-layout>
</template>

<script setup lang="ts">
import { ref } from "vue";


// 底部导航当前选中项
const tab = ref("menu");


</script>
