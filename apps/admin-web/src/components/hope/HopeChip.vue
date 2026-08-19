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

<style scoped>
.hope-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: var(--hope-radius-pill);
  font-size: 12px;
  font-weight: 500;
  gap: 6px;
}
.hope-chip--primary   { background: rgba(var(--hope-primary-rgb), 0.1); color: var(--hope-primary); }
.hope-chip--success   { background: rgba(var(--hope-success-rgb), 0.1); color: var(--hope-success); }
.hope-chip--warning   { background: rgba(var(--hope-warning-rgb), 0.1); color: var(--hope-warning); }
.hope-chip--error     { background: rgba(var(--hope-error-rgb), 0.1); color: var(--hope-error); }
.hope-chip--info      { background: rgba(var(--hope-info-rgb, 7,154,162), 0.1); color: var(--hope-info, #079aa2); }
.hope-chip--accent    { background: rgba(var(--hope-accent-rgb, 140,87,255), 0.1); color: var(--hope-accent, #8C57FF); }
</style>
