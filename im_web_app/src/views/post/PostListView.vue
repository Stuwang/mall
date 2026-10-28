<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import type { PostItem } from '@/api/types'
import PostImageGrid from '@/components/PostImageGrid.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { formatTime } from '@/utils/conv'
import { notifyError } from '@/utils/notify'

const router = useRouter()
const api = appContext.api

const refreshing = ref(false)
const loadingMore = ref(false)
const loading = ref(true)
const posts = ref<PostItem[]>([])
const hasMore = ref(false)

async function loadFirst(): Promise<void> {
  loading.value = true
  try {
    const reply = await api.post.recommend({ has_cursor: false, limit: 20 })
    posts.value = reply.list || []
    hasMore.value = Boolean(reply.has_more)
  } catch (error) {
    posts.value = []
    hasMore.value = false
    notifyError(error, '加载失败')
  } finally {
    loading.value = false
  }
}

async function onRefresh(): Promise<void> {
  refreshing.value = true
  try {
    const reply = await api.post.recommend({ has_cursor: false, limit: 20 })
    posts.value = reply.list || []
    hasMore.value = Boolean(reply.has_more)
  } catch (error) {
    notifyError(error, '刷新失败')
  } finally {
    refreshing.value = false
  }
}

async function loadMore(): Promise<void> {
  if (!hasMore.value || loadingMore.value || posts.value.length === 0) {
    return
  }
  loadingMore.value = true
  try {
    const last = posts.value[posts.value.length - 1]
    const reply = await api.post.recommend({
      has_cursor: true,
      cursor_id: last?.post_id ?? 0,
      limit: 20,
    })
    posts.value = posts.value.concat(reply.list || [])
    hasMore.value = Boolean(reply.has_more)
  } catch (error) {
    notifyError(error, '加载更多失败')
  } finally {
    loadingMore.value = false
  }
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

function openProfile(userId: number, ev: Event): void {
  ev.stopPropagation()
  void router.push({ name: routeNames.profile, params: { userId: String(userId) } })
}

function openCreate(): void {
  void router.push({ name: routeNames.postCreate })
}

function displayName(p: PostItem): string {
  return p.nickname || `用户${p.user_id}`
}

onMounted(() => {
  void loadFirst()
})
</script>

<template>
  <div class="page">
    <van-nav-bar title="帖子" fixed placeholder>
      <template #right>
        <span class="nav-action" @click="openCreate">发布</span>
      </template>
    </van-nav-bar>

    <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
      <div v-if="loading" class="state muted">加载中…</div>
      <van-empty v-else-if="posts.length === 0" description="还没有帖子" />
      <div v-else class="list">
        <article
          v-for="p in posts"
          :key="p.post_id"
          class="card"
          role="button"
          tabindex="0"
          @click="openDetail(p.post_id)"
          @keydown.enter.prevent="openDetail(p.post_id)"
        >
          <div class="author" @click="openProfile(p.user_id, $event)">
            <UserAvatar
              :user-id="p.user_id"
              :nickname="displayName(p)"
              :src="p.avatar"
              :size="36"
            />
            <div class="author-meta">
              <div class="name">{{ displayName(p) }}</div>
              <div class="muted">{{ formatTime(p.ct) }}</div>
            </div>
          </div>
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
        </article>
        <div v-if="hasMore" class="more">
          <van-button size="small" plain :loading="loadingMore" @click="loadMore">加载更多</van-button>
        </div>
      </div>
    </van-pull-refresh>
  </div>
</template>

<style scoped>
.nav-action {
  color: var(--im-primary);
  font-size: 14px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 999px;
  background: var(--im-primary-soft);
  cursor: pointer;
}

.state {
  padding: 48px 16px;
  text-align: center;
}

.list {
  padding: 10px 12px 16px;
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
  /* 长列表用 content-visibility，离屏卡不参与布局/绘制 */
  content-visibility: auto;
  contain-intrinsic-size: auto 220px;
}

.author {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.author-meta {
  min-width: 0;
}

.name {
  font-size: 15px;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  margin-top: 14px;
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
  font-weight: 500;
}

.more {
  display: flex;
  justify-content: center;
  padding: 8px 0 24px;
}
</style>
