<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import { AssetType, type UserInfo } from '@/api/types'
import { appContext } from '@/app/AppContext'
import { formatUnix } from '@/utils/format'
import { showError } from '@/utils/notify'

const userApi = appContext.api.users
const assetApi = appContext.api.assets
const loading = ref(false)
const list = ref<UserInfo[]>([])
const total = ref(0)

const query = reactive({
  user_id: undefined as number | undefined,
  keyword: '',
  page: 1,
  page_size: 20,
})

const pointsDialogVisible = ref(false)
const pointsLoading = ref(false)
const pointsSubmitting = ref(false)
const pointsUser = ref<UserInfo | null>(null)
const pointBalance = ref(0)
const grantForm = reactive({
  points: 100,
  remark: '',
})

async function loadList(): Promise<void> {
  loading.value = true
  try {
    const reply = await userApi.list({
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
    await userApi.ban({ user_id: row.id, ban_until: banUntil })
    ElMessage.success(banUntil === 0 ? '已解封' : '已封禁')
    await loadList()
  } catch (error) {
    if (error === 'cancel' || error === 'close') {
      return
    }
    showError(error, '操作失败')
  }
}

async function openPointsDialog(row: UserInfo): Promise<void> {
  pointsUser.value = row
  pointsDialogVisible.value = true
  pointsLoading.value = true
  grantForm.points = 100
  grantForm.remark = ''
  try {
    const reply = await assetApi.getUserAssets({
      user_id: row.id,
      type: AssetType.POINT,
    })
    const point = (reply.assets ?? []).find((item) => item.type === AssetType.POINT)
    pointBalance.value = point?.balance ?? 0
  } catch (error) {
    showError(error, '加载积分失败')
    pointsDialogVisible.value = false
  } finally {
    pointsLoading.value = false
  }
}

async function submitGrantPoints(): Promise<void> {
  if (!pointsUser.value) return
  if (!grantForm.points || grantForm.points <= 0) {
    ElMessage.error('积分数必须大于 0')
    return
  }
  pointsSubmitting.value = true
  try {
    const reply = await assetApi.grantPoints({
      user_id: pointsUser.value.id,
      points: grantForm.points,
      remark: grantForm.remark.trim() || undefined,
    })
    pointBalance.value = reply.asset?.balance ?? pointBalance.value + grantForm.points
    ElMessage.success('加积分成功')
    grantForm.points = 100
    grantForm.remark = ''
  } catch (error) {
    showError(error, '加积分失败')
  } finally {
    pointsSubmitting.value = false
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
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openPointsDialog(row)">积分</el-button>
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

    <el-dialog
      v-model="pointsDialogVisible"
      :title="pointsUser ? `用户积分 · ${pointsUser.nickname || pointsUser.username}` : '用户积分'"
      width="480px"
      destroy-on-close
    >
      <div v-loading="pointsLoading" class="points-body">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="用户 ID">{{ pointsUser?.id }}</el-descriptions-item>
          <el-descriptions-item label="当前积分">{{ pointBalance }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">手动加积分</el-divider>

        <el-form label-width="80px">
          <el-form-item label="积分数" required>
            <el-input-number v-model="grantForm.points" :min="1" :step="100" />
          </el-form-item>
          <el-form-item label="备注">
            <el-input
              v-model="grantForm.remark"
              maxlength="64"
              show-word-limit
              placeholder="可选，如活动赠送"
            />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="pointsDialogVisible = false">关闭</el-button>
        <el-button type="primary" :loading="pointsSubmitting" @click="submitGrantPoints">
          确认加积分
        </el-button>
      </template>
    </el-dialog>
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

.points-body {
  min-height: 120px;
}
</style>
