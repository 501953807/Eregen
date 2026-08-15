<template>
  <div class="hope-content-card">
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
}>()
</script>

<style scoped>
.hope-content-card {
  background: var(--hope-surface);
  border-radius: var(--hope-radius-lg);
  border: 1px solid var(--hope-border);
  box-shadow: var(--hope-shadow-card);
  overflow: hidden;
  transition: box-shadow 0.2s ease, transform 0.2s ease;
}
.hope-content-card:hover {
  box-shadow: var(--hope-shadow-card-hover);
  transform: translateY(-2px);
}
.hope-content-card__header {
  padding: 20px 22px 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.hope-content-card__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--hope-text);
  margin: 0;
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
