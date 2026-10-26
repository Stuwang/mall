import type { RouteRecordRaw } from 'vue-router'

export const routeNames = {
  online: 'online',
  convs: 'convs',
  mine: 'mine',
  login: 'login',
  register: 'register',
  chat: 'chat',
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
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]
