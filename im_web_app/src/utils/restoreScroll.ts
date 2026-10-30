import { onActivated, onUnmounted } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'

/**
 * 记住 window 滚动位置，页面从 KeepAlive 里再次显示时设回去。
 * pinTop.current 为 true 时，这次激活改回顶部（列表被整表刷新）。
 * 调用方的 onActivated 要写在本函数之前，这样能先打上 pinTop。
 */
export function useRestoreScroll(pinTop?: { current: boolean }): void {
  let top = 0

  onBeforeRouteLeave(() => {
    top = window.scrollY
  })

  onActivated(() => {
    const y = pinTop?.current ? 0 : top
    if (pinTop) {
      pinTop.current = false
    }
    if (y === 0) {
      top = 0
    }
    requestAnimationFrame(() => {
      window.scrollTo(0, y)
    })
  })

  onUnmounted(() => {
    top = 0
  })
}
