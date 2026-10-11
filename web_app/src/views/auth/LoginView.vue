<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useCartStore } from '@/stores/cartStore'
import { notifyError, notifySuccess } from '@/utils/notify'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const cartStore = useCartStore()

const loading = ref(false)
const form = reactive({
  username: '',
  password: '',
})

async function onSubmit(): Promise<void> {
  if (!form.username.trim() || !form.password) {
    notifyError(new Error('请输入账号和密码'))
    return
  }
  loading.value = true
  try {
    await authStore.login(form.username.trim(), form.password)
    await cartStore.refresh()
    notifySuccess('登录成功')
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    await router.replace(redirect)
  } catch (error) {
    notifyError(error, '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page auth-page">
    <van-nav-bar title="登录" left-arrow @click-left="router.back()" />
    <div class="auth-hero">
      <h1>Mall</h1>
      <p class="muted">登录后可使用购物车与订单</p>
    </div>
    <van-form @submit="onSubmit">
      <van-cell-group inset>
        <van-field
          v-model="form.username"
          name="username"
          label="账号"
          placeholder="用户名"
          clearable
          :rules="[{ required: true, message: '请填写账号' }]"
        />
        <van-field
          v-model="form.password"
          type="password"
          name="password"
          label="密码"
          placeholder="密码"
          :rules="[{ required: true, message: '请填写密码' }]"
        />
      </van-cell-group>
      <div class="auth-actions">
        <van-button round block type="primary" native-type="submit" :loading="loading">
          登录
        </van-button>
        <van-button round block plain type="primary" @click="router.push({ name: routeNames.register })">
          注册账号
        </van-button>
      </div>
    </van-form>
  </div>
</template>

<style scoped>
.auth-hero {
  padding: 32px 24px 12px;
}

.auth-hero h1 {
  margin: 0 0 8px;
  font-size: 32px;
  letter-spacing: 0.04em;
}

.auth-actions {
  display: grid;
  gap: 12px;
  margin: 24px 16px;
}
</style>
