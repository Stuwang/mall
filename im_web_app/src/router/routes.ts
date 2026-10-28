import type { RouteRecordRaw } from 'vue-router'

export const routeNames = {
  online: 'online',
  posts: 'posts',
  postCreate: 'postCreate',
  postDetail: 'postDetail',
  convs: 'convs',
  mine: 'mine',
  settings: 'settings',
  about: 'about',
  login: 'login',
  register: 'register',
  chat: 'chat',
  profile: 'profile',
  album: 'album',
} as const

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/TabLayout.vue'),
    children: [
      {
        path: '',
        name: routeNames.online,
        component: () => import('@/views/online/OnlineView.vue'),
        meta: { title: '在线', tab: true, requireAuth: true },
      },
      {
        path: 'posts',
        name: routeNames.posts,
        component: () => import('@/views/post/PostListView.vue'),
        meta: { title: '帖子', tab: true, requireAuth: true },
      },
      {
        path: 'convs',
        name: routeNames.convs,
        component: () => import('@/views/convs/ConvListView.vue'),
        meta: { title: '消息', tab: true, requireAuth: true },
      },
      {
        path: 'mine',
        name: routeNames.mine,
        component: () => import('@/views/mine/MineView.vue'),
        meta: { title: '我的', tab: true, requireAuth: true },
      },
    ],
  },
  {
    path: '/login',
    name: routeNames.login,
    component: () => import('@/views/auth/LoginView.vue'),
    meta: { public: true, title: '登录' },
  },
  {
    path: '/register',
    name: routeNames.register,
    component: () => import('@/views/auth/RegisterView.vue'),
    meta: { public: true, title: '注册' },
  },
  {
    path: '/chat/:peerId',
    name: routeNames.chat,
    component: () => import('@/views/chat/ChatView.vue'),
    meta: { title: '聊天', requireAuth: true },
  },
  {
    path: '/u/:userId',
    name: routeNames.profile,
    component: () => import('@/views/profile/ProfileView.vue'),
    meta: { title: '主页', requireAuth: true },
  },
  {
    path: '/u/:userId/album',
    name: routeNames.album,
    component: () => import('@/views/profile/AlbumView.vue'),
    meta: { title: '相册', requireAuth: true },
  },
  {
    path: '/settings',
    name: routeNames.settings,
    component: () => import('@/views/mine/SettingsView.vue'),
    meta: { title: '设置', requireAuth: true },
  },
  {
    path: '/about',
    name: routeNames.about,
    component: () => import('@/views/mine/AboutView.vue'),
    meta: { title: '关于', requireAuth: true },
  },
  {
    path: '/posts/new',
    name: routeNames.postCreate,
    component: () => import('@/views/post/PostCreateView.vue'),
    meta: { title: '发布', requireAuth: true },
  },
  {
    path: '/posts/:postId',
    name: routeNames.postDetail,
    component: () => import('@/views/post/PostDetailView.vue'),
    meta: { title: '帖子', requireAuth: true },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]
