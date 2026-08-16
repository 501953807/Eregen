<template>
  <div class="hope-tooltip">
    <span class="hope-tooltip-trigger">
      <slot />
    </span>
    <div class="hope-tooltip-content">{{ content }}</div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  content: string
}>()
</script>

<style scoped>
.hope-tooltip {
  position: relative;
  display: inline-block;
}
.hope-tooltip-trigger {
  cursor: default;
}
.hope-tooltip-content {
  position: absolute;
  z-index: 100;
  bottom: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%) translateY(4px);
  background: var(--hope-text);
  color: white;
  font-size: 12px;
  padding: 6px 12px;
  border-radius: var(--hope-radius-sm);
  white-space: nowrap;
  opacity: 0;
  visibility: hidden;
  transition: all 0.15s ease;
  pointer-events: none;
}
.hope-tooltip-content::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 5px solid transparent;
  border-top-color: var(--hope-text);
}
.hope-tooltip:hover .hope-tooltip-content {
  opacity: 1;
  visibility: visible;
  transform: translateX(-50%) translateY(0);
}
</style>
