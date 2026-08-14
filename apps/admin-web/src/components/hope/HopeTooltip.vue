<template>
  <div class="hope-tooltip">
    <span class="hope-tooltip-trigger">
      <slot />
    </span>
    <div class="hope-tooltip-content">{{ content }}</div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  content: string
}>()
</script>
