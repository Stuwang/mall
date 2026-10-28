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
const uploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
/** 本地预览 + 上传后的 uri */
const images = ref<{ preview: string; uri: string }[]>([])

const maxImages = 9

function pickImages(): void {
  if (images.value.length >= maxImages) {
    notifyError(null, `最多 ${maxImages} 张图`)
    return
  }
  fileInput.value?.click()
}

async function onFilesSelected(ev: Event): Promise<void> {
  const input = ev.target as HTMLInputElement
  const files = Array.from(input.files || [])
  input.value = ''
  if (files.length === 0) {
    return
  }
  const room = maxImages - images.value.length
  const picked = files.slice(0, room).filter((f) => f.type.startsWith('image/'))
  if (picked.length === 0) {
    notifyError(null, '请选择图片文件')
    return
  }
  uploading.value = true
  try {
    for (const file of picked) {
      const uploaded = await api.auth.uploadMedia(file)
      if (!uploaded?.uri) {
        notifyError(null, '上传未返回文件地址')
        continue
      }
      images.value.push({
        preview: uploaded.url || URL.createObjectURL(file),
        uri: uploaded.uri,
      })
    }
  } catch (error) {
    notifyError(error, '上传失败')
  } finally {
    uploading.value = false
  }
}

function removeImage(index: number): void {
  images.value.splice(index, 1)
}

async function onSubmit(): Promise<void> {
  const body = text.value.trim()
  if (!body && images.value.length === 0) {
    notifyError(null, '请输入内容或添加图片')
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
      image_uris: images.value.map((x) => x.uri),
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
        <span
          class="nav-action"
          :class="{ disabled: submitting || uploading }"
          @click="onSubmit"
        >
          发布
        </span>
      </template>
    </van-nav-bar>
    <div class="editor">
      <van-field
        v-model="text"
        type="textarea"
        rows="6"
        maxlength="1000"
        show-word-limit
        placeholder="说点什么…"
        autosize
      />
      <div class="pics">
        <div v-for="(img, i) in images" :key="i" class="pic">
          <img :src="img.preview" alt="" />
          <button type="button" class="pic-del" aria-label="删除" @click="removeImage(i)">
            <van-icon name="cross" />
          </button>
        </div>
        <button
          v-if="images.length < maxImages"
          type="button"
          class="pic add"
          :disabled="uploading"
          @click="pickImages"
        >
          <van-icon name="plus" />
          <span>{{ uploading ? '上传中' : '图片' }}</span>
        </button>
      </div>
      <input
        ref="fileInput"
        class="hidden-input"
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp"
        multiple
        @change="onFilesSelected"
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

.pics {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  padding: 0 12px 16px;
}

.pic {
  position: relative;
  aspect-ratio: 1;
  border-radius: 8px;
  overflow: hidden;
  background: #f2f3f5;
  border: none;
  padding: 0;
}

.pic img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.pic.add {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  color: #969799;
  font-size: 12px;
  cursor: pointer;
}

.pic.add:disabled {
  opacity: 0.6;
}

.pic-del {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 20px;
  height: 20px;
  border: none;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  cursor: pointer;
}

.hidden-input {
  display: none;
}
</style>
