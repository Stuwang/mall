<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import { CategoryStatus, type CategoryInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { formatUnix } from '@/utils/format'
import { categoryStatusLabel } from '@/utils/labels'
import { showError } from '@/utils/notify'

const api = appContext.api.categories
const loading = ref(false)
const list = ref<CategoryInfo[]>([])
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)

const form = reactive({
  name: '',
  sort: 0,
  status: CategoryStatus.ON as CategoryStatus,
})

async function loadList(): Promise<void> {
  loading.value = true
  try {
    const reply = await api.list({})
    list.value = reply.list ?? []
  } catch (error) {
    showError(error, '加载分类失败')
  } finally {
    loading.value = false
  }
}

function openCreate(): void {
  editingId.value = null
  form.name = ''
  form.sort = 0
  form.status = CategoryStatus.ON
  dialogVisible.value = true
}

function openEdit(row: CategoryInfo): void {
  editingId.value = row.id
  form.name = row.name
  form.sort = row.sort
  form.status = row.status
  dialogVisible.value = true
}

async function handleSave(): Promise<void> {
  if (!form.name.trim()) {
    ElMessage.warning('请填写分类名称')
    return
  }
  saving.value = true
  try {
    if (editingId.value == null) {
      await api.create({
        name: form.name.trim(),
        sort: form.sort,
        status: form.status,
      })
      ElMessage.success('创建成功')
    } else {
      await api.update({
        id: editingId.value,
        name: form.name.trim(),
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
      <el-button type="primary" @click="openCreate">新建分类</el-button>
      <el-button @click="loadList">刷新</el-button>
    </div>

    <el-table v-loading="loading" :data="list" stripe border>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="名称" min-width="160" />
      <el-table-column prop="sort" label="排序" width="100" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === CategoryStatus.ON ? 'success' : 'info'" size="small">
            {{ categoryStatusLabel(row.status) }}
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
      :title="editingId == null ? '新建分类' : '编辑分类'"
      width="420px"
    >
      <el-form label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="form.name" maxlength="64" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option :value="CategoryStatus.ON" label="启用" />
            <el-option :value="CategoryStatus.OFF" label="停用" />
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
