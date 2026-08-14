<template>
  <div class="hope-dropdown" :class="{ open: open, [`hope-dropdown--${direction}`]: direction }">
    <div ref="triggerRef" class="hope-dropdown-trigger" @click="toggle">
      <slot name="trigger" />
    </div>
    <div class="hope-dropdown-menu" ref="menuRef">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps<{
  direction?: 'up' | 'down'
}>()

const open = ref(false)
const triggerRef = ref<HTMLElement | null>(null)
const menuRef = ref<HTMLElement | null>(null)

function toggle() {
  open.value = !open.value
}

function close() {
  open.value = false
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})

function handleClickOutside(e: MouseEvent) {
  if (
    triggerRef.value && !triggerRef.value.contains(e.target as Node) &&
    menuRef.value && !menuRef.value.contains(e.target as Node)
  ) {
    close()
  }
}
</script>
