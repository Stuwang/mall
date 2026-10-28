<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { AlbumPhoto, GetProfileReply, PostItem } from '@/api/types'
import PostImageGrid from '@/components/PostImageGrid.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useImLive, useImStore } from '@/stores/imStore'
import { formatTime } from '@/utils/conv'
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
const posts = ref<PostItem[]>([])
const postsHasMore = ref(false)
const loadingMorePosts = ref(false)

const displayName = computed(
  () => profile.value?.nickname || `用户${profile.value?.user_id || userId.value}`,
)

async function loadAll(): Promise<void> {
  loading.value = true
  try {
    const [p, album, userPosts] = await Promise.all([
      api.profile.getProfile({ user_id: userId.value }),
      api.profile.listAlbum({ user_id: userId.value, before_id: 0, limit: 9 }),
      api.post.listUser({ user_id: userId.value, before_id: 0, limit: 20 }),
    ])
    profile.value = p
    photos.value = album.list || []
    posts.value = userPosts.list || []
    postsHasMore.value = Boolean(userPosts.has_more)
  } catch (error) {
    profile.value = null
    photos.value = []
    posts.value = []
    postsHasMore.value = false
    notifyError(error, '加载主页失败')
  } finally {
    loading.value = false
  }
}

async function loadMorePosts(): Promise<void> {
  if (!postsHasMore.value || loadingMorePosts.value || posts.value.length === 0) {
    return
  }
  loadingMorePosts.value = true
  try {
    const beforeId = posts.value[posts.value.length - 1]?.post_id ?? 0
    const reply = await api.post.listUser({
      user_id: userId.value,
      before_id: beforeId,
      limit: 20,
    })
    posts.value = posts.value.concat(reply.list || [])
    postsHasMore.value = Boolean(reply.has_more)
  } catch (error) {
    notifyError(error, '加载更多失败')
  } finally {
    loadingMorePosts.value = false
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

async function toggleLike(p: PostItem, ev: Event): Promise<void> {
  ev.stopPropagation()
  const prev = { liked: p.liked, like_count: p.like_count }
  p.liked = !p.liked
  p.like_count = Math.max(0, p.like_count + (p.liked ? 1 : -1))
  try {
    const reply = p.liked
      ? await api.post.like({ post_id: p.post_id })
      : await api.post.unlike({ post_id: p.post_id })
    if (reply.post) {
      Object.assign(p, reply.post)
    }
  } catch (error) {
    p.liked = prev.liked
    p.like_count = prev.like_count
    notifyError(error, '操作失败')
  }
}

function openDetail(postId: number): void {
  void router.push({ name: routeNames.postDetail, params: { postId: String(postId) } })
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
      <div class="hero im-fade-up">
        <div class="hero-bg" aria-hidden="true" />
        <UserAvatar
          :user-id="profile.user_id"
          :nickname="displayName"
          :src="profile.avatar"
          :size="88"
        />
        <div class="name">{{ displayName }}</div>
        <div class="uid">用户 ID {{ profile.user_id }}</div>
        <div v-if="!isSelf" class="actions">
          <van-button round type="primary" @click="startChat">发消息</van-button>
        </div>
      </div>

      <button type="button" class="im-section-title" @click="openAlbum">
        <span>相册</span>
        <span class="section-more">{{ profile.photo_count }} 张 · 全部</span>
      </button>

      <van-empty v-if="photos.length === 0" description="还没有照片" />
      <div v-else class="album-grid im-fade-up">
        <button
          v-for="p in photos"
          :key="p.id"
          type="button"
          class="album-cell"
          @click="openAlbum"
        >
          <img :src="p.image_url" alt="" />
        </button>
      </div>

      <div class="im-section-title static">
        <span>帖子</span>
      </div>

      <van-empty v-if="posts.length === 0" description="还没有帖子" />
      <div v-else class="post-list">
        <button
          v-for="p in posts"
          :key="p.post_id"
          type="button"
          class="card"
          @click="openDetail(p.post_id)"
        >
          <div class="post-meta muted">{{ formatTime(p.ct) }}</div>
          <div v-if="p.text" class="text">{{ p.text }}</div>
          <PostImageGrid v-if="p.image_urls?.length" :images="p.image_urls" />
          <div class="stats">
            <button type="button" class="stat" @click="toggleLike(p, $event)">
              <van-icon :name="p.liked ? 'good-job' : 'good-job-o'" />
              {{ p.like_count }}
            </button>
            <span class="stat">
              <van-icon name="chat-o" />
              {{ p.comment_count }}
            </span>
          </div>
        </button>
        <div v-if="postsHasMore" class="more">
          <van-button size="small" plain :loading="loadingMorePosts" @click="loadMorePosts">
            加载更多
          </van-button>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.state {
  padding: 48px 16px;
  text-align: center;
}

.hero {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  margin: 12px 16px 4px;
  padding: 28px 16px 22px;
  overflow: hidden;
  border-radius: 22px;
  background: var(--im-card);
  border: 1px solid var(--im-border);
  box-shadow: var(--im-shadow);
}

.hero-bg {
  position: absolute;
  inset: 0 0 auto 0;
  height: 96px;
  background:
    linear-gradient(135deg, rgba(15, 143, 134, 0.18), rgba(232, 135, 58, 0.12));
  pointer-events: none;
}

.name {
  position: relative;
  max-width: 100%;
  margin-top: 4px;
  font-size: 22px;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.uid {
  position: relative;
  font-size: 12px;
  color: var(--im-muted);
}

.actions {
  position: relative;
  margin-top: 10px;
  width: 100%;
  max-width: 200px;
}

.actions :deep(.van-button) {
  width: 100%;
}

.section-more {
  font-size: 12px;
  font-weight: 500;
  color: var(--im-muted);
}

.im-section-title {
  width: 100%;
  border: none;
  background: transparent;
  cursor: pointer;
}

.im-section-title.static {
  cursor: default;
}

.album-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
  margin: 0 16px 8px;
  padding: 6px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.55);
  border: 1px solid var(--im-border);
}

.album-cell {
  aspect-ratio: 1;
  padding: 0;
  border: none;
  border-radius: 8px;
  overflow: hidden;
  background: #dfe5ec;
  cursor: pointer;
}

.album-cell img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 0.25s ease;
}

.album-cell:active img {
  transform: scale(1.04);
}

.post-list {
  padding: 0 12px 24px;
}

.card {
  display: block;
  width: 100%;
  margin: 0 0 12px;
  padding: 16px;
  border: 1px solid var(--im-border);
  border-radius: var(--im-radius);
  background: var(--im-card);
  text-align: left;
  cursor: pointer;
  box-shadow: var(--im-shadow-sm);
}

.post-meta {
  margin-bottom: 8px;
  font-size: 12px;
}

.text {
  font-size: 15px;
  line-height: 1.55;
  color: var(--im-text);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 4;
  overflow: hidden;
  white-space: pre-wrap;
  word-break: break-word;
}

.stats {
  display: flex;
  gap: 18px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--im-border);
  color: var(--im-muted);
  font-size: 13px;
}

.stat {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: none;
  background: transparent;
  color: inherit;
  padding: 0;
  cursor: pointer;
  font-size: 13px;
}

.more {
  display: flex;
  justify-content: center;
  padding: 8px 0 16px;
}
</style>
