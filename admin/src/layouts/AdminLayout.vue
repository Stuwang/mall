<script setup lang="ts">
import {
  Box,
  Goods,
  List,
  Menu as MenuIcon,
  ShoppingCart,
  User,
} from '@element-plus/icons-vue'
import { computed } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const activeMenu = computed(() => route.path)

const menus = [
  { path: '/dashboard', title: '概览', icon: MenuIcon },
  { path: '/users', title: '用户管理', icon: User },
  { path: '/categories', title: '分类管理', icon: List },
  { path: '/products', title: '商品管理', icon: Goods },
  { path: '/orders', title: '订单管理', icon: ShoppingCart },
]

function handleLogout(): void {
  auth.logout()
  void router.replace({ name: routeNames.login })
}
</script>

<template>
  <el-container class="admin-layout">
    <el-aside width="220px" class="aside">
      <div class="brand">
        <el-icon :size="20"><Box /></el-icon>
        <span>Mall Admin</span>
      </div>
      <el-menu :default-active="activeMenu" router>
        <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.title }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-title">{{ route.meta.title }}</div>
        <div class="header-right">
          <span class="username">{{ auth.displayName }}</span>
          <el-button type="danger" link @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>
      <el-main class="main">
        <RouterView />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.admin-layout {
  min-height: 100vh;
}

.aside {
  background: #111827;
  color: #fff;
}

.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 56px;
  padding: 0 20px;
  font-weight: 700;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.aside :deep(.el-menu) {
  border-right: none;
  background: transparent;
}

.aside :deep(.el-menu-item) {
  color: #d1d5db;
}

.aside :deep(.el-menu-item.is-active) {
  background: #1f2937;
  color: #fff;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.username {
  color: #4b5563;
}

.main {
  background: #f3f4f6;
}
</style>
