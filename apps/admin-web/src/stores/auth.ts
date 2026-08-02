import { defineStore } from 'pinia'
import { ref, onMounted } from 'vue'
import router from '@/router'
// Use the configured apiClient which has proper baseURL for dev/prod environments
import apiClient from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('admin_token'))
  const user = ref<{ name: string; role: string } | null>(localStorage.getItem('admin_user') ? JSON.parse(localStorage.getItem('admin_user')!) : null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('admin_token', t)
  }

  function setUser(u: any) {
    user.value = u
    localStorage.setItem('admin_user', JSON.stringify(u))
  }

  function login(payload: { method: 'email' | 'phone'; credential: string; secret: string }) {
    return new Promise<void>((resolve, reject) => {
      loading.value = true
      error.value = null

      apiClient.post('/auth/login', payload)
        .then(response => {
          const body = response.data as any
          const loginData = body?.data
          if (!loginData) {
            throw new Error(body?.msg || '登录失败')
          }
          const { token, user } = loginData
          if (!token) throw new Error('未获取到 token')
          setToken(token)
          setUser(user)
          resolve()
        })
        .catch(err => {
          loading.value = false
          const errorMsg = err?.response?.data?.msg || err?.response?.data?.error || (err?.message || '登录失败，请重试')
          error.value = errorMsg
          reject(err)
        })
    })
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
    router.push('/login')
  }

  function hasPermission(resource: string): boolean {
    if (!user.value) return false
    return user.value.role === 'super_admin' || user.value.role === 'admin'
  }

  function isLoggedIn(): boolean {
    return token.value !== null && token.value !== ''
  }

  // On store initialization, check if there's a saved token
  onMounted(() => {
    if (token.value && !user.value) {
      // Try to refresh user info from stored data or just mark as logged in
      // In a real app, you might call /api/v1/auth/me here
    }
  })

  return {
    token,
    user,
    loading,
    error,
    setToken,
    setUser,
    login,
    logout,
    hasPermission,
    isLoggedIn,
  }
})
