<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useCartStore } from '@/stores/cartStore'
import { notifyError, notifySuccess } from '@/utils/notify'

const router = useRouter()
const authStore = useAuthStore()
const cartStore = useCartStore()

const loading = ref(false)
const form = reactive({
  username: '',
  nickname: '',
  password: '',
  confirm: '',
})

async function onSubmit(): Promise<void> {
  if (form.password !== form.confirm) {
    notifyError(new Error('两次密码不一致'))
    return
  }
  loading.value = true
  try {
    await authStore.register(
      form.username.trim(),
      form.password,
      form.nickname.trim() || undefined,
    )
    await cartStore.refresh()
    notifySuccess('注册成功')
    await router.replace({ name: routeNames.home })
  } catch (error) {
    notifyError(error, '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page auth-page">
    <van-nav-bar title="注册" left-arrow @click-left="router.back()" />
    <div class="auth-hero">
      <h1>创建账号</h1>
      <p class="muted">注册成功后将自动登录</p>
    </div>
    <van-form @submit="onSubmit">
      <van-cell-group inset>
        <van-field
          v-model="form.username"
          name="username"
          label="账号"
          placeholder="登录用户名"
          clearable
          :rules="[{ required: true, message: '请填写账号' }]"
        />
        <van-field
          v-model="form.nickname"
          name="nickname"
          label="昵称"
          placeholder="可选"
          clearable
        />
        <van-field
          v-model="form.password"
          type="password"
          name="password"
          label="密码"
          placeholder="密码"
          :rules="[{ required: true, message: '请填写密码' }]"
        />
        <van-field
          v-model="form.confirm"
          type="password"
          name="confirm"
          label="确认"
          placeholder="再次输入密码"
          :rules="[{ required: true, message: '请确认密码' }]"
        />
      </van-cell-group>
      <div class="auth-actions">
        <van-button round block type="primary" native-type="submit" :loading="loading">
          注册并登录
        </van-button>
        <van-button round block plain type="primary" @click="router.push({ name: routeNames.login })">
          已有账号去登录
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
  font-size: 28px;
}

.auth-actions {
  display: grid;
  gap: 12px;
  margin: 24px 16px;
}
</style>
