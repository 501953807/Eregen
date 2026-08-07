import { ref, computed } from 'vue'
import type { Ref } from 'vue'

export interface LoadingState {
  [key: string]: boolean
}

export function useLoading(initialKeys: string[] = []) {
  const states = ref<LoadingState>(
    initialKeys.reduce((acc, key) => {
      acc[key] = false
      return acc
    }, {} as LoadingState)
  )

  const loading = computed(() => Object.values(states.value).some(Boolean))

  function set(key: string, value: boolean) {
    states.value[key] = value
  }

  async function withLoading<T>(key: string, fn: () => Promise<T>): Promise<T> {
    set(key, true)
    try {
      return await fn()
    } finally {
      set(key, false)
    }
  }

  function reset() {
    for (const key of initialKeys) {
      states.value[key] = false
    }
  }

  return { states, loading, set, withLoading, reset }
}
