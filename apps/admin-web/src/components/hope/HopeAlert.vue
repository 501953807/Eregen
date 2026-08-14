<template>
  <div class="hope-alert" :class="`hope-alert--${color}`">
    <slot name="icon" v-if="$slots.icon" />
    <div class="hope-alert__content">
      <div v-if="title" class="hope-alert__title">{{ title }}</div>
      <slot />
    </div>
    <button v-if="closable" class="hope-alert__close" @click="$emit('close')">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
    </button>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  color?: 'primary' | 'success' | 'warning' | 'error' | 'info' | 'accent'
  title?: string
  closable?: boolean
}>(), {
  color: 'primary',
  closable: false,
})
defineEmits<{ close: [] }>()
</script>
