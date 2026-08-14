<template>
  <span class="hope-chip" :class="`hope-chip--${color}`">
    <slot />
  </span>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  color?: 'primary' | 'success' | 'warning' | 'error' | 'info' | 'accent'
}>(), {
  color: 'primary',
})
</script>
