import type { RouteRecordRaw } from 'vue-router'

export const routeNames = {
  home: 'home',
  cart: 'cart',
  mine: 'mine',
  login: 'login',
  register: 'register',
  productDetail: 'productDetail',
  orders: 'orders',
} as const

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/TabLayout.vue'),
    children: [
      {
        path: '',
        name: routeNames.home,
        component: () => import('@/views/home/HomeView.vue'),
        meta: { public: true, title: '首页', tab: true },
      },
      {
        path: 'cart',
        name: routeNames.cart,
        component: () => import('@/views/cart/CartView.vue'),
        meta: { title: '购物车', tab: true, requireAuth: true },
      },
      {
        path: 'mine',
        name: routeNames.mine,
        component: () => import('@/views/mine/MineView.vue'),
        meta: { public: true, title: '我的', tab: true },
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
    path: '/product/:id',
    name: routeNames.productDetail,
    component: () => import('@/views/product/ProductDetailView.vue'),
    meta: { public: true, title: '商品详情' },
  },
  {
    path: '/orders',
    name: routeNames.orders,
    component: () => import('@/views/order/OrderListView.vue'),
    meta: { title: '我的订单', requireAuth: true },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]
