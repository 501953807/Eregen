<template>
  <div class="hope-stat-card" :style="`--hope-primary-gradient: ${gradient || 'linear-gradient(135deg, #4A7C5F, #6FAF8F)'}`">
    <div class="hope-stat-card__icon" :class="`hope-stat-card__icon--${iconColor}`">
      <slot name="icon" />
    </div>
    <div class="hope-stat-card__value">{{ value }}</div>
    <div class="hope-stat-card__label">{{ label }}</div>
    <slot name="trend" />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  value: string | number
  label: string
  iconColor?: 'primary' | 'success' | 'warning' | 'error' | 'info' | 'accent'
  gradient?: string
}>(), {
  iconColor: 'primary',
})
</script>
