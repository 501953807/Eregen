<template>
  <div class="hope-stat-card" :style="{ '--hope-stat-gradient': gradient || 'linear-gradient(135deg, #3a57e8, #6f42c1)' }">
    <div class="hope-stat-card__icon-wrap" :class="`hope-stat-card__icon-wrap--${iconColor}`">
      <div class="hope-stat-card__icon">
        <slot name="icon" />
      </div>
    </div>
    <div class="hope-stat-card__content">
      <div class="hope-stat-card__value">{{ value }}</div>
      <div class="hope-stat-card__label">{{ label }}</div>
      <slot name="trend" />
    </div>
    <div class="hope-stat-card__progress" v-if="$slots.progress">
      <slot name="progress" />
    </div>
    <div class="hope-stat-card__bg-decoration"></div>
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

<style scoped>
.hope-stat-card {
  background: var(--hope-surface);
  border-radius: var(--hope-radius-lg);
  border: none;
  padding: 22px;
  transition: box-shadow 0.25s ease, transform 0.25s ease;
  position: relative;
  overflow: hidden;
  box-shadow: 0 10px 30px rgba(17,38,146,0.05);
}
.hope-stat-card::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 3px;
  background: var(--hope-stat-gradient);
  border-radius: var(--hope-radius-lg) var(--hope-radius-lg) 0 0;
  opacity: 1;
}
.hope-stat-card:hover {
  box-shadow: 0 16px 48px rgba(17,38,146,0.08);
  transform: translateY(-2px);
}
.hope-stat-card__icon-wrap {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
  position: relative;
  z-index: 1;
}
.hope-stat-card__icon-wrap--primary   { background: rgba(var(--hope-primary-rgb), 0.12); color: var(--hope-primary); }
.hope-stat-card__icon-wrap--success  { background: rgba(var(--hope-success-rgb), 0.12); color: var(--hope-success); }
.hope-stat-card__icon-wrap--warning  { background: rgba(var(--hope-warning-rgb), 0.12); color: var(--hope-warning); }
.hope-stat-card__icon-wrap--error    { background: rgba(var(--hope-error-rgb), 0.12); color: var(--hope-error); }
.hope-stat-card__icon-wrap--info     { background: rgba(var(--hope-info-rgb, 7,154,162), 0.12); color: var(--hope-info, #079aa2); }
.hope-stat-card__icon-wrap--accent   { background: rgba(var(--hope-accent-rgb, 140,87,255), 0.12); color: var(--hope-accent, #8C57FF); }
.hope-stat-card__icon {
  font-size: 24px;
  width: 24px;
  height: 24px;
}
.hope-stat-card__content {
  position: relative;
  z-index: 1;
}
.hope-stat-card__value {
  font-size: 30px;
  font-weight: 800;
  color: var(--hope-text);
  line-height: 1.1;
  margin-bottom: 4px;
  letter-spacing: -0.03em;
}
.hope-stat-card__label {
  font-size: 13px;
  color: var(--hope-text-muted);
  font-weight: 500;
}
.hope-stat-card__trend {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 600;
  margin-top: 10px;
  padding: 3px 8px;
  border-radius: 50px;
}
.hope-stat-card__trend-up   { background: rgba(var(--hope-success-rgb), 0.10); color: var(--hope-success); }
.hope-stat-card__trend-down { background: rgba(var(--hope-error-rgb), 0.10); color: var(--hope-error); }
.hope-stat-card__trend-neutral { background: rgba(var(--hope-text-muted-rgb, 148,169,162), 0.10); color: var(--hope-text-muted); }
.hope-stat-card__progress {
  margin-top: 14px;
  display: flex;
  align-items: center;
}
.hope-stat-card__bg-decoration {
  position: absolute;
  top: -20px;
  right: -20px;
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: var(--hope-stat-gradient);
  opacity: 0.05;
  z-index: 0;
}
</style>
