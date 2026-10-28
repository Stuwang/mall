<script setup lang="ts">
import { computed } from 'vue'
import { showImagePreview } from 'vant'

const props = defineProps<{
  images: string[]
}>()

const urls = computed(() => (props.images || []).filter(Boolean))

type GridCell = { src?: string; more?: number; index: number }

/** 实际渲染的格子：可能含 more 占位 */
const cells = computed((): GridCell[] => {
  const list = urls.value
  const n = list.length
  if (n === 0) {
    return []
  }
  if (n <= 4) {
    return list.map((src, index) => ({ src, index }))
  }
  if (n <= 8) {
    // 四宫格：前 3 张 + 更多
    return [
      { src: list[0], index: 0 },
      { src: list[1], index: 1 },
      { src: list[2], index: 2 },
      { more: n - 3, index: 3, src: list[3] },
    ]
  }
  // 九宫格：前 9 张；超过 9 时最后一格更多
  if (n === 9) {
    return list.map((src, index) => ({ src, index }))
  }
  const out: GridCell[] = list.slice(0, 8).map((src, index) => ({ src, index }))
  out.push({ more: n - 8, index: 8, src: list[8] })
  return out
})

const layoutClass = computed(() => {
  const n = urls.value.length
  if (n <= 0) return ''
  if (n === 1) return 'layout-1'
  if (n === 2) return 'layout-2'
  if (n === 3) return 'layout-3'
  if (n === 4) return 'layout-4'
  if (n <= 8) return 'layout-4'
  return 'layout-9'
})

function onClick(cell: { src?: string; more?: number; index: number }, ev: Event): void {
  ev.stopPropagation()
  const start = cell.more != null ? cell.index : cell.index
  showImagePreview({
    images: urls.value,
    startPosition: Math.min(start, urls.value.length - 1),
    closeable: true,
  })
}
</script>

<template>
  <div v-if="urls.length > 0" class="grid" :class="layoutClass" @click.stop>
    <button
      v-for="(cell, i) in cells"
      :key="i"
      type="button"
      class="cell"
      @click="onClick(cell, $event)"
    >
      <img v-if="cell.src" :src="cell.src" alt="" />
      <span v-if="cell.more" class="more">+{{ cell.more }}</span>
    </button>
  </div>
</template>

<style scoped>
.grid {
  display: grid;
  gap: 4px;
  margin-top: 10px;
  width: 100%;
}

.layout-1 {
  grid-template-columns: 1fr;
  max-width: 260px;
}

.layout-1 .cell {
  aspect-ratio: 4 / 3;
  max-height: 280px;
}

.layout-2 {
  grid-template-columns: 1fr 1fr;
  max-width: 100%;
}

.layout-3 {
  grid-template-columns: 1fr 1fr 1fr;
}

.layout-4 {
  grid-template-columns: 1fr 1fr;
  max-width: 280px;
}

.layout-9 {
  grid-template-columns: 1fr 1fr 1fr;
  max-width: 100%;
}

.cell {
  position: relative;
  padding: 0;
  border: none;
  border-radius: 6px;
  overflow: hidden;
  background: #f2f3f5;
  aspect-ratio: 1;
  cursor: pointer;
}

.cell img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.more {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 20px;
  font-weight: 600;
}
</style>
