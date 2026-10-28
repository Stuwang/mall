<script setup lang="ts">
import { onUnmounted, ref } from 'vue'
import { imClient } from '@/im/ImClient'

type EffectKind = 'cheers' | 'car' | 'castle' | 'default'

const bubbles = [
  { id: 1, left: '18%', delay: '0.05s', glyph: '✨' },
  { id: 2, left: '30%', delay: '0.18s', glyph: '🥂' },
  { id: 3, left: '42%', delay: '0.08s', glyph: '✨' },
  { id: 4, left: '54%', delay: '0.28s', glyph: '✦' },
  { id: 5, left: '66%', delay: '0.12s', glyph: '✨' },
  { id: 6, left: '76%', delay: '0.22s', glyph: '⭐' },
]

const stars = [
  { id: 1, x: '-120px', y: '-90px', delay: '0s', glyph: '✦' },
  { id: 2, x: '110px', y: '-80px', delay: '0.05s', glyph: '⭐' },
  { id: 3, x: '-90px', y: '70px', delay: '0.1s', glyph: '✨' },
  { id: 4, x: '100px', y: '80px', delay: '0.08s', glyph: '✦' },
  { id: 5, x: '0px', y: '-130px', delay: '0.12s', glyph: '🎉' },
  { id: 6, x: '20px', y: '120px', delay: '0.16s', glyph: '✨' },
]

const show = ref(false)
const kind = ref<EffectKind>('default')
const playId = ref(0)
let timer = 0

const unsub = imClient.onGiftEffect((event) => {
  const key = event.effect.trim().toLowerCase()
  kind.value = key === 'cheers' || key === 'car' || key === 'castle' ? key : 'default'
  playId.value += 1
  show.value = true
  window.clearTimeout(timer)
  timer = window.setTimeout(() => {
    show.value = false
  }, 2200)
})

onUnmounted(() => {
  unsub()
  window.clearTimeout(timer)
})
</script>

<template>
  <div v-if="show" :key="playId" class="fx" :class="kind" aria-hidden="true">
    <template v-if="kind === 'cheers'">
      <span class="glass left">🥂</span>
      <span class="glass right">🥂</span>
      <span
        v-for="item in bubbles"
        :key="item.id"
        class="floater"
        :style="{ left: item.left, animationDelay: item.delay }"
      >
        {{ item.glyph }}
      </span>
    </template>
    <template v-else-if="kind === 'car'">
      <span class="speed a" />
      <span class="speed b" />
      <span class="speed c" />
      <span class="car">🚗</span>
    </template>
    <template v-else-if="kind === 'castle'">
      <span class="burst">🎆</span>
      <span class="burst late">✨</span>
      <span class="castle">🏰</span>
    </template>
    <template v-else>
      <span class="present">🎁</span>
      <span
        v-for="item in stars"
        :key="item.id"
        class="star"
        :style="{ '--x': item.x, '--y': item.y, animationDelay: item.delay }"
      >
        {{ item.glyph }}
      </span>
    </template>
  </div>
</template>

<style scoped>
.fx {
  position: fixed;
  inset: 0;
  z-index: 4000;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  pointer-events: none;
}

.fx.cheers {
  background: radial-gradient(circle at 50% 48%, rgba(232, 175, 58, 0.32), transparent 58%);
}

.fx.car {
  align-items: flex-end;
  background: linear-gradient(180deg, transparent 40%, rgba(28, 36, 48, 0.12) 100%);
}

.fx.castle {
  background: radial-gradient(circle at 50% 60%, rgba(120, 96, 196, 0.28), transparent 62%);
}

.fx.default {
  background: radial-gradient(circle at 50% 50%, rgba(15, 143, 134, 0.22), transparent 60%);
}

.glass {
  font-size: 72px;
  line-height: 1;
}

.glass.left {
  animation: clink-left 2.2s ease forwards;
}

.glass.right {
  animation: clink-right 2.2s ease forwards;
}

.floater {
  position: absolute;
  bottom: 18%;
  font-size: 22px;
  animation: float-up 2.2s ease forwards;
}

.car {
  position: absolute;
  bottom: 16%;
  left: 0;
  font-size: 68px;
  line-height: 1;
  animation: drive 2.2s linear forwards;
}

.speed {
  position: absolute;
  bottom: 22%;
  height: 3px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.8);
  animation: drive 2.2s linear forwards;
}

.speed.a {
  width: 72px;
  bottom: 24%;
}

.speed.b {
  width: 46px;
  bottom: 19%;
}

.speed.c {
  width: 28px;
  bottom: 28%;
}

.castle {
  font-size: 104px;
  line-height: 1;
  animation: rise 2.2s ease forwards;
}

.burst {
  position: absolute;
  left: 50%;
  top: 42%;
  font-size: 42px;
  animation: spark 2.2s ease forwards;
}

.burst.late {
  margin-left: 120px;
  margin-top: -80px;
  animation-delay: 0.15s;
}

.present {
  font-size: 92px;
  line-height: 1;
  animation: pop 2.2s ease forwards;
}

.star {
  position: absolute;
  left: 50%;
  top: 46%;
  font-size: 26px;
  animation: burst 2.2s ease forwards;
}

@keyframes clink-left {
  0% {
    transform: translate(-90px, 36px) rotate(-20deg);
    opacity: 0;
  }
  20% {
    opacity: 1;
  }
  45% {
    transform: translate(8px, 0) rotate(10deg);
  }
  58% {
    transform: translate(-8px, 6px) rotate(-6deg);
  }
  100% {
    transform: translate(-24px, -16px) rotate(-8deg);
    opacity: 0;
  }
}

@keyframes clink-right {
  0% {
    transform: translate(90px, 36px) rotate(20deg);
    opacity: 0;
  }
  20% {
    opacity: 1;
  }
  45% {
    transform: translate(-8px, 0) rotate(-10deg);
  }
  58% {
    transform: translate(8px, 6px) rotate(6deg);
  }
  100% {
    transform: translate(24px, -16px) rotate(8deg);
    opacity: 0;
  }
}

@keyframes float-up {
  0% {
    transform: translateY(20px) scale(0.6);
    opacity: 0;
  }
  20% {
    opacity: 1;
  }
  100% {
    transform: translateY(-46vh) scale(1);
    opacity: 0;
  }
}

@keyframes drive {
  from {
    transform: translateX(-24vw);
  }
  to {
    transform: translateX(108vw);
  }
}

@keyframes rise {
  0% {
    transform: translateY(48px) scale(0.55);
    opacity: 0;
  }
  22% {
    opacity: 1;
  }
  48% {
    transform: translateY(0) scale(1.06);
  }
  100% {
    transform: translateY(-18px) scale(1);
    opacity: 0;
  }
}

@keyframes spark {
  0% {
    transform: translate(-50%, -50%) scale(0.4);
    opacity: 0;
  }
  30% {
    opacity: 1;
  }
  100% {
    transform: translate(-50%, calc(-50% - 40px)) scale(1.2);
    opacity: 0;
  }
}

@keyframes pop {
  0% {
    transform: scale(0.2) rotate(-8deg);
    opacity: 0;
  }
  18% {
    transform: scale(1.12) rotate(4deg);
    opacity: 1;
  }
  32% {
    transform: scale(1) rotate(0deg);
  }
  100% {
    transform: scale(1.08);
    opacity: 0;
  }
}

@keyframes burst {
  0% {
    transform: translate(-50%, -50%) scale(0.4);
    opacity: 0;
  }
  22% {
    opacity: 1;
  }
  100% {
    transform: translate(calc(-50% + var(--x)), calc(-50% + var(--y))) scale(1);
    opacity: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .fx * {
    animation: none;
  }
}
</style>
