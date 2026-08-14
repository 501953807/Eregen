<template>
  <span
    class="hope-badge"
    :class="`hope-badge--${color}`"
    :data-type="type"
  >
    <span v-if="type === 'dot'" class="hope-badge--dot"></span>
    <slot />
  </span>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  color?: 'primary' | 'success' | 'warning' | 'error' | 'info' | 'accent'
  type?: 'text' | 'dot'
}>(), {
  color: 'primary',
  type: 'text',
})
</script>
