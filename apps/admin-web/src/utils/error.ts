import { ElMessage } from 'element-plus'

export function handleApiError(error: unknown, fallbackMsg = '操作失败，请稍后重试'): void {
  if (typeof error !== 'object' || error === null) {
    ElMessage.error(fallbackMsg)
    return
  }
  const e = error as Record<string, unknown>
  const message = (e.message as string) || (e.error as string) || fallbackMsg
  ElMessage.error(message)
}

export function handleApiSuccess(msg = '操作成功'): void {
  ElMessage.success(msg)
}
