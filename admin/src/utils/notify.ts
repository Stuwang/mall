import { ElMessage } from 'element-plus'
import { ApiError } from '@/api/errors/ApiError'

export function showError(error: unknown, fallback = '操作失败'): void {
  if (error instanceof ApiError) {
    ElMessage.error(error.msg || fallback)
    return
  }
  if (error instanceof Error) {
    ElMessage.error(error.message || fallback)
    return
  }
  ElMessage.error(fallback)
}
