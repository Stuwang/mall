<script setup lang="ts">
import { computed } from 'vue'
import { avatarColor, avatarLabel } from '@/utils/avatar'

const props = withDefaults(
  defineProps<{
    userId: number
    nickname?: string
    size?: number
  }>(),
  { nickname: '', size: 36 },
)

const style = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
  fontSize: `${Math.max(12, Math.round(props.size * 0.4))}px`,
  background: avatarColor(props.userId),
}))

const text = computed(() => avatarLabel(props.nickname, props.userId))
</script>

<template>
  <div class="user-avatar" :style="style" aria-hidden="true">{{ text }}</div>
</template>

<style scoped>
.user-avatar {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  color: #fff;
  font-weight: 600;
  line-height: 1;
  user-select: none;
}
</style>
