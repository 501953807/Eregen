import { defineStore } from 'pinia'
import { ref } from 'vue'
import { usersApi } from '@/api/users'
import type { User, ElderlyProfile } from '@/types'

export const useUsersStore = defineStore('users', () => {
  const familyUsers = ref<User[]>([])
  const elderlyProfiles = ref<ElderlyProfile[]>([])
  const loading = ref(false)

  async function fetchFamily(params?: Record<string, any>) {
    loading.value = true
    try {
      const res = await usersApi.list({ ...params, role: 'family' })
      familyUsers.value = res.data.data || []
    } finally {
      loading.value = false
    }
  }

  async function fetchElderly(params?: Record<string, any>) {
    loading.value = true
    try {
      const res = await usersApi.listElderly(params)
      elderlyProfiles.value = res.data.data || []
    } finally {
      loading.value = false
    }
  }

  return { familyUsers, elderlyProfiles, loading, fetchFamily, fetchElderly }
})
