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
      const existing = new Set(selectedIds.value)
      items.forEach(id => existing.add(id as unknown as T))
      selectedIds.value = [...existing] as T[]
    } else {
      selectedIds.value = []
    }
    allSelected.value = val
    indeterminate.value = false
  }

  function toggleRow(row: T, selected: boolean) {
    if (selected) {
      if (!selectedIds.value.includes(row as unknown as T)) {
        selectedIds.value = [...selectedIds.value, row as unknown as T]
      }
    } else {
      selectedIds.value = selectedIds.value.filter(id => id !== row)
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
    return selectedIds.value.includes(id as unknown as T)
  }

  return { selectedIds, allSelected, indeterminate, selectedCount, toggleSelectAll, toggleRow, clearSelection, isSelected }
}
