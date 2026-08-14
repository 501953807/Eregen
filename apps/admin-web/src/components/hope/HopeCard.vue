<template>
  <div class="hope-content-card">
    <div v-if="$slots.header || title" class="hope-content-card__header">
      <div>
        <div v-if="title" class="hope-content-card__title">{{ title }}</div>
        <div v-if="subtitle" class="hope-content-card__subtitle">{{ subtitle }}</div>
      </div>
      <slot name="header" />
    </div>
    <div class="hope-content-card__body">
      <slot />
    </div>
    <div v-if="$slots.footer" class="hope-content-card__footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  subtitle?: string
}>()
</script>
