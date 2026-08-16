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
.hope-chip--primary   { background: rgba(58,87,232,0.1); color: var(--hope-primary); }
.hope-chip--success   { background: rgba(26,160,83,0.1); color: #1aa053; }
.hope-chip--warning   { background: rgba(250,169,56,0.1); color: var(--hope-warning); }
.hope-chip--error     { background: rgba(192,50,33,0.1); color: #c03221; }
.hope-chip--info      { background: rgba(7,154,162,0.1); color: #079aa2; }
.hope-chip--accent    { background: rgba(140,87,255,0.1); color: var(--hope-accent); }
</style>
