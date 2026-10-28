<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { notifyError, notifySuccess } from '@/utils/notify'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

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
    <div class="auth-hero im-fade-up">
      <div class="brand">IM</div>
      <h1>欢迎回来</h1>
      <p>一对一即时消息 · 帖子与相册</p>
    </div>
    <van-form class="auth-form im-fade-up" @submit="onSubmit">
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
        <van-button
          round
          block
          plain
          type="primary"
          @click="router.push({ name: routeNames.register })"
        >
          注册账号
        </van-button>
      </div>
    </van-form>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 24px 0 48px;
}

.auth-hero {
  padding: 12px 28px 28px;
}

.brand {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 56px;
  height: 56px;
  padding: 0 14px;
  margin-bottom: 18px;
  border-radius: 18px;
  background: linear-gradient(145deg, var(--im-primary), var(--im-primary-deep));
  color: #fff;
  font-family: Outfit, 'Noto Sans SC', sans-serif;
  font-size: 22px;
  font-weight: 700;
  letter-spacing: 0.06em;
  box-shadow: 0 14px 28px rgba(15, 143, 134, 0.28);
}

.auth-hero h1 {
  margin: 0 0 8px;
  font-size: 28px;
  font-weight: 700;
  letter-spacing: 0.01em;
}

.auth-hero p {
  margin: 0;
  color: var(--im-text-secondary);
  font-size: 14px;
}

.auth-form {
  animation-delay: 0.08s;
}

.auth-actions {
  display: grid;
  gap: 12px;
  margin: 24px 16px;
}
</style>
