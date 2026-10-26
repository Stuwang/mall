<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { notifyError, notifySuccess } from '@/utils/notify'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const form = reactive({
  username: '',
  password: '',
  nickname: '',
})

async function onSubmit(): Promise<void> {
  if (!form.username.trim() || !form.password) {
    notifyError(new Error('请输入账号和密码'))
    return
  }
  loading.value = true
  try {
    await authStore.register(
      form.username.trim(),
      form.password,
      form.nickname.trim() || undefined,
    )
    notifySuccess('注册成功')
    await router.replace({ name: routeNames.convs })
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
      </van-cell-group>
      <div class="auth-actions">
        <van-button round block type="primary" native-type="submit" :loading="loading">
          注册并登录
        </van-button>
      </div>
    </van-form>
  </div>
</template>

<style scoped>
.auth-actions {
  margin: 24px 16px;
}
</style>
