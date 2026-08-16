<template>
  <div class="hope-alert" :class="`hope-alert--${color}`">
    <div class="hope-alert__icon">
      <slot name="icon" v-if="!$slots.icon" />
      <slot name="icon" v-else />
    </div>
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

<style scoped>
.hope-alert {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  border-radius: var(--hope-radius-md);
  border: 1px solid transparent;
  margin-bottom: 16px;
}
.hope-alert__icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  margin-top: 2px;
  color: currentColor;
}
.hope-alert__content {
  flex: 1;
}
.hope-alert__title {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 4px;
}
.hope-alert__close {
  width: 24px;
  height: 24px;
  border-radius: var(--hope-radius-sm);
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: currentColor;
  opacity: 0.6;
  transition: opacity 0.2s;
}
.hope-alert__close:hover { opacity: 1; }

/* Variant styles */
.hope-alert--primary {
  background: var(--hope-primary-light);
  border-color: var(--hope-primary-light);
  color: var(--hope-primary);
}
.hope-alert--success {
  background: var(--hope-success-light);
  border-color: var(--hope-success-light);
  color: var(--hope-success);
}
.hope-alert--warning {
  background: var(--hope-warning-light);
  border-color: var(--hope-warning-light);
  color: var(--hope-warning);
}
.hope-alert--error {
  background: var(--hope-danger-light);
  border-color: var(--hope-danger-light);
  color: var(--hope-danger);
}
.hope-alert--info {
  background: var(--hope-info-light);
  border-color: var(--hope-info-light);
  color: var(--hope-info);
}
.hope-alert--accent {
  background: rgba(140,87,255,0.1);
  border-color: rgba(140,87,255,0.2);
  color: var(--hope-accent);
}
</style>
