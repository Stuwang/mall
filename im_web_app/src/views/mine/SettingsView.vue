<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { appContext } from '@/app/AppContext'
import { notifyError, notifySuccess } from '@/utils/notify'

const router = useRouter()
const authStore = useAuthStore()
const api = appContext.api

const showNickname = ref(false)
const nicknameDraft = ref('')
const savingNickname = ref(false)

const showPassword = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const savingPassword = ref(false)

const uploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

function openNickname(): void {
  nicknameDraft.value = authStore.user?.nickname || authStore.user?.username || ''
  showNickname.value = true
}

async function onNicknameClose(action: string): Promise<boolean> {
  if (action !== 'confirm') {
    return true
  }
  const nickname = nicknameDraft.value.trim()
  if (!nickname) {
    notifyError(null, '昵称不能为空')
    return false
  }
  savingNickname.value = true
  try {
    await authStore.updateProfile({
      nickname,
      sex: authStore.user?.sex ?? 0,
    })
    notifySuccess('昵称已更新')
    return true
  } catch (error) {
    notifyError(error, '修改昵称失败')
    return false
  } finally {
    savingNickname.value = false
  }
}

function openPassword(): void {
  oldPassword.value = ''
  newPassword.value = ''
  confirmPassword.value = ''
  showPassword.value = true
}

async function onPasswordClose(action: string): Promise<boolean> {
  if (action !== 'confirm') {
    return true
  }
  const oldPwd = oldPassword.value
  const newPwd = newPassword.value
  const confirm = confirmPassword.value
  if (!oldPwd || !newPwd) {
    notifyError(null, '请填写原密码和新密码')
    return false
  }
  if (newPwd !== confirm) {
    notifyError(null, '两次新密码不一致')
    return false
  }
  if (oldPwd === newPwd) {
    notifyError(null, '新密码不能与原密码相同')
    return false
  }
  savingPassword.value = true
  try {
    await api.auth.changePassword({
      old_password: oldPwd,
      new_password: newPwd,
    })
    notifySuccess('密码已修改')
    return true
  } catch (error) {
    notifyError(error, '修改密码失败')
    return false
  } finally {
    savingPassword.value = false
  }
}

function pickAvatar(): void {
  fileInput.value?.click()
}

async function onAvatarSelected(ev: Event): Promise<void> {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) {
    return
  }
  if (!file.type.startsWith('image/')) {
    notifyError(null, '请选择图片文件')
    return
  }
  uploading.value = true
  try {
    await authStore.uploadAvatar(file)
    notifySuccess('头像已更新')
  } catch (error) {
    notifyError(error, '上传头像失败')
  } finally {
    uploading.value = false
  }
}

function logout(): void {
  authStore.logout()
  notifySuccess('已退出')
  void router.replace({ name: routeNames.login })
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="设置" left-arrow fixed placeholder @click-left="router.back()" />

    <van-cell-group inset title="账号">
      <van-cell title="修改昵称" is-link :value="authStore.displayName" @click="openNickname" />
      <van-cell title="修改头像" is-link :value="uploading ? '上传中…' : ''" @click="pickAvatar" />
      <van-cell title="修改密码" is-link @click="openPassword" />
    </van-cell-group>

    <van-cell-group inset title="其他">
      <van-cell title="退出登录" is-link @click="logout" />
    </van-cell-group>

    <input
      ref="fileInput"
      class="hidden-input"
      type="file"
      accept="image/jpeg,image/png,image/gif,image/webp"
      @change="onAvatarSelected"
    />

    <van-dialog
      v-model:show="showNickname"
      title="修改昵称"
      show-cancel-button
      :confirm-button-loading="savingNickname"
      :before-close="onNicknameClose"
    >
      <div class="dialog-body">
        <van-field v-model="nicknameDraft" maxlength="32" placeholder="输入新昵称" />
      </div>
    </van-dialog>

    <van-dialog
      v-model:show="showPassword"
      title="修改密码"
      show-cancel-button
      :confirm-button-loading="savingPassword"
      :before-close="onPasswordClose"
    >
      <div class="dialog-body">
        <van-field
          v-model="oldPassword"
          type="password"
          maxlength="64"
          placeholder="原密码"
        />
        <van-field
          v-model="newPassword"
          type="password"
          maxlength="64"
          placeholder="新密码"
        />
        <van-field
          v-model="confirmPassword"
          type="password"
          maxlength="64"
          placeholder="确认新密码"
        />
      </div>
    </van-dialog>
  </div>
</template>

<style scoped>
.page :deep(.van-cell-group) {
  margin-top: 8px;
}

.hidden-input {
  display: none;
}

.dialog-body {
  padding: 8px 16px 16px;
}
</style>
