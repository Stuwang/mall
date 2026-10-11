<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { appContext } from '@/app/AppContext'
import { AssetType } from '@/api/types'
import { routeNames } from '@/router/routes'
import { formatFenToYuan } from '@/utils/format'
import { notifyError } from '@/utils/notify'
import { pointsToCents } from '@/utils/points'

const router = useRouter()
const mallApi = appContext.api.mall
const assetApi = appContext.api.assets
const orderApi = appContext.api.orders

const packages = [100, 500, 1000, 2000, 5000] as const

const loading = ref(true)
const submitting = ref(false)
const pointsPerYuan = ref(100)
const balance = ref(0)
const selected = ref<number>(500)
const customPoints = ref<number | undefined>(undefined)

const buyPoints = computed(() => {
  if (customPoints.value && customPoints.value > 0) {
    return Math.floor(customPoints.value)
  }
  return selected.value
})

const payFen = computed(() => pointsToCents(buyPoints.value, pointsPerYuan.value))

async function load(): Promise<void> {
  loading.value = true
  try {
    const [mall, assets] = await Promise.all([
      mallApi.getMallInfo({}),
      assetApi.getUserAssets({ type: AssetType.POINT }),
    ])
    pointsPerYuan.value = mall.points_per_yuan || 100
    const point = (assets.assets ?? []).find((item) => item.type === AssetType.POINT)
    balance.value = point?.balance ?? 0
  } catch (error) {
    notifyError(error, '加载失败')
  } finally {
    loading.value = false
  }
}

function pickPackage(points: number): void {
  selected.value = points
  customPoints.value = undefined
}

async function onSubmit(): Promise<void> {
  if (buyPoints.value <= 0) {
    notifyError(new Error('请选择或输入要购买的积分'))
    return
  }
  if (payFen.value <= 0) {
    notifyError(new Error('应付金额无效，请检查汇率配置'))
    return
  }
  submitting.value = true
  try {
    const reply = await orderApi.createPointOrder({ points: buyPoints.value })
    await router.push({
      name: routeNames.payCheckout,
      params: { id: String(reply.order.id) },
    })
  } catch (error) {
    notifyError(error, '创建积分订单失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <div class="page">
    <van-nav-bar title="购买积分" left-arrow fixed placeholder @click-left="router.back()" />

    <van-skeleton title :row="8" :loading="loading">
      <div class="hero">
        <div class="label">当前积分</div>
        <div class="balance">{{ balance }}</div>
        <p class="muted">充值比例：1 元 = {{ pointsPerYuan }} 积分</p>
      </div>

      <van-cell-group inset title="选择套餐">
        <div class="packages">
          <button
            v-for="item in packages"
            :key="item"
            type="button"
            class="pkg"
            :class="{ active: !customPoints && selected === item }"
            @click="pickPackage(item)"
          >
            <div class="pkg-points">{{ item }}</div>
            <div class="pkg-price">¥{{ formatFenToYuan(pointsToCents(item, pointsPerYuan)) }}</div>
          </button>
        </div>
      </van-cell-group>

      <van-cell-group inset class="block" title="自定义数量">
        <van-field
          v-model.number="customPoints"
          type="digit"
          label="积分"
          placeholder="输入积分数"
          clearable
        />
      </van-cell-group>

      <van-cell-group inset class="block">
        <van-cell title="将获得积分" :value="String(buyPoints)" />
        <van-cell title="应付金额">
          <template #value>
            <span class="price">{{ formatFenToYuan(payFen) }}</span>
          </template>
        </van-cell>
      </van-cell-group>

      <div class="actions">
        <van-button round block type="danger" :loading="submitting" @click="onSubmit">
          确认购买
        </van-button>
      </div>
    </van-skeleton>
  </div>
</template>

<style scoped>
.hero {
  margin: 12px 16px;
  padding: 24px 16px;
  color: #fff;
  text-align: center;
  background: linear-gradient(145deg, #ff8a3d, #ee0a24 70%);
  border-radius: 14px;
}

.label {
  font-size: 13px;
  opacity: 0.9;
}

.balance {
  margin: 6px 0;
  font-size: 36px;
  font-weight: 700;
}

.hero .muted {
  margin: 0;
  color: rgba(255, 255, 255, 0.88);
}

.packages {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  padding: 12px 16px 16px;
}

.pkg {
  padding: 12px 8px;
  color: inherit;
  text-align: center;
  cursor: pointer;
  background: #f7f8fa;
  border: 1px solid transparent;
  border-radius: 10px;
}

.pkg.active {
  background: #fff1f0;
  border-color: #ee0a24;
}

.pkg-points {
  font-size: 18px;
  font-weight: 700;
}

.pkg-price {
  margin-top: 4px;
  font-size: 12px;
  color: #969799;
}

.block {
  margin-top: 12px;
}

.actions {
  margin: 24px 16px;
}
</style>
