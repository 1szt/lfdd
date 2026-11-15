<template>
  <router-view />
</template>

<script setup lang="ts">
// 数字输入框获得焦点即全选文本，改价/改数量时可直接覆盖输入
document.addEventListener("focusin", e => {
  const el = e.target;
  if (el instanceof HTMLInputElement && el.type === "number") {
    try {
      el.select();
    } catch {
      // 个别浏览器不支持数字输入框全选，忽略即可
    }
  }
});
</script>
