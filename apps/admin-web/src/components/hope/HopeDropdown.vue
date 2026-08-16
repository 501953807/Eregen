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

<style scoped>
.hope-dropdown {
  position: relative;
  display: inline-block;
}
.hope-dropdown.open .hope-dropdown-menu {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}
.hope-dropdown-trigger {
  cursor: pointer;
  display: inline-flex;
  align-items: center;
}
.hope-dropdown-menu {
  position: absolute;
  z-index: 50;
  min-width: 180px;
  background: var(--hope-surface);
  border: 1px solid var(--hope-border);
  border-radius: var(--hope-radius-md);
  box-shadow: var(--hope-shadow-lg);
  padding: 6px;
  opacity: 0;
  visibility: hidden;
  transform: translateY(4px);
  transition: all 0.15s ease;
}
.hope-dropdown--up .hope-dropdown-menu {
  bottom: 100%;
  top: auto;
  margin-bottom: 6px;
  transform: translateY(4px);
}
.hope-dropdown.open.hope-dropdown--up .hope-dropdown-menu {
  transform: translateY(0);
}
.hope-dropdown-menu > * {
  display: block;
  width: 100%;
}
</style>
