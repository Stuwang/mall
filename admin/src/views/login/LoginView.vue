<script setup lang="ts">
import { Lock, User } from '@element-plus/icons-vue'
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { showError } from '@/utils/notify'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

async function handleSubmit(): Promise<void> {
  if (!form.username || !form.password) {
    showError(new Error('请输入用户名和密码'))
    return
  }
  loading.value = true
  try {
    await auth.login(form.username.trim(), form.password)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
    await router.replace(redirect || { name: routeNames.dashboard })
  } catch (error) {
    showError(error, '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-panel">
      <h1>Mall Admin</h1>
      <p class="subtitle">商城后台管理</p>
      <el-form label-position="top" @submit.prevent="handleSubmit">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="管理员账号" :prefix-icon="User" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            placeholder="密码"
            :prefix-icon="Lock"
            @keyup.enter="handleSubmit"
          />
        </el-form-item>
        <el-button type="primary" class="submit" :loading="loading" @click="handleSubmit">
          登录
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  background:
    radial-gradient(circle at 20% 20%, #dbeafe 0, transparent 40%),
    radial-gradient(circle at 80% 0%, #e0e7ff 0, transparent 35%),
    #f8fafc;
}

.login-panel {
  width: min(400px, calc(100vw - 32px));
  background: #fff;
  border-radius: 12px;
  padding: 32px 28px;
  box-shadow: 0 16px 40px rgba(15, 23, 42, 0.08);
}

h1 {
  margin: 0;
  font-size: 28px;
}

.subtitle {
  margin: 8px 0 24px;
  color: #6b7280;
}

.submit {
  width: 100%;
  margin-top: 8px;
}
</style>
