<template>
  <div class="hope-avatar" :class="`hope-avatar--${size}`" :data-status="status">
    <img v-if="src" :src="src" :alt="name" />
    <span v-else>{{ shortName }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  name?: string
  src?: string
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  status?: 'online' | 'busy' | 'away' | 'offline'
}>(), {
  name: '',
  src: '',
  size: 'md',
  status: 'offline',
})

const shortName = computed(() => {
  if (props.src) return ''
  return props.name ? props.name.charAt(0).toUpperCase() : '?'
})
</script>
