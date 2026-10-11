<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import {
  CategoryStatus,
  ProductStatus,
  type CategoryInfo,
  type ProductInfo,
} from '@/api/types'
import { appContext } from '@/app/AppContext'
import { formatFenToYuan, formatUnix, yuanToFen } from '@/utils/format'
import { productStatusLabel } from '@/utils/labels'
import { showError } from '@/utils/notify'

const productApi = appContext.api.products
const categoryApi = appContext.api.categories

const loading = ref(false)
const saving = ref(false)
const list = ref<ProductInfo[]>([])
const categories = ref<CategoryInfo[]>([])
const total = ref(0)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)

const query = reactive({
  keyword: '',
  category_id: 0 as number,
  status: 0 as ProductStatus,
  page: 1,
  page_size: 20,
})

const form = reactive({
  category_id: 0,
  name: '',
  description: '',
  priceYuan: '0.00',
  stock: 0,
  imagesText: '',
  status: ProductStatus.ON as ProductStatus,
})

function categoryName(id: number): string {
  return categories.value.find((c) => c.id === id)?.name ?? String(id || '-')
}

function parseImages(text: string): string[] {
  return text
    .split(/\n|,/)
    .map((s) => s.trim())
    .filter(Boolean)
}

async function loadCategories(): Promise<void> {
  const reply = await categoryApi.list({})
  categories.value = (reply.list ?? []).filter((c) => c.status !== CategoryStatus.OFF)
}

async function loadList(): Promise<void> {
  loading.value = true
  try {
    const reply = await productApi.list({
      keyword: query.keyword.trim(),
      category_id: query.category_id || 0,
      status: query.status || ProductStatus.UNSPECIFIED,
      page: query.page,
      page_size: query.page_size,
    })
    list.value = reply.list ?? []
    total.value = Number(reply.total ?? 0)
  } catch (error) {
    showError(error, '加载商品失败')
  } finally {
    loading.value = false
  }
}

function handleSearch(): void {
  query.page = 1
  void loadList()
}

function openCreate(): void {
  editingId.value = null
  form.category_id = categories.value[0]?.id ?? 0
  form.name = ''
  form.description = ''
  form.priceYuan = '0.00'
  form.stock = 0
  form.imagesText = ''
  form.status = ProductStatus.ON
  dialogVisible.value = true
}

function openEdit(row: ProductInfo): void {
  editingId.value = row.id
  form.category_id = row.category_id
  form.name = row.name
  form.description = row.description
  form.priceYuan = formatFenToYuan(row.price)
  form.stock = row.stock
  form.imagesText = (row.images ?? []).join('\n')
  form.status = row.status
  dialogVisible.value = true
}

async function handleSave(): Promise<void> {
  if (!form.name.trim()) {
    ElMessage.warning('请填写商品名称')
    return
  }
  if (!form.category_id) {
    ElMessage.warning('请选择分类')
    return
  }
  saving.value = true
  try {
    const images = parseImages(form.imagesText)
    const price = yuanToFen(form.priceYuan)
    if (editingId.value == null) {
      await productApi.create({
        category_id: form.category_id,
        name: form.name.trim(),
        description: form.description,
        price,
        stock: form.stock,
        images,
        status: form.status,
      })
      ElMessage.success('创建成功')
    } else {
      const id = editingId.value
      await productApi.update({
        id,
        name: form.name.trim(),
        description: form.description,
        price,
        images,
      })
      if (form.category_id) {
        await productApi.updateCategory({ id, category_id: form.category_id })
      }
      await productApi.updateStock({ id, stock: form.stock })
      await productApi.setStatus({ id, status: form.status })
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

async function toggleStatus(row: ProductInfo): Promise<void> {
  const next = row.status === ProductStatus.ON ? ProductStatus.OFF : ProductStatus.ON
  try {
    await productApi.setStatus({ id: row.id, status: next })
    ElMessage.success(next === ProductStatus.ON ? '已上架' : '已下架')
    await loadList()
  } catch (error) {
    showError(error, '更新状态失败')
  }
}

async function quickStock(row: ProductInfo): Promise<void> {
  try {
    const { value } = await ElMessageBox.prompt('请输入库存数量', '调整库存', {
      inputValue: String(row.stock),
      inputPattern: /^\d+$/,
      inputErrorMessage: '请输入非负整数',
    })
    await productApi.updateStock({ id: row.id, stock: Number(value) })
    ElMessage.success('库存已更新')
    await loadList()
  } catch (error) {
    if (error === 'cancel' || error === 'close') {
      return
    }
    showError(error, '更新库存失败')
  }
}

onMounted(async () => {
  try {
    await loadCategories()
  } catch (error) {
    showError(error, '加载分类失败')
  }
  await loadList()
})
</script>

<template>
  <div class="page-card">
    <div class="page-toolbar">
      <el-input
        v-model="query.keyword"
        clearable
        placeholder="商品关键词"
        style="width: 180px"
        @keyup.enter="handleSearch"
      />
      <el-select v-model="query.category_id" clearable placeholder="分类" style="width: 160px">
        <el-option :value="0" label="全部分类" />
        <el-option
          v-for="c in categories"
          :key="c.id"
          :value="c.id"
          :label="c.name"
        />
      </el-select>
      <el-select v-model="query.status" clearable placeholder="状态" style="width: 120px">
        <el-option :value="0" label="全部状态" />
        <el-option :value="ProductStatus.ON" label="上架" />
        <el-option :value="ProductStatus.OFF" label="下架" />
      </el-select>
      <el-button type="primary" @click="handleSearch">查询</el-button>
      <el-button type="success" @click="openCreate">新建商品</el-button>
    </div>

    <el-table v-loading="loading" :data="list" stripe border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="名称" min-width="160" />
      <el-table-column label="分类" min-width="120">
        <template #default="{ row }">{{ categoryName(row.category_id) }}</template>
      </el-table-column>
      <el-table-column label="价格(元)" width="100">
        <template #default="{ row }">{{ formatFenToYuan(row.price) }}</template>
      </el-table-column>
      <el-table-column prop="stock" label="库存" width="80" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === ProductStatus.ON ? 'success' : 'info'" size="small">
            {{ productStatusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="更新时间" min-width="160">
        <template #default="{ row }">{{ formatUnix(row.ut) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="warning" @click="quickStock(row)">库存</el-button>
          <el-button link type="danger" @click="toggleStatus(row)">
            {{ row.status === ProductStatus.ON ? '下架' : '上架' }}
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
      v-model="dialogVisible"
      :title="editingId == null ? '新建商品' : '编辑商品'"
      width="560px"
    >
      <el-form label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.category_id" style="width: 100%">
            <el-option
              v-for="c in categories"
              :key="c.id"
              :value="c.id"
              :label="c.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="价格(元)">
          <el-input v-model="form.priceYuan" />
        </el-form-item>
        <el-form-item label="库存">
          <el-input-number v-model="form.stock" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option :value="ProductStatus.ON" label="上架" />
            <el-option :value="ProductStatus.OFF" label="下架" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="图片 URL">
          <el-input
            v-model="form.imagesText"
            type="textarea"
            :rows="3"
            placeholder="每行一个 URL，或逗号分隔（上传功能后续接入）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
