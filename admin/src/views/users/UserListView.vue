<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import type { UserInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { formatUnix } from '@/utils/format'
import { showError } from '@/utils/notify'

const api = appContext.api.users
const loading = ref(false)
const list = ref<UserInfo[]>([])
const total = ref(0)

const query = reactive({
  user_id: undefined as number | undefined,
  keyword: '',
  page: 1,
  page_size: 20,
})

async function loadList(): Promise<void> {
  loading.value = true
  try {
    const reply = await api.list({
      user_id: query.user_id || 0,
      keyword: query.keyword.trim(),
      page: query.page,
      page_size: query.page_size,
    })
    list.value = reply.list ?? []
    total.value = Number(reply.total ?? 0)
  } catch (error) {
    showError(error, '加载用户失败')
  } finally {
    loading.value = false
  }
}

function handleSearch(): void {
  query.page = 1
  void loadList()
}

async function handleBan(row: UserInfo): Promise<void> {
  try {
    const { value } = await ElMessageBox.prompt(
      '请输入封禁截止时间（本地时间，格式 YYYY-MM-DD HH:mm:ss）。留空并确认将解封。',
      row.banned ? '调整封禁 / 解封' : '封禁用户',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputPlaceholder: '例如 2026-12-31 23:59:59，空=解封',
        inputValue: '',
      },
    )
    const trimmed = value.trim()
    let banUntil = 0
    if (trimmed) {
      const ts = Math.floor(new Date(trimmed.replace(/-/g, '/')).getTime() / 1000)
      if (!Number.isFinite(ts) || ts <= 0) {
        ElMessage.error('时间格式无效')
        return
      }
      banUntil = ts
    }
    await api.ban({ user_id: row.id, ban_until: banUntil })
    ElMessage.success(banUntil === 0 ? '已解封' : '已封禁')
    await loadList()
  } catch (error) {
    if (error === 'cancel' || error === 'close') {
      return
    }
    showError(error, '操作失败')
  }
}

onMounted(() => {
  void loadList()
})
</script>

<template>
  <div class="page-card">
    <div class="page-toolbar">
      <el-input
        v-model="query.keyword"
        clearable
        placeholder="昵称 / 账号"
        style="width: 200px"
        @keyup.enter="handleSearch"
      />
      <el-input-number
        v-model="query.user_id"
        :min="0"
        :controls="false"
        placeholder="用户 ID"
      />
      <el-button type="primary" @click="handleSearch">查询</el-button>
      <div class="grow" />
      <span class="total">共 {{ total }} 人</span>
    </div>

    <el-table v-loading="loading" :data="list" stripe border>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="账号" min-width="120" />
      <el-table-column prop="nickname" label="昵称" min-width="120" />
      <el-table-column label="管理员" width="90">
        <template #default="{ row }">
          <el-tag :type="row.is_admin ? 'danger' : 'info'" size="small">
            {{ row.is_admin ? '是' : '否' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="封禁" width="90">
        <template #default="{ row }">
          <el-tag :type="row.banned ? 'warning' : 'success'" size="small">
            {{ row.banned ? '封禁中' : '正常' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="封禁至" min-width="160">
        <template #default="{ row }">{{ formatUnix(row.ban_until) }}</template>
      </el-table-column>
      <el-table-column label="创建时间" min-width="160">
        <template #default="{ row }">{{ formatUnix(row.ct) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="handleBan(row)">
            {{ row.banned ? '解封/改期' : '封禁' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        background
        layout="total, prev, pager, next, sizes"
        :total="total"
        :page-sizes="[10, 20, 50]"
        @current-change="loadList"
        @size-change="
          () => {
            query.page = 1
            loadList()
          }
        "
      />
    </div>
  </div>
</template>

<style scoped>
.total {
  color: #6b7280;
}

.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
