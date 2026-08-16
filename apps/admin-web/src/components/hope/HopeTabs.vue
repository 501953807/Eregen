<template>
  <div class="hope-tabs">
    <div class="hope-tabs__nav" role="tablist">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        role="tab"
        :aria-selected="modelValue === tab.value"
        :class="['hope-tabs__tab', { 'active': modelValue === tab.value }]"
        @click="$emit('update:modelValue', tab.value)"
      >
        <span v-if="tab.icon" class="hope-tabs__tab-icon">
          <slot :name="`icon-${tab.value}`" />
        </span>
        {{ tab.label }}
        <span v-if="tab.badge" class="hope-tabs__badge">{{ tab.badge }}</span>
      </button>
    </div>
    <div class="hope-tabs__content" role="tabpanel">
      <slot :name="`content-${modelValue}`" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: string | number
  tabs: Array<{ value: string | number; label: string; badge?: string | number; icon?: string }>
}>()

defineEmits<{ 'update:modelValue': [value: string | number] }>()
</script>

<style scoped>
.hope-tabs {
  width: 100%;
}
.hope-tabs__nav {
  display: flex;
  gap: 4px;
  border-bottom: 2px solid var(--hope-border);
  margin-bottom: 20px;
}
.hope-tabs__tab {
  padding: 12px 20px;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: var(--hope-text-secondary);
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
  transition: color 0.2s ease;
  font-family: inherit;
}
.hope-tabs__tab:hover {
  color: var(--hope-primary);
}
.hope-tabs__tab.active {
  color: var(--hope-primary);
  font-weight: 600;
}
.hope-tabs__tab.active::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--hope-primary);
  border-radius: 2px 2px 0 0;
}
.hope-tabs__tab-icon {
  width: 16px;
  height: 16px;
}
.hope-tabs__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 10px;
  background: var(--hope-danger);
  color: white;
  font-size: 11px;
  font-weight: 600;
}
.hope-tabs__content {
  padding: 4px 0;
}
</style>
