<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { CoinChangeGiftSend, CoinChangeRecharge, type LedgerItem } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { formatCoinChange } from '@/utils/coin'
import { formatTime, newClientOrderId } from '@/utils/conv'
import { notifyError, notifyFail } from '@/utils/notify'

const packs = [10, 50, 100, 500, 1000] as const
const ledgerLimit = 20

const router = useRouter()
const api = appContext.api

const balance = ref<number | null>(null)
const coinsPerYuan = ref<number | null>(null)
const selectedPack = ref<number>(100)
const customCoins = ref('')
const creating = ref(false)
const refreshing = ref(false)
const ledgerLoading = ref(false)
const loadingMore = ref(false)
const ledger = ref<LedgerItem[]>([])
const hasMore = ref(false)

function parsePositiveInt(raw: string): number | null {
  const text = raw.trim()
  if (!/^[1-9]\d*$/.test(text)) {
    return null
  }
  const n = Number(text)
  if (!Number.isSafeInteger(n) || n <= 0) {
    return null
  }
  return n
}

function resolveCoins(): number | null {
  if (customCoins.value.trim()) {
    return parsePositiveInt(customCoins.value)
  }
  return selectedPack.value > 0 ? selectedPack.value : null
}

function selectPack(coins: number): void {
  selectedPack.value = coins
  customCoins.value = ''
}

function typeLabel(changeType: number): string {
  if (changeType === CoinChangeRecharge) {
    return '充值'
  }
  if (changeType === CoinChangeGiftSend) {
    return '送礼'
  }
  return '金币变动'
}

async function reload(): Promise<void> {
  const [bal, page] = await Promise.all([
    api.coin.balance(),
    api.coin.ledger({ limit: ledgerLimit }),
  ])
  balance.value = bal.balance
  coinsPerYuan.value = bal.coins_per_yuan
  ledger.value = page.list || []
  hasMore.value = Boolean(page.has_more)
}

async function loadFirst(): Promise<void> {
  ledgerLoading.value = true
  try {
    await reload()
  } catch (error) {
    notifyError(error, '加载金币失败')
  } finally {
    ledgerLoading.value = false
  }
}

async function onRefresh(): Promise<void> {
  refreshing.value = true
  try {
    await reload()
  } catch (error) {
    notifyError(error, '加载金币失败')
  } finally {
    refreshing.value = false
  }
}

async function loadMore(): Promise<void> {
  if (!hasMore.value || loadingMore.value || ledger.value.length === 0) {
    return
  }
  const beforeId = ledger.value[ledger.value.length - 1]?.id
  if (!beforeId) {
    return
  }
  loadingMore.value = true
  try {
    const page = await api.coin.ledger({ before_id: beforeId, limit: ledgerLimit })
    ledger.value = ledger.value.concat(page.list || [])
    hasMore.value = Boolean(page.has_more)
  } catch (error) {
    notifyError(error, '加载更多失败')
  } finally {
    loadingMore.value = false
  }
}

async function onRecharge(): Promise<void> {
  const coins = resolveCoins()
  if (coins == null) {
    notifyFail('请输入正整数金币数量')
    return
  }
  if (creating.value) {
    return
  }
  creating.value = true
  try {
    const order = await api.coin.createRecharge({
      coins,
      client_order_id: newClientOrderId(),
    })
    await router.push({
      name: routeNames.coinPay,
      params: { orderId: String(order.id) },
    })
  } catch (error) {
    notifyError(error, '创建订单失败')
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  void loadFirst()
})
</script>

<template>
  <div class="page coin-page">
    <van-nav-bar title="我的金币" left-arrow fixed placeholder @click-left="router.back()" />

    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <section class="balance-card im-fade-up">
        <div class="label">当前金币</div>
        <div class="amount">{{ balance == null ? '—' : balance }}</div>
        <div class="rate">1 元 = {{ coinsPerYuan == null ? '—' : coinsPerYuan }} 金币</div>
      </section>

      <section class="recharge im-card im-fade-up">
        <div class="section-label">选择充值数量</div>
        <div class="packs">
          <button
            v-for="n in packs"
            :key="n"
            type="button"
            class="pack"
            :class="{ on: !customCoins.trim() && selectedPack === n }"
            @click="selectPack(n)"
          >
            {{ n }}
          </button>
        </div>
        <van-field
          v-model="customCoins"
          type="digit"
          maxlength="9"
          placeholder="自定义正整数"
          class="custom"
        />
        <van-button type="primary" block :loading="creating" @click="onRecharge">充值</van-button>
      </section>

      <div class="im-section-title">
        <span>明细</span>
      </div>

      <div v-if="ledgerLoading && ledger.length === 0" class="state muted">加载中…</div>
      <van-empty v-else-if="ledger.length === 0" description="暂无明细" />
      <div v-else class="ledger">
        <div v-for="item in ledger" :key="item.id" class="ledger-item im-card">
          <div class="ledger-main">
            <div class="remark">{{ item.remark || typeLabel(item.change_type) }}</div>
            <div class="muted">{{ formatTime(item.ct) }}</div>
          </div>
          <div class="delta" :class="{ plus: item.change > 0, minus: item.change < 0 }">
            {{ formatCoinChange(item.change) }}
          </div>
        </div>
        <div v-if="hasMore" class="more">
          <van-button size="small" plain :loading="loadingMore" @click="loadMore">
            加载更多
          </van-button>
        </div>
      </div>
    </van-pull-refresh>
  </div>
</template>

<style scoped>
.coin-page {
  padding-bottom: calc(24px + var(--im-safe-bottom));
}

.balance-card {
  margin: 16px 16px 12px;
  padding: 22px 20px 18px;
  border-radius: 22px;
  color: #fff;
  background:
    radial-gradient(120px 80px at 90% 0%, rgba(255, 255, 255, 0.28), transparent 70%),
    linear-gradient(145deg, #e8a23a, #d4782a 55%, #0f8f86);
  box-shadow: 0 16px 32px rgba(212, 120, 42, 0.28);
}

.label {
  font-size: 13px;
  opacity: 0.86;
}

.amount {
  margin-top: 4px;
  font-family: Outfit, 'Noto Sans SC', sans-serif;
  font-size: 48px;
  font-weight: 700;
  line-height: 1.1;
  letter-spacing: -0.03em;
}

.rate {
  margin-top: 8px;
  font-size: 13px;
  opacity: 0.9;
}

.recharge {
  margin: 0 16px;
  padding: 16px;
}

.section-label {
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 700;
}

.packs {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.pack {
  height: 44px;
  border: 1px solid var(--im-border);
  border-radius: 12px;
  background: #f6f8fa;
  color: var(--im-text);
  font-size: 16px;
  font-weight: 700;
  font-family: Outfit, 'Noto Sans SC', sans-serif;
  cursor: pointer;
}

.pack.on {
  border-color: transparent;
  background: var(--im-primary);
  color: #fff;
  box-shadow: 0 8px 16px rgba(15, 143, 134, 0.22);
}

.custom {
  margin: 12px 0;
  padding: 0;
  overflow: hidden;
  border-radius: 12px;
  background: #f3f5f8;
}

.recharge :deep(.van-button) {
  margin-top: 4px;
}

.ledger {
  padding: 0 16px 8px;
}

.ledger-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
  padding: 14px 16px;
}

.ledger-main {
  min-width: 0;
}

.remark {
  font-size: 15px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.delta {
  flex-shrink: 0;
  font-family: Outfit, 'Noto Sans SC', sans-serif;
  font-size: 18px;
  font-weight: 700;
}

.delta.plus {
  color: var(--im-success);
}

.delta.minus {
  color: var(--im-danger);
}

.state,
.more {
  display: flex;
  justify-content: center;
  padding: 12px 0 20px;
}
</style>
