<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import UserAvatar from '@/components/UserAvatar.vue'
import { routeNames } from '@/router/routes'
import { useAuthStore } from '@/stores/authStore'
import { useImLive, useImStore } from '@/stores/imStore'
import { notifyError, notifySuccess } from '@/utils/notify'

const router = useRouter()
const authStore = useAuthStore()
const imStore = useImStore()
useImLive()

const showNickname = ref(false)
const nicknameDraft = ref('')
const savingNickname = ref(false)
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

function openProfile(): void {
  const id = authStore.user?.id
  if (!id) {
    notifyError(null, '未登录')
    return
  }
  void router.push({ name: routeNames.profile, params: { userId: String(id) } })
}

function logout(): void {
  authStore.logout()
  notifySuccess('已退出')
  void router.replace({ name: routeNames.login })
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="我的" />
    <div class="profile">
      <button type="button" class="avatar-btn" :disabled="uploading" @click="pickAvatar">
        <UserAvatar
          :user-id="authStore.user?.id ?? 0"
          :nickname="authStore.displayName"
          :src="authStore.user?.avatar"
          :size="72"
        />
        <span class="avatar-tip">{{ uploading ? '上传中…' : '点击更换头像' }}</span>
      </button>
      <input
        ref="fileInput"
        class="hidden-input"
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp"
        @change="onAvatarSelected"
      />
    </div>
    <van-cell-group inset>
      <van-cell title="主页" is-link @click="openProfile" />
      <van-cell title="昵称" is-link :value="authStore.displayName" @click="openNickname" />
      <van-cell title="用户 ID" :value="String(authStore.user?.id ?? '')" />
      <van-cell title="账号" :value="authStore.user?.username || '-'" />
      <van-cell title="长连接" :value="imStore.wsConnected ? '已连接' : '未连接'" />
    </van-cell-group>
    <div class="actions">
      <van-button round block type="danger" plain @click="logout">退出登录</van-button>
    </div>

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
  </div>
</template>

<style scoped>
.profile {
  display: flex;
  justify-content: center;
  padding: 24px 16px 8px;
}

.avatar-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  border: none;
  background: transparent;
  cursor: pointer;
  padding: 0;
}

.avatar-btn:disabled {
  opacity: 0.7;
}

.avatar-tip {
  font-size: 13px;
  color: #969799;
}

.hidden-input {
  display: none;
}

.actions {
  margin: 24px 16px;
}

.dialog-body {
  padding: 8px 16px 16px;
}
</style>
