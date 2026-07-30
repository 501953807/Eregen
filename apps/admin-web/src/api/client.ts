import axios, { type AxiosRequestConfig, type AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'
// Removed top-level import of useAuthStore to avoid circular dependency during module initialization

// Determine base URL: in dev, use explicit backend URL; in prod, use relative API path
let baseURL: string
if (import.meta.env.DEV) {
  // In development, connect directly to admin-api server
  baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8089'
} else {
  // In production, use relative path served by same origin
  baseURL = '/api/v1'
}

// API client instance
const apiClient = axios.create({
  baseURL,
  timeout: 60000,
})

// Request interceptor - add Authorization header (lazy access to authStore via function)
apiClient.interceptors.request.use(async (config) => {
  // Lazily load the authStore when needed to avoid circular dependencies
  const { useAuthStore } = await import('@/stores/auth')
  const authStore = useAuthStore()
  const token = authStore.token
  if (token && config.headers) {
    config.headers['Authorization'] = `Bearer ${token}`
  }
  return config
}, (error: AxiosError) => Promise.reject(error))

// Response interceptor - handle errors
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const statusCode = error.response?.status

    // Handle 401 Unauthorized
    if (statusCode === 401) {
      // Lazily load authStore for logout
      const { useAuthStore } = await import('@/stores/auth')
      const authStore = useAuthStore()
      authStore.logout()

      // Don't show message if user is already on login page trying to refresh
      if (window.location.pathname !== '/login') {
        ElMessage.warning('会话已过期，请重新登录')
      }

      const redirectPath = error.config?.params?.redirect || '/'
      return router.push({ path: '/login', query: { redirect: redirectPath } })
    }

    // Handle 403 Forbidden
    if (statusCode === 403) {
      ElMessage.error('无访问权限')
      return router.push('/login')
    }

    // Handle 500 Server Error
    if (statusCode >= 500) {
      ElMessage.error('服务器内部错误，请稍后重试')
    }

    // Return the error response data message if available
    if (error.response?.data?.error) {
      ElMessage.error(error.response.data.error)
    }

    return Promise.reject(error)
  }
)

export default apiClient
