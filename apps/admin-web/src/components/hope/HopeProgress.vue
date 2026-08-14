<template>
  <div class="hope-progress" :class="`hope-progress--${size}`">
    <div class="hope-progress__label" v-if="$slots.label || showLabel">
      <span>{{ label }}</span>
      <span class="hope-progress__value">{{ value }}%</span>
    </div>
    <div class="hope-progress__bar" :class="`hope-progress--${color}`" :style="{ width: value + '%' }">
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  value: number
  color?: 'primary' | 'success' | 'warning' | 'error' | 'info' | 'accent'
  size?: 'sm' | 'md' | 'lg'
  label?: string
  showLabel?: boolean
}>(), {
  color: 'primary',
  size: 'md',
  showLabel: false,
})
</script>
