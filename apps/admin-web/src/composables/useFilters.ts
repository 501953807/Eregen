import { ref } from 'vue'

export interface FilterState<T = string> {
  [key: string]: T | T[] | null
}

export function useFilters<T extends FilterState>(initial: T) {
  const filters = ref<T>({ ...initial })

  function setFilter<K extends keyof T>(key: K, value: T[K]) {
    filters.value[key] = value
  }

  function reset() {
    filters.value = { ...initial }
  }

  return { filters, setFilter, reset }
}
