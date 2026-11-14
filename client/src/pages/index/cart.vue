<template>
  <q-page class="cart-page column" :style-fn="pageStyle">
    <!-- 标题栏 -->
    <div class="row items-center justify-between q-px-md q-py-sm">
      <span class="text-subtitle1 text-weight-medium">购物车</span>
      <q-btn
        v-if="cart.length > 0"
        flat
        dense
        size="sm"
        color="negative"
        icon="delete_sweep"
        label="清空"
        @click="clearCart"
      />
    </div>

    <!-- 空态 -->
    <div v-if="cart.length === 0" class="column items-center justify-center empty-box">
      <q-icon name="shopping_cart" size="64px" color="grey-4" />
      <div class="text-h6 text-grey-7 q-mt-md">购物车是空的</div>
      <q-btn class="q-mt-md" color="primary" label="去点单" to="/" />
    </div>

    <!-- 明细列表（带图片） -->
    <div v-else class="cart-list">
      <div v-for="line in cart" :key="line.id" class="cart-line row items-center">
        <div class="cart-thumb">
          <img v-if="line.image" :src="imgSrc(line.image)" :alt="line.name" />
          <q-icon v-else name="shopping_bag" size="28px" color="grey-5" />
        </div>
        <div class="col">
          <div class="text-weight-medium ellipsis">{{ line.name }}</div>
          <div v-if="line.spec" class="text-caption text-grey-7 ellipsis">{{ line.spec }}</div>
          <div class="text-negative text-weight-bold q-mt-xs">
            ¥{{ fmtPrice(line.price) }}
            <span class="text-caption text-grey-6 text-weight-regular"> / 件</span>
          </div>
        </div>
        <div class="column items-end q-gutter-xs">
          <div class="row items-center q-gutter-sm">
            <q-btn
              round
              dense
              color="primary"
              icon="remove"
              size="md"
              aria-label="减少"
              @click="changeQty(line, -1)"
            />
            <q-input
              :model-value="line.qty"
              type="number"
              dense
              borderless
              class="qty-input"
              aria-label="修改数量"
              @change="onQtyInput(line, $event)"
            />
            <q-btn
              round
              dense
              color="primary"
              icon="add"
              size="md"
              aria-label="增加"
              @click="changeQty(line, 1)"
            />
            <q-btn
              flat
              dense
              round
              color="negative"
              icon="delete"
              size="sm"
              aria-label="删除该商品"
              @click="confirmRemove(line)"
            />
          </div>
          <div class="text-caption text-grey-6">小计 ¥{{ fmtPrice(line.price * line.qty) }}</div>
        </div>
      </div>
    </div>

    <!-- 底部合计条 -->
    <div v-if="cart.length > 0" class="cart-summary row items-center q-px-md">
      <div class="col">
        <div class="text-caption text-grey-7" style="line-height: 1">共 {{ cartCount }} 件</div>
        <div class="text-h6 text-negative text-weight-bold" style="line-height: 1.3">
          ¥{{ fmtPrice(cartTotal) }}
        </div>
      </div>
      <q-btn color="primary" label="去结算" unelevated :loading="ordering" @click="checkout" />
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useQuasar } from "quasar";
import { useRouter } from "vue-router";
import {
  cart,
  cartCount,
  cartTotal,
  changeQty,
  clearCart,
  fmtPrice,
  imgSrc,
  removeLine,
  setQty
} from "../../utils/cart";

const $q = useQuasar();
const router = useRouter();
const ordering = ref(false);

// QPage styleFn：header/footer 占位后剩余高度，让页面撑满、列表区各自滚动
function pageStyle(offset: number, height: number) {
  if (height > 0) {
    return { height: `${height - offset}px` };
  }
  return { height: offset !== 0 ? `calc(100vh - ${offset}px)` : "100vh" };
}

// 数量输入框直接改数字：0/非法 → 移除该行，超库存由 setQty 封顶
function onQtyInput(line: { id: string; name: string; price: number; stock: number; image?: string; spec?: string }, v: string | number) {
  const n = Math.floor(Number(v));
  setQty(line, Number.isFinite(n) ? n : 0);
}

// 删除单个商品：二次确认，防止手误点掉
function confirmRemove(line: { id: string; name: string }) {
  $q.dialog({
    title: "删除商品",
    message: `确定把「${line.name}」从购物车删除吗？`,
    cancel: true,
    persistent: true
  }).onOk(() => {
    removeLine(line.id);
  });
}

// 结算：购物车明细提交后端建订单（价格由后端按库结算），成功后清空并跳订单页
async function checkout() {
  if (ordering.value) return;
  if (cart.value.length === 0) {
    $q.notify({ message: "购物车是空的", color: "orange" });
    return;
  }
  ordering.value = true;
  try {
    const res = await fetch("/api/order/create", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        items: cart.value.map((l) => ({ id: l.id, qty: l.qty }))
      })
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.message || `下单失败（${res.status}）`);

    clearCart();
    $q.notify({ message: `下单成功 ${data.order_no}`, color: "positive" });
    void router.push("/order");
  } catch (e) {
    $q.notify({ message: e instanceof Error ? e.message : String(e), color: "negative" });
  } finally {
    ordering.value = false;
  }
}
</script>

<style scoped lang="scss">
.cart-page {
  overflow: hidden;
  background: #fff;
}

.empty-box {
  flex: 1;
}

.cart-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 0 16px;
}

.cart-line {
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;

  &:last-child {
    border-bottom: none;
  }
}

.cart-thumb {
  width: 64px;
  height: 64px;
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

/* 数量输入框：直接改数字（0 = 移除），隐藏数字框 spinner */
.qty-input {
  width: 48px;

  :deep(input) {
    text-align: center;
    font-size: 15px;
    font-weight: 700;
    padding: 0;
    -moz-appearance: textfield;
  }

  :deep(input::-webkit-outer-spin-button),
  :deep(input::-webkit-inner-spin-button) {
    -webkit-appearance: none;
    margin: 0;
  }
}

.cart-summary {
  gap: 12px;
  padding-top: 10px;
  padding-bottom: 12px;
  border-top: 1px solid #f0f0f0;
  background: #fff;
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.06);
}
</style>
