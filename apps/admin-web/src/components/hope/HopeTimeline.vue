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
