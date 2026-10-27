<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { AlbumPhoto, GetProfileReply } from '@/api/types'
import UserAvatar from '@/components/UserAvatar.vue'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useImLive, useImStore } from '@/stores/imStore'
import { notifyError } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const imStore = useImStore()
const api = appContext.api
useImLive()

const userId = computed(() => Number(route.params.userId))
const isSelf = computed(() => userId.value === (authStore.user?.id ?? 0))

const loading = ref(true)
const profile = ref<GetProfileReply | null>(null)
const photos = ref<AlbumPhoto[]>([])

const displayName = computed(
  () => profile.value?.nickname || `用户${profile.value?.user_id || userId.value}`,
)

async function loadAll(): Promise<void> {
  loading.value = true
  try {
    const [p, album] = await Promise.all([
      api.profile.getProfile({ user_id: userId.value }),
      api.profile.listAlbum({ user_id: userId.value, before_id: 0, limit: 10 }),
    ])
    profile.value = p
    photos.value = album.list || []
  } catch (error) {
    profile.value = null
    photos.value = []
    notifyError(error, '加载主页失败')
  } finally {
    loading.value = false
  }
}

function openAlbum(): void {
  void router.push({ name: routeNames.album, params: { userId: String(userId.value) } })
}

function startChat(): void {
  if (!profile.value) {
    return
  }
  imStore.openLocalChat(
    profile.value.user_id,
    profile.value.nickname || `用户${profile.value.user_id}`,
    profile.value.avatar || '',
  )
  void router.push({ name: routeNames.chat, params: { peerId: String(profile.value.user_id) } })
}

onMounted(() => {
  void loadAll()
})

watch(userId, () => {
  void loadAll()
})
</script>

<template>
  <div class="page">
    <van-nav-bar title="主页" left-arrow fixed placeholder @click-left="router.back()" />

    <div v-if="loading" class="state muted">加载中…</div>
    <template v-else-if="profile">
      <div class="header">
        <UserAvatar
          :user-id="profile.user_id"
          :nickname="displayName"
          :src="profile.avatar"
          :size="80"
        />
        <div class="name">{{ displayName }}</div>
        <div class="muted">用户 ID {{ profile.user_id }}</div>
        <div class="muted">相册 {{ profile.photo_count }} 张</div>
      </div>

      <div class="actions">
        <van-button v-if="!isSelf" round block type="primary" @click="startChat">发消息</van-button>
      </div>

      <button type="button" class="section" @click="openAlbum">
        <span>相册</span>
        <span class="section-more">查看全部</span>
      </button>

      <van-empty v-if="photos.length === 0" description="还没有照片" />
      <van-swipe
        v-else
        class="album-swipe"
        :loop="false"
        :show-indicators="photos.length > 1"
        indicator-color="#1989fa"
        @click="openAlbum"
      >
        <van-swipe-item v-for="p in photos" :key="p.id">
          <div class="slide">
            <img :src="p.image_url" alt="" />
            <div v-if="p.caption" class="slide-caption">{{ p.caption }}</div>
          </div>
        </van-swipe-item>
      </van-swipe>
    </template>
  </div>
</template>

<style scoped>
.state {
  padding: 48px 16px;
  text-align: center;
}

.header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 24px 16px 8px;
}

.name {
  max-width: 100%;
  font-size: 20px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.actions {
  padding: 12px 16px 4px;
}

.section {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  margin: 8px 0 0;
  padding: 12px 16px;
  border: none;
  background: transparent;
  font-size: 16px;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
}

.section-more {
  font-size: 13px;
  font-weight: 400;
  color: #969799;
}

.album-swipe {
  margin: 0 16px 24px;
  border-radius: 12px;
  overflow: hidden;
  background: #f2f3f5;
  cursor: pointer;
}

.slide {
  position: relative;
  height: 220px;
}

.slide img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.slide-caption {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 8px 12px;
  color: #fff;
  font-size: 13px;
  background: linear-gradient(transparent, rgba(0, 0, 0, 0.55));
}
</style>
