import { ref } from 'vue'

export interface DialogState<T = unknown> {
  visible: boolean
  data: T | null
  mode: 'create' | 'edit'
}

export function useDialog<T = unknown>() {
  const state = ref<DialogState<T>>({
    visible: false,
    data: null,
    mode: 'create',
  })

  function open(mode: 'create' | 'edit' = 'create', data?: T) {
    state.value = { visible: true, data: data ?? null, mode }
  }

  function close() {
    state.value = { visible: false, data: null, mode: 'create' }
  }

  function reset() {
    state.value = { visible: false, data: null, mode: 'create' }
  }

  return { state, open, close, reset }
}
