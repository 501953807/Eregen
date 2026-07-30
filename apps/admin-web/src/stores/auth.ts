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

  function login(username: string, password: string) {
    return new Promise<void>((resolve, reject) => {
      loading.value = true
      error.value = null

      // Call the API endpoint using configured apiClient (has baseURL)
      apiClient.post('/api/v1/auth/login', { username, password })
        .then(response => {
          const { token: jwtToken, user: userInfo } = response.data.data
          setToken(jwtToken)
          setUser(userInfo)
          resolve()
        })
        .catch(err => {
          loading.value = false
          const errorMsg = err.response?.data?.error || (err.message || 'Login failed. Please check username and password.')
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
