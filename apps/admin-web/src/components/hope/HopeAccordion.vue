<template>
  <div class="hope-accordion">
    <div
      v-for="(item, i) in items"
      :key="i"
      class="hope-accordion-item"
      :class="{ active: openIndex === i }"
    >
      <button class="hope-accordion-header" @click="openIndex = openIndex === i ? null : i">
        <span>{{ item.title }}</span>
        <svg class="hope-accordion-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="6 9 12 15 18 9"/></svg>
      </button>
      <div class="hope-accordion-body">
        <slot :name="`body-${i}`" :item="item" :index="i" v-if="$slots[`body-${i}`]" />
        <template v-else>{{ item.body }}</template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const openIndex = ref<number | null>(null)

defineProps<{
  items: Array<{ title: string; body?: string }>
}>()
</script>

<style scoped>
.hope-accordion {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.hope-accordion-item {
  border: 1px solid var(--hope-border);
  border-radius: var(--hope-radius-md);
  overflow: hidden;
  transition: border-color 0.2s;
}
.hope-accordion-item.active {
  border-color: var(--hope-primary);
  box-shadow: 0 0 0 3px rgba(58,87,232,0.08);
}
.hope-accordion-header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: var(--hope-surface);
  border: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: var(--hope-text);
  text-align: left;
  font-family: inherit;
  transition: background 0.2s;
}
.hope-accordion-header:hover {
  background: var(--hope-surface-light);
}
.hope-accordion-icon {
  color: var(--hope-text-muted);
  transition: transform 0.25s ease;
  flex-shrink: 0;
}
.hope-accordion-item.active .hope-accordion-icon {
  transform: rotate(180deg);
  color: var(--hope-primary);
}
.hope-accordion-body {
  padding: 0 16px;
  max-height: 0;
  overflow: hidden;
  transition: max-height 0.3s ease, padding 0.3s ease;
  background: var(--hope-surface);
  color: var(--hope-text-secondary);
  font-size: 14px;
  line-height: 1.6;
}
.hope-accordion-item.active .hope-accordion-body {
  padding: 12px 16px 16px;
  max-height: 500px;
}
</style>
