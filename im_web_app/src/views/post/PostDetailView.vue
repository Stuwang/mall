<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { showConfirmDialog } from 'vant'
import { useRoute, useRouter } from 'vue-router'
import type { CommentItem, PostItem } from '@/api/types'
import PostImageGrid from '@/components/PostImageGrid.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { formatTime, newClientCommentId } from '@/utils/conv'
import { notifyError, notifySuccess } from '@/utils/notify'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const api = appContext.api

const postId = computed(() => Number(route.params.postId))
const isSelf = computed(() => post.value?.user_id === (authStore.user?.id ?? 0))

const loading = ref(true)
const post = ref<PostItem | null>(null)
const comments = ref<CommentItem[]>([])
const hasMore = ref(false)
const loadingMore = ref(false)
const draft = ref('')
const sending = ref(false)
const liking = ref(false)

async function loadAll(): Promise<void> {
  loading.value = true
  try {
    const [detail, cmt] = await Promise.all([
      api.post.detail({ post_id: postId.value }),
      api.post.listComment({ post_id: postId.value, after_id: 0, limit: 30 }),
    ])
    post.value = detail.post
    comments.value = cmt.list || []
    hasMore.value = Boolean(cmt.has_more)
  } catch (error) {
    post.value = null
    comments.value = []
    notifyError(error, '加载失败')
  } finally {
    loading.value = false
  }
}

async function loadMoreComments(): Promise<void> {
  if (!hasMore.value || loadingMore.value || comments.value.length === 0) {
    return
  }
  loadingMore.value = true
  try {
    const afterId = comments.value[comments.value.length - 1]?.comment_id ?? 0
    const reply = await api.post.listComment({
      post_id: postId.value,
      after_id: afterId,
      limit: 30,
    })
    comments.value = comments.value.concat(reply.list || [])
    hasMore.value = Boolean(reply.has_more)
  } catch (error) {
    notifyError(error, '加载评论失败')
  } finally {
    loadingMore.value = false
  }
}

async function toggleLike(): Promise<void> {
  if (!post.value || liking.value) {
    return
  }
  const p = post.value
  const prev = { liked: p.liked, like_count: p.like_count }
  p.liked = !p.liked
  p.like_count = Math.max(0, p.like_count + (p.liked ? 1 : -1))
  liking.value = true
  try {
    const reply = p.liked
      ? await api.post.like({ post_id: p.post_id })
      : await api.post.unlike({ post_id: p.post_id })
    if (reply.post) {
      post.value = reply.post
    }
  } catch (error) {
    p.liked = prev.liked
    p.like_count = prev.like_count
    notifyError(error, '操作失败')
  } finally {
    liking.value = false
  }
}

async function sendComment(): Promise<void> {
  const body = draft.value.trim()
  if (!body) {
    return
  }
  sending.value = true
  try {
    const reply = await api.post.createComment({
      post_id: postId.value,
      client_comment_id: newClientCommentId(),
      text: body,
    })
    if (reply.comment) {
      comments.value = comments.value.concat(reply.comment)
    }
    if (reply.post) {
      post.value = reply.post
    }
    draft.value = ''
  } catch (error) {
    notifyError(error, '评论失败')
  } finally {
    sending.value = false
  }
}

async function onDeletePost(): Promise<void> {
  try {
    await showConfirmDialog({ title: '删除这条帖子？' })
  } catch {
    return
  }
  try {
    await api.post.delete({ post_id: postId.value })
    notifySuccess('已删除')
    void router.replace({ name: routeNames.posts })
  } catch (error) {
    notifyError(error, '删除失败')
  }
}

async function onDeleteComment(c: CommentItem): Promise<void> {
  try {
    await showConfirmDialog({ title: '删除这条评论？' })
  } catch {
    return
  }
  try {
    const reply = await api.post.deleteComment({ comment_id: c.comment_id })
    comments.value = comments.value.filter((x) => x.comment_id !== c.comment_id)
    if (reply.post) {
      post.value = reply.post
    }
    notifySuccess('已删除')
  } catch (error) {
    notifyError(error, '删除失败')
  }
}

function openProfile(userId: number): void {
  void router.push({ name: routeNames.profile, params: { userId: String(userId) } })
}

function displayName(nickname: string, userId: number): string {
  return nickname || `用户${userId}`
}

onMounted(() => {
  void loadAll()
})

watch(postId, () => {
  void loadAll()
})
</script>

<template>
  <div class="page detail-page">
    <van-nav-bar title="帖子" left-arrow fixed placeholder @click-left="router.back()">
      <template v-if="isSelf" #right>
        <span class="nav-danger" @click="onDeletePost">删除</span>
      </template>
    </van-nav-bar>

    <div v-if="loading" class="state muted">加载中…</div>
    <template v-else-if="post">
      <div class="content">
        <button type="button" class="author" @click="openProfile(post.user_id)">
          <UserAvatar
            :user-id="post.user_id"
            :nickname="displayName(post.nickname, post.user_id)"
            :src="post.avatar"
            :size="40"
          />
          <div class="author-meta">
            <div class="name">{{ displayName(post.nickname, post.user_id) }}</div>
            <div class="muted">{{ formatTime(post.ct) }}</div>
          </div>
        </button>
        <div v-if="post.text" class="text">{{ post.text }}</div>
        <PostImageGrid v-if="post.image_urls?.length" :images="post.image_urls" />
        <div class="stats">
          <button type="button" class="stat" @click="toggleLike">
            <van-icon :name="post.liked ? 'good-job' : 'good-job-o'" />
            {{ post.like_count }}
          </button>
          <span class="stat">
            <van-icon name="chat-o" />
            {{ post.comment_count }}
          </span>
        </div>
      </div>

      <div class="section-title">评论</div>
      <van-empty v-if="comments.length === 0" description="还没有评论" />
      <div v-else class="comments">
        <div v-for="c in comments" :key="c.comment_id" class="cmt">
          <button type="button" class="cmt-avatar" @click="openProfile(c.user_id)">
            <UserAvatar
              :user-id="c.user_id"
              :nickname="displayName(c.nickname, c.user_id)"
              :src="c.avatar"
              :size="32"
            />
          </button>
          <div class="cmt-body">
            <div class="cmt-head">
              <span class="cmt-name" @click="openProfile(c.user_id)">
                {{ displayName(c.nickname, c.user_id) }}
              </span>
              <span class="muted">{{ formatTime(c.ct) }}</span>
            </div>
            <div class="cmt-text">{{ c.text }}</div>
            <button
              v-if="c.user_id === authStore.user?.id"
              type="button"
              class="cmt-del"
              @click="onDeleteComment(c)"
            >
              删除
            </button>
          </div>
        </div>
        <div v-if="hasMore" class="more">
          <van-button size="small" plain :loading="loadingMore" @click="loadMoreComments">
            加载更多
          </van-button>
        </div>
      </div>
    </template>

    <div v-if="post" class="composer">
      <van-field v-model="draft" maxlength="300" placeholder="写评论…" />
      <van-button size="small" type="primary" :loading="sending" @click="sendComment">发送</van-button>
    </div>
  </div>
</template>

<style scoped>
.detail-page {
  padding-bottom: 64px;
}

.nav-danger {
  color: #ee0a24;
  font-size: 15px;
  cursor: pointer;
}

.state {
  padding: 48px 16px;
  text-align: center;
}

.content {
  margin: 12px;
  padding: 16px;
  border-radius: var(--im-radius);
  background: var(--im-card);
  border: 1px solid var(--im-border);
  box-shadow: var(--im-shadow-sm);
}

.author {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  margin: 0 0 12px;
  padding: 0;
  border: none;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.author-meta {
  min-width: 0;
}

.name {
  font-size: 15px;
  font-weight: 600;
}

.text {
  font-size: 16px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

.stats {
  display: flex;
  gap: 16px;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid var(--im-border);
  color: var(--im-muted);
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
  font-size: 14px;
}

.section-title {
  padding: 4px 16px 8px;
  font-size: 14px;
  color: #969799;
}

.comments {
  padding: 0 12px 24px;
}

.cmt {
  display: flex;
  gap: 10px;
  margin: 0 0 10px;
  padding: 12px;
  border-radius: var(--im-radius-sm);
  background: var(--im-card);
  border: 1px solid var(--im-border);
  box-shadow: var(--im-shadow-sm);
}

.cmt-avatar {
  flex-shrink: 0;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
}

.cmt-body {
  flex: 1;
  min-width: 0;
}

.cmt-head {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}

.cmt-name {
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.cmt-text {
  font-size: 14px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.cmt-del {
  margin-top: 6px;
  border: none;
  background: transparent;
  color: #ee0a24;
  font-size: 12px;
  padding: 0;
  cursor: pointer;
}

.more {
  display: flex;
  justify-content: center;
  padding: 8px 0 16px;
}

.composer {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px calc(10px + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border-top: 1px solid var(--im-border);
}
</style>
