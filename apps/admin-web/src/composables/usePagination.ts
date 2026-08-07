import { ref, computed } from 'vue'

export interface PaginationState {
  page: number
  pageSize: number
  total: number
}

export function usePagination(initialPageSize = 20) {
  const state = ref<PaginationState>({
    page: 1,
    pageSize: initialPageSize,
    total: 0,
  })

  const totalPages = computed(() => Math.ceil(state.value.total / state.value.pageSize))
  const hasNext = computed(() => state.value.page < totalPages.value)
  const hasPrev = computed(() => state.value.page > 1)

  function setPage(page: number) {
    state.value.page = Math.max(1, Math.min(page, totalPages.value))
  }

  function setPageSize(size: number) {
    state.value.pageSize = size
    state.value.page = 1
  }

  function setTotal(total: number) {
    state.value.total = total
  }

  function reset() {
    state.value = { page: 1, pageSize: initialPageSize, total: 0 }
  }

  return { state, totalPages, hasNext, hasPrev, setPage, setPageSize, setTotal, reset }
}
