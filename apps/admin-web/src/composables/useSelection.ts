import { ref, computed } from 'vue'

export interface SelectionState<T = string> {
  selectedIds: T[]
  allSelected: boolean
  indeterminate: boolean
}

export function useSelection<T = string>() {
  const selectedIds = ref<T[]>([])
  const allSelected = ref(false)
  const indeterminate = ref(false)

  const selectedCount = computed(() => selectedIds.value.length)

  function toggleSelectAll(val: boolean, items: T[]) {
    if (val) {
      selectedIds.value = [...new Set([...selectedIds.value, ...items])] as T[]
    } else {
      selectedIds.value = []
    }
    allSelected.value = val
    indeterminate.value = false
  }

  function toggleRow(row: T, selected: boolean) {
    if (selected) {
      const ids = selectedIds.value as unknown as T[]
      if (!ids.includes(row)) {
        selectedIds.value = [...ids, row]
      }
    } else {
      const ids = selectedIds.value as unknown as T[]
      selectedIds.value = ids.filter(id => id !== row)
    }
    indeterminate.value = selectedIds.value.length > 0 && !allSelected.value
    allSelected.value = false
  }

  function clearSelection() {
    selectedIds.value = []
    allSelected.value = false
    indeterminate.value = false
  }

  function isSelected(id: T): boolean {
    return (selectedIds.value as unknown as T[]).includes(id)
  }

  return { selectedIds, allSelected, indeterminate, selectedCount, toggleSelectAll, toggleRow, clearSelection, isSelected }
}
