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

function pickPhotos(): void {
  fileInput.value?.click()
}

async function onPhotosSelected(ev: Event): Promise<void> {
  const input = ev.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  if (files.length === 0) {
    return
  }
  const picked = files.filter((f) => f.type.startsWith('image/'))
  if (picked.length === 0) {
    notifyError(null, '请选择图片文件')
    return
  }
  adding.value = true
  let ok = 0
  try {
    for (const file of picked) {
      const uploaded = await api.auth.uploadMedia(file)
      if (!uploaded?.uri) {
        notifyError(null, '上传未返回文件地址')
        continue
      }
      const reply = await api.profile.addAlbumPhoto({ image_uri: uploaded.uri })
      if (reply.photo) {
        photos.value = [reply.photo, ...photos.value]
        ok += 1
      }
    }
    if (ok > 0) {
      notifySuccess(ok === 1 ? '已添加' : `已添加 ${ok} 张`)
    }
  } catch (error) {
    notifyError(error, '添加失败')
  } finally {
    adding.value = false
  }
}

async function onDelete(photo: AlbumPhoto, ev: Event): Promise<void> {
  ev.stopPropagation()
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
      <van-button round block type="primary" :loading="adding" @click="pickPhotos">添加照片</van-button>
      <input
        ref="fileInput"
        class="hidden-input"
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp"
        multiple
        @change="onPhotosSelected"
      />
    </div>

    <div v-if="loading" class="state muted">加载中…</div>
    <van-empty v-else-if="photos.length === 0" description="还没有照片" />
    <template v-else>
      <div class="grid">
        <div v-for="p in photos" :key="p.id" class="cell">
          <button type="button" class="thumb" @click="openPreview(p.image_url)">
            <img :src="p.image_url" alt="" />
          </button>
          <button
            v-if="isSelf"
            type="button"
            class="del"
            aria-label="删除"
            @click="onDelete(p, $event)"
          >
            <van-icon name="cross" />
          </button>
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

.grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
  margin: 8px 12px 16px;
  padding: 8px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.55);
  border: 1px solid var(--im-border);
}

.cell {
  position: relative;
  aspect-ratio: 1;
}

.thumb {
  width: 100%;
  height: 100%;
  padding: 0;
  border: none;
  border-radius: 8px;
  overflow: hidden;
  background: #dfe5ec;
  cursor: pointer;
}

.thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.del {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 22px;
  height: 22px;
  border: none;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  cursor: pointer;
}

.more {
  display: flex;
  justify-content: center;
  padding: 0 16px 32px;
}
</style>
