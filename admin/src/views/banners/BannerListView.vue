<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import { MiscStatus, type BannerInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import ImageUpload from '@/components/ImageUpload.vue'
import { formatUnix } from '@/utils/format'
import { miscStatusLabel } from '@/utils/labels'
import { showError } from '@/utils/notify'

const api = appContext.api.banners
const loading = ref(false)
const list = ref<BannerInfo[]>([])
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)

const form = reactive({
  image_url: '',
  sort: 0,
  status: MiscStatus.ON as MiscStatus,
})

async function loadList(): Promise<void> {
  loading.value = true
  try {
    const reply = await api.list({})
    list.value = reply.list ?? []
  } catch (error) {
    showError(error, '加载 Banner 失败')
  } finally {
    loading.value = false
  }
}

function openCreate(): void {
  editingId.value = null
  form.image_url = ''
  form.sort = 0
  form.status = MiscStatus.ON
  dialogVisible.value = true
}

function openEdit(row: BannerInfo): void {
  editingId.value = row.id
  form.image_url = row.image_url
  form.sort = row.sort
  form.status = row.status
  dialogVisible.value = true
}

async function handleSave(): Promise<void> {
  if (!form.image_url.trim()) {
    ElMessage.warning('请上传 Banner 图片')
    return
  }
  saving.value = true
  try {
    if (editingId.value == null) {
      await api.create({
        image_url: form.image_url.trim(),
        sort: form.sort,
        status: form.status,
      })
      ElMessage.success('创建成功')
    } else {
      await api.update({
        id: editingId.value,
        image_url: form.image_url.trim(),
        sort: form.sort,
        status: form.status,
      })
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    await loadList()
  } catch (error) {
    showError(error, '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadList()
})
</script>

<template>
  <div class="page-card">
    <div class="page-toolbar">
      <el-button type="primary" @click="openCreate">新建 Banner</el-button>
      <el-button @click="loadList">刷新</el-button>
    </div>

    <el-table v-loading="loading" :data="list" stripe border>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column label="预览" width="140">
        <template #default="{ row }">
          <el-image
            v-if="row.image_url"
            :src="row.image_url"
            fit="cover"
            style="width: 100px; height: 56px"
            :preview-src-list="[row.image_url]"
          />
        </template>
      </el-table-column>
      <el-table-column prop="image_url" label="图片 URL" min-width="220" show-overflow-tooltip />
      <el-table-column prop="sort" label="排序" width="100" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === MiscStatus.ON ? 'success' : 'info'" size="small">
            {{ miscStatusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" min-width="160">
        <template #default="{ row }">{{ formatUnix(row.ut) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog
      v-model="dialogVisible"
      :title="editingId == null ? '新建 Banner' : '编辑 Banner'"
      width="480px"
    >
      <el-form label-width="90px">
        <el-form-item label="图片">
          <ImageUpload v-model="form.image_url" :limit="1" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option :value="MiscStatus.ON" label="启用" />
            <el-option :value="MiscStatus.OFF" label="停用" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
