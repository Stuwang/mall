<script setup lang="ts">
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, type UploadRequestOptions, type UploadUserFile } from 'element-plus'
import { computed, ref, watch } from 'vue'
import { appContext } from '@/app/AppContext'
import { showError } from '@/utils/notify'

const props = withDefaults(
  defineProps<{
    /** 图片地址列表或单图；上传后为可访问 url，写库时服务端 Normalize */
    modelValue?: string | string[]
    limit?: number
  }>(),
  {
    limit: 9,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string | string[]]
}>()

const mediaApi = appContext.api.media
const fileList = ref<UploadUserFile[]>([])
const syncing = ref(false)

const isSingle = computed(() => props.limit <= 1)

function asList(value: string | string[] | undefined): string[] {
  if (value == null) return []
  if (Array.isArray(value)) return value.map((s) => String(s).trim()).filter(Boolean)
  const s = String(value).trim()
  return s ? [s] : []
}

function emitValue(urls: string[]): void {
  emit('update:modelValue', isSingle.value ? (urls[0] ?? '') : urls)
}

watch(
  () => props.modelValue,
  (val) => {
    if (syncing.value) return
    const urls = asList(val)
    fileList.value = urls.map((url, i) => ({
      name: displayNameFromURL(url, i),
      url,
      status: 'success',
      uid: Date.now() + i,
    }))
  },
  { immediate: true },
)

function displayNameFromURL(url: string, index: number): string {
  try {
    const path = url.split('?')[0] ?? url
    const base = path.split('/').pop()
    if (base) return base
  } catch {
    /* ignore */
  }
  return `image-${index + 1}`
}

async function handleUpload(options: UploadRequestOptions): Promise<void> {
  try {
    const reply = await mediaApi.upload(options.file as File)
    const next = [...asList(props.modelValue), reply.url].slice(0, props.limit)
    syncing.value = true
    emitValue(next)
    // 列表展示用服务端随机文件名，不用用户本地原名
    fileList.value = next.map((url, i) => ({
      name: displayNameFromURL(url, i),
      url,
      status: 'success',
      uid: Date.now() + i,
    }))
    syncing.value = false
    options.onSuccess?.(reply)
  } catch (error) {
    showError(error, '上传失败')
    options.onError?.(error as never)
  }
}

function handleRemove(_file: UploadUserFile, files: UploadUserFile[]): void {
  const next = files.map((f) => f.url).filter((u): u is string => Boolean(u))
  syncing.value = true
  emitValue(next)
  syncing.value = false
}

function handleExceed(): void {
  ElMessage.warning(`最多上传 ${props.limit} 张图片`)
}
</script>

<template>
  <el-upload
    v-model:file-list="fileList"
    list-type="picture-card"
    :limit="limit"
    accept="image/jpeg,image/png,image/gif,image/webp,.jpg,.jpeg,.png,.gif,.webp"
    :http-request="handleUpload"
    :on-remove="handleRemove"
    :on-exceed="handleExceed"
  >
    <el-icon><Plus /></el-icon>
  </el-upload>
</template>
