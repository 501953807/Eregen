<template>
  <div class="hope-skeleton" :class="`hope-skeleton--${type}`" :style="styleObj"></div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  type?: 'text' | 'title' | 'circle' | 'block'
  width?: string
  height?: string
}>(), {
  type: 'text',
  width: '',
  height: '',
})

const styleObj = computed(() => {
  const s: Record<string, string> = {}
  if (props.width) s.width = props.width
  if (props.height) s.height = props.height
  return s
})
</script>

<style scoped>
.hope-skeleton {
  background: linear-gradient(90deg, var(--hope-border) 25%, var(--hope-surface-light) 50%, var(--hope-border) 75%);
  background-size: 200% 100%;
  border-radius: var(--hope-radius-sm);
  animation: skeleton-loading 1.5s ease-in-out infinite;
}
@keyframes skeleton-loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
.hope-skeleton--text { height: 14px; width: 100%; }
.hope-skeleton--title { height: 20px; width: 60%; }
.hope-skeleton--circle { border-radius: 50%; width: 40px; height: 40px; }
.hope-skeleton--block { width: 100%; height: 120px; border-radius: var(--hope-radius-md); }
</style>
