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

<style scoped>
.hope-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: var(--hope-radius-pill);
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
  line-height: 1.5;
  transition: all 0.2s ease;
}
.hope-badge--dot { padding: 0; width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.hope-badge--primary   { background: rgba(58,87,232,0.12); color: var(--hope-primary); }
.hope-badge--success   { background: rgba(26,160,83,0.12); color: #1aa053; }
.hope-badge--warning   { background: rgba(250,169,56,0.12); color: var(--hope-warning); }
.hope-badge--error     { background: rgba(192,50,33,0.12); color: #c03221; }
.hope-badge--info      { background: rgba(7,154,162,0.12); color: #079aa2; }
.hope-badge--accent    { background: rgba(140,87,255,0.12); color: var(--hope-accent); }
</style>
