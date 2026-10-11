import type { RouteRecordRaw } from 'vue-router'

export const routeNames = {
  login: 'login',
  dashboard: 'dashboard',
  users: 'users',
  userDetail: 'userDetail',
  categories: 'categories',
  products: 'products',
  productDetail: 'productDetail',
  orders: 'orders',
  banners: 'banners',
  announcements: 'announcements',
} as const

export const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: routeNames.login,
    component: () => import('@/views/login/LoginView.vue'),
    meta: { public: true, title: '登录' },
  },
  {
    path: '/',
    component: () => import('@/layouts/AdminLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: routeNames.dashboard,
        component: () => import('@/views/dashboard/DashboardView.vue'),
        meta: { title: '概览', menu: true },
      },
      {
        path: 'users',
        name: routeNames.users,
        component: () => import('@/views/users/UserListView.vue'),
        meta: { title: '用户管理', menu: true },
      },
      {
        path: 'users/:id',
        name: routeNames.userDetail,
        component: () => import('@/views/users/UserDetailView.vue'),
        meta: { title: '用户详情', menuParent: '/users' },
      },
      {
        path: 'categories',
        name: routeNames.categories,
        component: () => import('@/views/categories/CategoryListView.vue'),
        meta: { title: '分类管理', menu: true },
      },
      {
        path: 'products',
        name: routeNames.products,
        component: () => import('@/views/products/ProductListView.vue'),
        meta: { title: '商品管理', menu: true },
      },
      {
        path: 'products/:id',
        name: routeNames.productDetail,
        component: () => import('@/views/products/ProductDetailView.vue'),
        meta: { title: '商品详情', menuParent: '/products' },
      },
      {
        path: 'orders',
        name: routeNames.orders,
        component: () => import('@/views/orders/OrderListView.vue'),
        meta: { title: '订单管理', menu: true },
      },
      {
        path: 'banners',
        name: routeNames.banners,
        component: () => import('@/views/banners/BannerListView.vue'),
        meta: { title: 'Banner 管理', menu: true },
      },
      {
        path: 'announcements',
        name: routeNames.announcements,
        component: () => import('@/views/announcements/AnnouncementListView.vue'),
        meta: { title: '公告管理', menu: true },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard',
  },
]
