<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { showConfirmDialog } from 'vant'
import { useRouter } from 'vue-router'
import { appContext } from '@/app/AppContext'
import { PayType } from '@/api/types'
import { routeNames } from '@/router/routes'
import { useCartStore } from '@/stores/cartStore'
import { formatFenToYuan } from '@/utils/format'
import { notifyError, notifySuccess } from '@/utils/notify'

const router = useRouter()
const cartStore = useCartStore()
const orderApi = appContext.api.orders

const checkedIds = ref<number[]>([])
const submitting = ref(false)
const refreshing = ref(false)

const checkedItems = computed(() =>
  cartStore.items.filter((item) => checkedIds.value.includes(item.id)),
)

const checkedAmountFen = computed(() =>
  checkedItems.value.reduce((sum, item) => sum + item.product_price * item.quantity, 0),
)

async function reload(): Promise<void> {
  try {
    await cartStore.refresh()
    const valid = new Set(cartStore.items.map((item) => item.id))
    checkedIds.value = checkedIds.value.filter((id) => valid.has(id))
  } finally {
    refreshing.value = false
  }
}

async function onRefresh(): Promise<void> {
  refreshing.value = true
  await reload()
}

async function onQuantityChange(id: number, quantity: number): Promise<void> {
  try {
    await cartStore.updateQuantity(id, quantity)
    await reload()
  } catch (error) {
    notifyError(error, '更新数量失败')
    await reload()
  }
}

async function onRemove(id: number): Promise<void> {
  try {
    await showConfirmDialog({ title: '移除商品', message: '确定从购物车移除？' })
    await cartStore.remove(id)
    checkedIds.value = checkedIds.value.filter((item) => item !== id)
  } catch (error) {
    if (error === 'cancel') return
    notifyError(error, '移除失败')
  }
}

async function checkout(): Promise<void> {
  if (checkedItems.value.length === 0) {
    notifyError(new Error('请先选择商品'))
    return
  }
  submitting.value = true
  try {
    const created = await orderApi.createCartOrder({
      cart_item_ids: checkedItems.value.map((item) => item.id),
    })
    const orderId = created.order.id
    await orderApi.pay({ order_id: orderId, pay_type: PayType.MOCK })
    await orderApi.mockPaySuccess({ order_id: orderId })
    notifySuccess('下单并支付成功')
    checkedIds.value = []
    await reload()
    await router.push({ name: routeNames.orders })
  } catch (error) {
    notifyError(error, '结算失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  void reload()
})
</script>

<template>
  <div class="cart-page">
    <van-nav-bar title="购物车" />

    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <van-checkbox-group v-model="checkedIds">
        <div v-for="item in cartStore.items" :key="item.id" class="cart-item">
          <van-checkbox :name="item.id" />
          <van-image
            class="cover"
            width="72"
            height="72"
            fit="cover"
            radius="6"
            :src="item.product_cover"
          />
          <div class="meta">
            <div class="name">{{ item.product_name }}</div>
            <div class="price">{{ formatFenToYuan(item.product_price) }}</div>
            <div class="ops">
              <van-stepper
                :model-value="item.quantity"
                integer
                min="0"
                :max="Math.max(item.product_stock, item.quantity)"
                @change="(value: string) => onQuantityChange(item.id, Number(value))"
              />
              <van-button size="mini" plain type="danger" @click="onRemove(item.id)">
                删除
              </van-button>
            </div>
          </div>
        </div>
      </van-checkbox-group>

      <van-empty v-if="!cartStore.loading && cartStore.isEmpty" description="购物车空空如也">
        <van-button round type="primary" @click="router.push({ name: routeNames.home })">
          去逛逛
        </van-button>
      </van-empty>
    </van-pull-refresh>

    <van-submit-bar
      v-if="!cartStore.isEmpty"
      :price="checkedAmountFen"
      button-text="结算"
      :loading="submitting"
      safe-area-inset-bottom
      @submit="checkout"
    />
  </div>
</template>

<style scoped>
.cart-page {
  padding-bottom: 100px;
}

.cart-item {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  margin: 10px 12px;
  padding: 12px;
  background: #fff;
  border-radius: 10px;
}

.meta {
  flex: 1;
  min-width: 0;
}

.name {
  margin-bottom: 6px;
  overflow: hidden;
  font-size: 14px;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.ops {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8px;
}
</style>
