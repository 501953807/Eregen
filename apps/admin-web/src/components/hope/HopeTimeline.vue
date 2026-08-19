<template>
  <div class="hope-timeline" :class="{ 'hope-timeline--horizontal': horizontal }">
    <div
      v-for="(item, i) in items"
      :key="i"
      class="hope-timeline-item"
    >
      <div class="hope-timeline-dot" :class="`hope-timeline-dot--${item.color || 'primary'}`">
      </div>
      <div class="hope-timeline-content">
        <div v-if="item.title" class="hope-timeline-title">{{ item.title }}</div>
        <div v-if="item.meta" class="hope-timeline-meta">
          <slot name="meta" :item="item" :index="i" />
          <template v-if="!$slots.meta">{{ item.meta }}</template>
        </div>
        <div class="hope-timeline-body">
          <slot :name="`body-${i}`" :item="item" :index="i" v-if="$slots[`body-${i}`]" />
          <template v-else>{{ item.body }}</template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  items: Array<{
    title?: string
    meta?: string
    body?: string
    color?: 'primary' | 'success' | 'warning' | 'error' | 'info'
  }>
  horizontal?: boolean
}>()
</script>

<style scoped>
.hope-timeline {
  display: flex;
  flex-direction: column;
  gap: 0;
  position: relative;
  padding-left: 28px;
}
.hope-timeline--horizontal {
  flex-direction: row;
  padding-left: 0;
  padding-top: 28px;
  gap: 24px;
  overflow-x: auto;
}
.hope-timeline-item {
  position: relative;
  padding-bottom: 24px;
}
.hope-timeline-item:last-child {
  padding-bottom: 0;
}
.hope-timeline-dot {
  position: absolute;
  left: -28px;
  top: 2px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--hope-primary);
  box-shadow: 0 0 0 3px rgba(var(--hope-primary-rgb), 0.15);
  z-index: 1;
}
.hope-timeline-dot--success { background: var(--hope-success); box-shadow: 0 0 0 3px rgba(var(--hope-success-rgb), 0.15); }
.hope-timeline-dot--warning { background: var(--hope-warning); box-shadow: 0 0 0 3px rgba(var(--hope-warning-rgb), 0.15); }
.hope-timeline-dot--error   { background: var(--hope-error); box-shadow: 0 0 0 3px rgba(var(--hope-error-rgb), 0.15); }
.hope-timeline-dot--info    { background: #079aa2; box-shadow: 0 0 0 3px rgba(7,154,162,0.15); }

/* Connector line */
.hope-timeline-item:not(:last-child)::before {
  content: '';
  position: absolute;
  left: -23px;
  top: 14px;
  bottom: -10px;
  width: 2px;
  background: var(--hope-border);
}

.hope-timeline-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--hope-text);
  margin-bottom: 2px;
}
.hope-timeline-meta {
  font-size: 12px;
  color: var(--hope-text-muted);
  margin-bottom: 4px;
}
.hope-timeline-body {
  font-size: 13px;
  color: var(--hope-text-secondary);
  line-height: 1.6;
}

/* Horizontal mode */
.hope-timeline--horizontal .hope-timeline-dot {
  left: 50%;
  top: -28px;
  transform: translateX(-50%);
}
.hope-timeline--horizontal .hope-timeline-item:not(:last-child)::before {
  left: 0;
  right: 0;
  top: -23px;
  bottom: auto;
  height: 2px;
  width: auto;
}
</style>
