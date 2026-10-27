<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { showConfirmDialog } from 'vant'
import { useRoute, useRouter } from 'vue-router'
import type { AlbumPhoto } from '@/api/types'
import ImageViewer from '@/components/ImageViewer.vue'
import { appContext } from '@/app/AppContext'
import { useAuthStore } from '@/stores/authStore'
import { notifyError, notifySuccess } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const api = appContext.api

const userId = computed(() => Number(route.params.userId))
const isSelf = computed(() => userId.value === (authStore.user?.id ?? 0))

const loading = ref(true)
const loadingMore = ref(false)
const adding = ref(false)
const photos = ref<AlbumPhoto[]>([])
const hasMore = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const previewUrl = ref('')
const previewVisible = ref(false)

async function loadFirst(): Promise<void> {
  loading.value = true
  try {
    const album = await api.profile.listAlbum({
      user_id: userId.value,
      before_id: 0,
      limit: 30,
    })
    photos.value = album.list || []
    hasMore.value = Boolean(album.has_more)
  } catch (error) {
    photos.value = []
    hasMore.value = false
    notifyError(error, '加载相册失败')
  } finally {
    loading.value = false
  }
}

async function loadMore(): Promise<void> {
  if (!hasMore.value || loadingMore.value || photos.value.length === 0) {
    return
  }
  loadingMore.value = true
  try {
    const beforeId = photos.value[photos.value.length - 1]?.id ?? 0
    const album = await api.profile.listAlbum({
      user_id: userId.value,
      before_id: beforeId,
      limit: 30,
    })
    photos.value = photos.value.concat(album.list || [])
    hasMore.value = Boolean(album.has_more)
  } catch (error) {
    notifyError(error, '加载更多失败')
  } finally {
    loadingMore.value = false
  }
}

function pickPhoto(): void {
  fileInput.value?.click()
}

async function onPhotoSelected(ev: Event): Promise<void> {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) {
    return
  }
  if (!file.type.startsWith('image/')) {
    notifyError(null, '请选择图片文件')
    return
  }
  adding.value = true
  try {
    const uploaded = await api.auth.uploadMedia(file)
    if (!uploaded?.uri) {
      notifyError(null, '上传未返回文件地址')
      return
    }
    const reply = await api.profile.addAlbumPhoto({ image_uri: uploaded.uri })
    if (reply.photo) {
      photos.value = [reply.photo, ...photos.value]
    }
    notifySuccess('已添加')
  } catch (error) {
    notifyError(error, '添加失败')
  } finally {
    adding.value = false
  }
}

async function onDelete(photo: AlbumPhoto): Promise<void> {
  try {
    await showConfirmDialog({ title: '删除这张照片？' })
  } catch {
    return
  }
  try {
    await api.profile.deleteAlbumPhoto({ photo_id: photo.id })
    photos.value = photos.value.filter((p) => p.id !== photo.id)
    notifySuccess('已删除')
  } catch (error) {
    notifyError(error, '删除失败')
  }
}

function openPreview(url: string): void {
  if (!url) {
    return
  }
  previewUrl.value = url
  previewVisible.value = true
}

onMounted(() => {
  void loadFirst()
})

watch(userId, () => {
  void loadFirst()
})
</script>

<template>
  <div class="page">
    <van-nav-bar title="相册" left-arrow fixed placeholder @click-left="router.back()" />

    <div v-if="isSelf" class="actions">
      <van-button round block type="primary" :loading="adding" @click="pickPhoto">添加照片</van-button>
      <input
        ref="fileInput"
        class="hidden-input"
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp"
        @change="onPhotoSelected"
      />
    </div>

    <div v-if="loading" class="state muted">加载中…</div>
    <van-empty v-else-if="photos.length === 0" description="还没有照片" />
    <template v-else>
      <div class="list">
        <div v-for="p in photos" :key="p.id" class="row">
          <button type="button" class="thumb" @click="openPreview(p.image_url)">
            <img :src="p.image_url" alt="" />
          </button>
          <div class="body">
            <div class="caption">{{ p.caption || '无说明' }}</div>
          </div>
          <button v-if="isSelf" type="button" class="del" @click="onDelete(p)">删除</button>
        </div>
      </div>
      <div v-if="hasMore" class="more">
        <van-button size="small" plain :loading="loadingMore" @click="loadMore">加载更多</van-button>
      </div>
    </template>

    <ImageViewer v-model:show="previewVisible" :src="previewUrl" />
  </div>
</template>

<style scoped>
.state {
  padding: 48px 16px;
  text-align: center;
}

.actions {
  padding: 12px 16px 4px;
}

.hidden-input {
  display: none;
}

.list {
  padding: 8px 12px 16px;
}

.row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 8px;
  padding: 10px;
  border-radius: 12px;
  background: #fff;
}

.thumb {
  flex-shrink: 0;
  width: 72px;
  height: 72px;
  padding: 0;
  border: none;
  border-radius: 8px;
  overflow: hidden;
  background: #f2f3f5;
  cursor: pointer;
}

.thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.body {
  flex: 1;
  min-width: 0;
}

.caption {
  font-size: 14px;
  color: #323233;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.del {
  flex-shrink: 0;
  border: none;
  background: transparent;
  color: #ee0a24;
  font-size: 14px;
  cursor: pointer;
}

.more {
  display: flex;
  justify-content: center;
  padding: 0 16px 32px;
}
</style>
