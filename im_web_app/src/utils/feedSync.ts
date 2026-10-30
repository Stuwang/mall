import type { PostItem } from '@/api/types'

/** 缓存页面之间同步帖子、相册、金币变更。每个页面自己记游标，互不消费掉。 */
export type FeedEvent =
  | { id: number; kind: 'reload-posts' }
  | { id: number; kind: 'reload-profile'; userId: number }
  | {
      id: number
      kind: 'patch'
      postId: number
      liked: boolean
      likeCount: number
      commentCount: number
    }
  | { id: number; kind: 'remove'; postId: number }
  | { id: number; kind: 'album'; userId: number }
  | { id: number; kind: 'coin' }

const maxEvents = 40
let seq = 0
const events: FeedEvent[] = []

type NewFeedEvent = FeedEvent extends infer E ? (E extends { id: number } ? Omit<E, 'id'> : never) : never

function push(event: NewFeedEvent): void {
  seq += 1
  events.push({ ...event, id: seq } as FeedEvent)
  if (events.length > maxEvents) {
    events.splice(0, events.length - maxEvents)
  }
}

export function feedCursor(): number {
  return seq
}

export function eventsSince(afterId: number): { missed: boolean; events: FeedEvent[] } {
  const missed = events.length > 0 && events[0].id > afterId + 1
  return {
    missed,
    events: events.filter((e) => e.id > afterId),
  }
}

export function reloadPosts(): void {
  push({ kind: 'reload-posts' })
}

export function reloadProfilePosts(userId: number): void {
  if (userId <= 0) {
    return
  }
  push({ kind: 'reload-profile', userId })
}

export function patchPost(post: Pick<PostItem, 'post_id' | 'liked' | 'like_count' | 'comment_count'>): void {
  push({
    kind: 'patch',
    postId: post.post_id,
    liked: post.liked,
    likeCount: post.like_count,
    commentCount: post.comment_count,
  })
}

export function removePost(postId: number): void {
  if (postId <= 0) {
    return
  }
  push({ kind: 'remove', postId })
}

export function touchAlbum(userId: number): void {
  if (userId <= 0) {
    return
  }
  push({ kind: 'album', userId })
}

export function touchCoin(): void {
  push({ kind: 'coin' })
}

/** 就地改列表。返回 true 表示这一页的帖子要整表重拉。 */
export function applyPostEvents(
  list: PostItem[],
  batch: FeedEvent[],
  scope: { mode: 'feed' } | { mode: 'profile'; userId: number },
): boolean {
  let reload = false
  for (const event of batch) {
    if (event.kind === 'reload-posts' && scope.mode === 'feed') {
      reload = true
      continue
    }
    if (
      event.kind === 'reload-profile' &&
      scope.mode === 'profile' &&
      event.userId === scope.userId
    ) {
      reload = true
      continue
    }
    if (event.kind === 'remove') {
      const index = list.findIndex((p) => p.post_id === event.postId)
      if (index >= 0) {
        list.splice(index, 1)
      }
      continue
    }
    if (event.kind === 'patch') {
      const item = list.find((p) => p.post_id === event.postId)
      if (item) {
        item.liked = event.liked
        item.like_count = event.likeCount
        item.comment_count = event.commentCount
      }
    }
  }
  return reload
}
