import type { RouteRecordRaw } from 'vue-router'

export const routeNames = {
  home: 'home',
  cart: 'cart',
  mine: 'mine',
  login: 'login',
  register: 'register',
  productDetail: 'productDetail',
  orders: 'orders',
  orderDetail: 'orderDetail',
  payCheckout: 'payCheckout',
  mockPay: 'mockPay',
  pointsPurchase: 'pointsPurchase',
  help: 'help',
  about: 'about',
  contact: 'contact',
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
    path: '/orders/:id',
    name: routeNames.orderDetail,
    component: () => import('@/views/order/OrderDetailView.vue'),
    meta: { title: '订单详情', requireAuth: true },
  },
  {
    path: '/pay/:id',
    name: routeNames.payCheckout,
    component: () => import('@/views/order/PayCheckoutView.vue'),
    meta: { title: '确认支付', requireAuth: true },
  },
  {
    path: '/mock-pay/:id',
    name: routeNames.mockPay,
    component: () => import('@/views/order/MockPayView.vue'),
    meta: { title: 'Mock 支付', requireAuth: true },
  },
  {
    path: '/points',
    name: routeNames.pointsPurchase,
    component: () => import('@/views/points/PointsPurchaseView.vue'),
    meta: { title: '购买积分', requireAuth: true },
  },
  {
    path: '/help',
    name: routeNames.help,
    component: () => import('@/views/mine/ContentView.vue'),
    meta: { public: true, title: '帮助中心', contentKind: 'help' },
  },
  {
    path: '/about',
    name: routeNames.about,
    component: () => import('@/views/mine/ContentView.vue'),
    meta: { public: true, title: '关于我们', contentKind: 'about' },
  },
  {
    path: '/contact',
    name: routeNames.contact,
    component: () => import('@/views/mine/ContentView.vue'),
    meta: { public: true, title: '联系我们', contentKind: 'contact' },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]
