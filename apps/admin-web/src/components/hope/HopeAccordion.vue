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
