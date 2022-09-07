<template>
  <canvas ref="confettiEl" class="confetti"></canvas>
</template>
<style scoped>
.confetti {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  right: 0;
  width: 100vw;
  height: 100vh;
  z-index: 999999;
  pointer-events: none;
}
</style>
<script setup lang="ts">
import confetti from "canvas-confetti";
import {onMounted, ref} from "vue";

const confettiEl = ref<HTMLCanvasElement>();
let confettiRef: ReturnType<typeof confetti.create>;
onMounted(() => {
  confettiRef = confetti.create(confettiEl.value!, {
    resize: true,
    useWorker: true,
    disableForReducedMotion: true,
  });
})
function fire(options?: confetti.Options) {
  return confettiRef(options) as Promise<unknown>;
}
defineExpose({fire});
</script>
