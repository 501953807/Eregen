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

<style scoped>
.hope-progress {
  width: 100%;
}
.hope-progress__label {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  font-weight: 500;
  color: var(--hope-text-secondary);
  margin-bottom: 6px;
}
.hope-progress__value {
  font-weight: 600;
  color: var(--hope-text);
}
.hope-progress__bar {
  height: 6px;
  border-radius: var(--hope-radius-pill);
  background: var(--hope-border);
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}
.hope-progress__bar::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.3), transparent);
  animation: progress-shine 2s infinite;
}
@keyframes progress-shine {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

/* Sizes */
.hope-progress--sm .hope-progress__bar { height: 4px; }
.hope-progress--lg .hope-progress__bar { height: 10px; }

/* Variants */
.hope-progress--primary .hope-progress__bar   { background: var(--hope-primary); }
.hope-progress--success .hope-progress__bar   { background: #1aa053; }
.hope-progress--warning .hope-progress__bar   { background: var(--hope-warning); }
.hope-progress--error .hope-progress__bar     { background: var(--hope-error); }
.hope-progress--info .hope-progress__bar      { background: #079aa2; }
.hope-progress--accent .hope-progress__bar    { background: var(--hope-accent); }
</style>
