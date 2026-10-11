<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { appContext } from '@/app/AppContext'
import type { ContactInfo } from '@/api/types'
import { notifyError } from '@/utils/notify'

type ContentKind = 'help' | 'about' | 'contact'

const route = useRoute()
const router = useRouter()
const mallApi = appContext.api.mall

const loading = ref(true)
const helpCenter = ref('')
const aboutUs = ref('')
const contact = ref<ContactInfo | null>(null)

const kind = computed<ContentKind>(() => {
  const value = String(route.meta.contentKind || 'help')
  if (value === 'about' || value === 'contact' || value === 'help') return value
  return 'help'
})

const title = computed(() => {
  switch (kind.value) {
    case 'about':
      return '关于我们'
    case 'contact':
      return '联系我们'
    default:
      return '帮助中心'
  }
})

const textBody = computed(() => {
  if (kind.value === 'about') return aboutUs.value || '暂无内容'
  if (kind.value === 'help') return helpCenter.value || '暂无内容'
  return ''
})

async function load(): Promise<void> {
  loading.value = true
  try {
    const reply = await mallApi.getMallInfo({})
    helpCenter.value = reply.help_center || ''
    aboutUs.value = reply.about_us || ''
    contact.value = reply.contact ?? null
  } catch (error) {
    notifyError(error, '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <div class="page">
    <van-nav-bar :title="title" left-arrow fixed placeholder @click-left="router.back()" />

    <van-skeleton title :row="10" :loading="loading">
      <div v-if="kind === 'contact'" class="panel">
        <van-cell-group inset>
          <van-cell title="客服电话" :value="contact?.phone || '-'" />
          <van-cell title="邮箱" :value="contact?.email || '-'" />
          <van-cell title="微信" :value="contact?.wechat || '-'" />
          <van-cell title="地址" :label="contact?.address || '-'" />
        </van-cell-group>
      </div>
      <div v-else class="article">
        <pre>{{ textBody }}</pre>
      </div>
    </van-skeleton>
  </div>
</template>

<style scoped>
.panel {
  padding-top: 12px;
}

.article {
  margin: 12px;
  padding: 16px;
  background: #fff;
  border-radius: 12px;
}

.article pre {
  margin: 0;
  font-family: inherit;
  font-size: 14px;
  line-height: 1.7;
  color: #323233;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
