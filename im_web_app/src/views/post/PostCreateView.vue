<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { appContext } from '@/app/AppContext'
import { routeNames } from '@/router/routes'
import { newClientPostId } from '@/utils/conv'
import { notifyError, notifySuccess } from '@/utils/notify'

const router = useRouter()
const api = appContext.api

const text = ref('')
const submitting = ref(false)

async function onSubmit(): Promise<void> {
  const body = text.value.trim()
  if (!body) {
    notifyError(null, '请输入内容')
    return
  }
  if (body.length > 1000) {
    notifyError(null, '最多 1000 字')
    return
  }
  submitting.value = true
  try {
    await api.post.create({
      client_post_id: newClientPostId(),
      text: body,
    })
    notifySuccess('已发布')
    void router.replace({ name: routeNames.posts })
  } catch (error) {
    notifyError(error, '发布失败')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="page">
    <van-nav-bar title="发布" left-arrow fixed placeholder @click-left="router.back()">
      <template #right>
        <span class="nav-action" :class="{ disabled: submitting }" @click="onSubmit">发布</span>
      </template>
    </van-nav-bar>
    <div class="editor">
      <van-field
        v-model="text"
        type="textarea"
        rows="8"
        maxlength="1000"
        show-word-limit
        placeholder="说点什么…"
        autosize
      />
    </div>
  </div>
</template>

<style scoped>
.nav-action {
  color: var(--im-primary);
  font-size: 15px;
  padding: 0 4px;
  cursor: pointer;
}

.nav-action.disabled {
  opacity: 0.5;
  pointer-events: none;
}

.editor {
  margin: 12px;
  border-radius: 12px;
  overflow: hidden;
  background: #fff;
}
</style>
