import { showFailToast, showSuccessToast, showToast } from 'vant'
import { ApiError } from '@/api/errors/ApiError'

export function notifySuccess(message: string): void {
  showSuccessToast(message)
}

export function notifyFail(message: string): void {
  showFailToast(message)
}

export function notifyInfo(message: string): void {
  showToast(message)
}

export function notifyError(error: unknown, fallback = '请求失败'): void {
  if (error instanceof ApiError) {
    notifyFail(error.msg || fallback)
    return
  }
  if (error instanceof Error && error.message) {
    notifyFail(error.message)
    return
  }
  notifyFail(fallback)
}
