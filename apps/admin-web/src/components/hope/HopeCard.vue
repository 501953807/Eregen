<template>
  <div class="hope-content-card" :class="{ 'hope-content-card--elevated': elevated }">
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
  elevated?: boolean
}>()
</script>

<style scoped>
.hope-content-card {
  background: var(--hope-surface);
  border-radius: var(--hope-radius-lg);
  border: none;
  box-shadow: 0 10px 30px rgba(17,38,146,0.05);
  overflow: hidden;
  transition: box-shadow 0.25s ease, transform 0.25s ease;
}
.hope-content-card:hover {
  box-shadow: 0 16px 48px rgba(17,38,146,0.08);
  transform: translateY(-2px);
}
.hope-content-card--elevated {
  box-shadow: 0 16px 48px rgba(17,38,146,0.08);
}
.hope-content-card__header {
  padding: 20px 22px 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}
.hope-content-card__title {
  font-size: 15px;
  font-weight: 600;
  color: var(--hope-text);
  margin: 0;
  letter-spacing: -0.01em;
}
.hope-content-card__subtitle {
  font-size: 13px;
  color: var(--hope-text-muted);
  margin-top: 2px;
}
.hope-content-card__body { padding: 16px 22px; }
.hope-content-card__footer {
  padding: 14px 22px;
  border-top: 1px solid var(--hope-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
</style>
