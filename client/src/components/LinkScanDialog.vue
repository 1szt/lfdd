<template>
  <q-dialog
    :model-value="modelValue"
    persistent
    @update:model-value="v => emit('update:modelValue', v)"
    @hide="stopScan"
  >
    <q-card class="link-scan-card">
      <q-card-section class="text-subtitle1">添加第三方链接</q-card-section>

      <q-card-section class="q-pt-none">
        <!-- 输入框右侧「扫一扫」：打开摄像头识别二维码并填入链接 -->
        <div class="row no-wrap items-start">
          <q-input
            v-model="url"
            class="col"
            outlined
            dense
            autofocus
            placeholder="粘贴 http/https 链接"
            @keyup.enter="submit"
          />
          <q-btn
            class="q-ml-sm scan-btn"
            outline
            dense
            no-caps
            color="primary"
            icon="qr_code_scanner"
            :label="scanning ? '停止' : '扫一扫'"
            @click="toggleScan"
          />
        </div>

        <!-- 取景框：扫描中才显示 -->
        <div v-if="scanning" class="scan-box q-mt-sm">
          <video ref="videoEl" class="scan-video" muted playsinline />
          <div class="scan-frame" />
          <div class="scan-hint">将二维码放入框内，识别后自动填入</div>
        </div>

        <div class="text-caption text-grey-6 q-mt-sm">
          图片/视频按扩展名识别，其它链接按网页嵌套显示（对方站点禁止嵌套时会打不开）
        </div>
      </q-card-section>

      <q-card-actions align="right">
        <q-btn flat label="取消" @click="close" />
        <q-btn
          color="primary"
          label="添加"
          :disable="!url.trim()"
          @click="submit"
        />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref } from "vue";
import { useQuasar } from "quasar";
import jsQR from "jsqr";

defineProps<{ modelValue: boolean }>();
const emit = defineEmits<{
  (e: "update:modelValue", v: boolean): void;
  (e: "ok", url: string): void;
}>();

const $q = useQuasar();
const url = ref("");
const scanning = ref(false);
const videoEl = ref<HTMLVideoElement | null>(null);

let stream: MediaStream | null = null;
let rafId = 0;
let lastDecode = 0;
let canvas: HTMLCanvasElement | null = null;

function describeError(e: unknown) {
  return e instanceof Error ? e.message : String(e);
}

// 摄像头扫码：getUserMedia + jsQR 逐帧解码（降采样到 640px 宽，限制解码频率省电）
async function startScan() {
  if (scanning.value) return;
  if (!navigator.mediaDevices?.getUserMedia) {
    $q.notify({
      message: "当前环境不支持摄像头扫码，请手动粘贴链接",
      color: "negative"
    });
    return;
  }
  scanning.value = true;
  try {
    stream = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: { ideal: "environment" } },
      audio: false
    });
    await nextTick(); // 等取景框渲染出 video 元素
    const v = videoEl.value;
    if (!v) return;
    v.srcObject = stream;
    await v.play();
    rafId = requestAnimationFrame(tick);
  } catch (e) {
    stopScan();
    $q.notify({
      message: `打开摄像头失败：${describeError(e)}`,
      color: "negative"
    });
  }
}

function tick(ts: number) {
  if (!scanning.value) return;
  rafId = requestAnimationFrame(tick);
  if (ts - lastDecode < 150) return; // 每秒约 6 次解码
  lastDecode = ts;

  const v = videoEl.value;
  if (!v || v.readyState < 2 || !v.videoWidth) return;

  const scale = Math.min(1, 640 / v.videoWidth);
  const w = Math.round(v.videoWidth * scale);
  const h = Math.round(v.videoHeight * scale);
  if (!canvas) canvas = document.createElement("canvas");
  canvas.width = w;
  canvas.height = h;
  const ctx = canvas.getContext("2d", { willReadFrequently: true });
  if (!ctx) return;
  ctx.drawImage(v, 0, 0, w, h);
  const image = ctx.getImageData(0, 0, w, h);
  const code = jsQR(image.data, w, h, { inversionAttempts: "dontInvert" });
  if (code?.data) {
    url.value = code.data.trim();
    stopScan();
    $q.notify({ message: "已识别二维码", color: "positive" });
  }
}

function stopScan() {
  scanning.value = false;
  if (rafId) {
    cancelAnimationFrame(rafId);
    rafId = 0;
  }
  if (stream) {
    for (const t of stream.getTracks()) t.stop();
    stream = null;
  }
  if (videoEl.value) videoEl.value.srcObject = null;
}

function toggleScan() {
  if (scanning.value) stopScan();
  else void startScan();
}

function close() {
  stopScan();
  emit("update:modelValue", false);
}

function submit() {
  const v = url.value.trim();
  if (!v) return;
  stopScan();
  emit("ok", v);
  emit("update:modelValue", false);
}

// 组件销毁（如切换页面）时释放摄像头
onBeforeUnmount(stopScan);
</script>

<style scoped lang="scss">
.link-scan-card {
  width: 420px;
  max-width: 92vw;
}

.scan-btn {
  height: 40px;
  border-radius: 4px;
}

.scan-box {
  position: relative;
  height: 240px;
  border-radius: 8px;
  overflow: hidden;
  background: #000;
}

.scan-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* 中间取景框：四个角的描边效果用外框 + 内框模拟 */
.scan-frame {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 160px;
  height: 160px;
  transform: translate(-50%, -50%);
  border: 2px solid rgba(255, 255, 255, 0.9);
  border-radius: 8px;
  box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.25);
}

.scan-hint {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 8px;
  text-align: center;
  font-size: 12px;
  color: #fff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.6);
}
</style>
