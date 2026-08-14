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
