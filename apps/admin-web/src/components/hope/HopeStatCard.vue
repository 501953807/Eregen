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

<style scoped>
.hope-stat-card {
  background: var(--hope-surface);
  border-radius: var(--hope-radius-lg);
  border: 1px solid var(--hope-border);
  padding: 22px;
  transition: box-shadow 0.2s ease, transform 0.2s ease;
  position: relative;
  overflow: hidden;
}
.hope-stat-card::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 3px;
  background: var(--hope-primary-gradient);
  border-radius: var(--hope-radius-lg) var(--hope-radius-lg) 0 0;
  opacity: 0;
  transition: opacity 0.2s;
}
.hope-stat-card:hover {
  box-shadow: var(--hope-shadow-md);
  transform: translateY(-2px);
}
.hope-stat-card:hover::before { opacity: 1; }
.hope-stat-card__icon {
  width: 52px; height: 52px;
  border-radius: 14px;
  display: flex; align-items: center; justify-content: center;
  font-size: 24px;
  margin-bottom: 16px;
}
.hope-stat-card__icon--primary   { background: rgba(58,87,232,0.12); color: #3a57e8; }
.hope-stat-card__icon--success  { background: rgba(26,160,83,0.12); color: #1aa053; }
.hope-stat-card__icon--warning  { background: rgba(250,169,56,0.12); color: #FAA938; }
.hope-stat-card__icon--error    { background: rgba(192,50,33,0.12); color: #c03221; }
.hope-stat-card__icon--info     { background: rgba(7,154,162,0.12); color: #079aa2; }
.hope-stat-card__icon--accent   { background: rgba(140,87,255,0.12); color: #8C57FF; }
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
  border-radius: var(--hope-radius-pill);
}
.hope-stat-card__trend-up   { background: var(--hope-success-light); color: #2D5AA0; }
.hope-stat-card__trend-down { background: var(--hope-error-light); color: #8B2020; }
.hope-stat-card__trend-neutral { background: rgba(148,169,162,0.12); color: #6b7280; }
</style>
