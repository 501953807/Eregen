import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('admin_token') || '')
  const user = ref<{ name: string; role: string } | null>(null)

  function login(t: string, u: any) {
    token.value = t
    user.value = u
    localStorage.setItem('admin_token', t)
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('admin_token')
  }

  function hasPermission(resource: string): boolean {
    if (!user.value) return false
    return user.value.role === 'super_admin' || user.value.role === 'admin'
  }

  return { token, user, login, logout, hasPermission }
})
