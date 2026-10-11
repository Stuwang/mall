import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { CartItem } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { useAuthStore } from '@/stores/authStore'

/**
 * 购物车状态：内部委托 CartApiService
 */
export const useCartStore = defineStore('cart', () => {
  const api = appContext.api.cart
  const items = ref<CartItem[]>([])
  const totalQuantity = ref(0)
  const loading = ref(false)

  const isEmpty = computed(() => items.value.length === 0)

  async function refresh(): Promise<void> {
    const auth = useAuthStore()
    if (!auth.isLoggedIn) {
      items.value = []
      totalQuantity.value = 0
      return
    }
    loading.value = true
    try {
      const reply = await api.list({})
      items.value = reply.list ?? []
      totalQuantity.value = reply.total_quantity ?? 0
    } finally {
      loading.value = false
    }
  }

  async function add(productId: number, quantity = 1): Promise<void> {
    await api.add({ product_id: productId, quantity })
    await refresh()
  }

  async function updateQuantity(id: number, quantity: number): Promise<void> {
    await api.updateQuantity({ id, quantity })
    await refresh()
  }

  async function remove(id: number): Promise<void> {
    await api.remove({ id })
    await refresh()
  }

  function clearLocal(): void {
    items.value = []
    totalQuantity.value = 0
  }

  return {
    items,
    totalQuantity,
    loading,
    isEmpty,
    refresh,
    add,
    updateQuantity,
    remove,
    clearLocal,
  }
})
