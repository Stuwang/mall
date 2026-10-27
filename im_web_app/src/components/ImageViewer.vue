<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'

const props = defineProps<{
  show: boolean
  src: string
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
}>()

const MIN_SCALE = 1
const MAX_SCALE = 4
const TAP_SLOP = 8

const scale = ref(1)
const rotate = ref(0)
const offsetX = ref(0)
const offsetY = ref(0)
const smooth = ref(false)
const stageRef = ref<HTMLElement | null>(null)

const points = new Map<number, { x: number; y: number }>()
let pinchDist = 0
let pinchScale = 1
let pan: { x: number; y: number; ox: number; oy: number } | null = null
let moved = false
let lastTapAt = 0

const transform = computed(
  () =>
    `translate(${offsetX.value}px, ${offsetY.value}px) scale(${scale.value}) rotate(${rotate.value}deg)`,
)

function clampScale(value: number): number {
  return Math.min(MAX_SCALE, Math.max(MIN_SCALE, value))
}

function resetView(): void {
  scale.value = MIN_SCALE
  rotate.value = 0
  offsetX.value = 0
  offsetY.value = 0
}

function restore(): void {
  smooth.value = true
  resetView()
}

function close(): void {
  emit('update:show', false)
}

function setScale(value: number, animate: boolean): void {
  smooth.value = animate
  const next = Math.round(clampScale(value) * 100) / 100
  scale.value = next
  if (next <= MIN_SCALE) {
    offsetX.value = 0
    offsetY.value = 0
  }
}

function rotateClockwise(): void {
  smooth.value = true
  rotate.value = (rotate.value + 90) % 360
}

function distance(a: { x: number; y: number }, b: { x: number; y: number }): number {
  return Math.hypot(a.x - b.x, a.y - b.y)
}

function onPointerDown(ev: PointerEvent): void {
  const stage = stageRef.value
  if (!stage) {
    return
  }
  stage.setPointerCapture(ev.pointerId)
  smooth.value = false
  const starting = points.size === 0
  points.set(ev.pointerId, { x: ev.clientX, y: ev.clientY })
  if (starting) {
    moved = false
  }
  if (points.size >= 2) {
    const [a, b] = [...points.values()]
    pinchDist = distance(a, b) || 1
    pinchScale = scale.value
    pan = null
    return
  }
  pan = { x: ev.clientX, y: ev.clientY, ox: offsetX.value, oy: offsetY.value }
}

function onPointerMove(ev: PointerEvent): void {
  if (!points.has(ev.pointerId)) {
    return
  }
  points.set(ev.pointerId, { x: ev.clientX, y: ev.clientY })
  if (points.size >= 2 && pinchDist > 0) {
    const [a, b] = [...points.values()]
    const next = clampScale(pinchScale * (distance(a, b) / pinchDist))
    if (Math.abs(next - scale.value) > 0.01) {
      moved = true
    }
    setScale(next, false)
    return
  }
  if (!pan || scale.value <= MIN_SCALE) {
    if (pan && Math.hypot(ev.clientX - pan.x, ev.clientY - pan.y) > TAP_SLOP) {
      moved = true
    }
    return
  }
  const dx = ev.clientX - pan.x
  const dy = ev.clientY - pan.y
  if (Math.hypot(dx, dy) > TAP_SLOP) {
    moved = true
  }
  offsetX.value = pan.ox + dx
  offsetY.value = pan.oy + dy
}

function onPointerUp(ev: PointerEvent): void {
  const wasTap = !moved && points.size <= 1
  const target = ev.target
  points.delete(ev.pointerId)
  if (points.size < 2) {
    pinchDist = 0
  }
  if (points.size === 1) {
    const rest = [...points.values()][0]
    pan = { x: rest.x, y: rest.y, ox: offsetX.value, oy: offsetY.value }
  } else {
    pan = null
  }
  if (wasTap && target instanceof Element && target.classList.contains('stage')) {
    close()
  }
}

function onImageClick(): void {
  if (moved) {
    moved = false
    return
  }
  const now = Date.now()
  if (now - lastTapAt < 280) {
    lastTapAt = 0
    setScale(scale.value > MIN_SCALE ? MIN_SCALE : 2, true)
    return
  }
  lastTapAt = now
}

function onWheel(ev: WheelEvent): void {
  ev.preventDefault()
  setScale(scale.value + (ev.deltaY < 0 ? 0.25 : -0.25), false)
}

function onKeydown(ev: KeyboardEvent): void {
  if (ev.key === 'Escape') {
    close()
  }
}

watch(
  () => props.show,
  (visible) => {
    if (visible) {
      resetView()
      lastTapAt = 0
      window.addEventListener('keydown', onKeydown)
      document.body.style.overflow = 'hidden'
      return
    }
    window.removeEventListener('keydown', onKeydown)
    document.body.style.overflow = ''
    points.clear()
    pan = null
    pinchDist = 0
  },
)

watch(stageRef, (el, prev) => {
  prev?.removeEventListener('wheel', onWheel)
  el?.addEventListener('wheel', onWheel, { passive: false })
})

onUnmounted(() => {
  stageRef.value?.removeEventListener('wheel', onWheel)
  window.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="viewer" role="dialog" aria-modal="true" aria-label="图片预览">
      <button type="button" class="close" aria-label="关闭" @click="close">
        <van-icon name="cross" />
      </button>
      <div
        ref="stageRef"
        class="stage"
        @pointerdown="onPointerDown"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
        @pointercancel="onPointerUp"
      >
        <img
          class="photo"
          :class="{ smooth }"
          :src="src"
          alt="图片预览"
          draggable="false"
          :style="{ transform }"
          @click.stop="onImageClick"
        />
      </div>
      <div class="toolbar">
        <button type="button" aria-label="缩小" @click="setScale(scale - 0.5, true)">
          <van-icon name="minus" />
        </button>
        <button type="button" aria-label="放大" @click="setScale(scale + 0.5, true)">
          <van-icon name="plus" />
        </button>
        <button type="button" aria-label="旋转" @click="rotateClockwise">
          <van-icon name="replay" />
        </button>
        <button type="button" aria-label="还原" @click="restore">
          <van-icon name="revoke" />
        </button>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.viewer {
  position: fixed;
  inset: 0;
  z-index: 3000;
  background: rgba(0, 0, 0, 0.92);
}

.stage {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  touch-action: none;
}

.photo {
  max-width: min(92vw, 92dvh);
  max-height: min(92vw, 92dvh);
  object-fit: contain;
  transform-origin: center center;
  user-select: none;
  -webkit-user-drag: none;
}

.photo.smooth {
  transition: transform 0.2s ease;
}

.close {
  position: absolute;
  top: calc(12px + env(safe-area-inset-top, 0px));
  right: 12px;
  z-index: 1;
  width: 40px;
  height: 40px;
  border: none;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  font-size: 20px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.toolbar {
  position: absolute;
  left: 0;
  right: 0;
  bottom: calc(16px + env(safe-area-inset-bottom, 0px));
  z-index: 1;
  display: flex;
  justify-content: center;
  gap: 12px;
}

.toolbar button {
  width: 44px;
  height: 44px;
  border: none;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  font-size: 20px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}
</style>
