<template>
  <div class="hope-avatar" :class="`hope-avatar--${size}`" :data-status="status">
    <img v-if="src" :src="src" :alt="name" class="hope-avatar__img" />
    <span v-else class="hope-avatar__text">{{ shortName }}</span>
    <span v-if="status !== 'offline'" class="hope-avatar__status"></span>
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

<style scoped>
.hope-avatar {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--hope-primary-gradient);
  color: white;
  font-weight: 600;
  flex-shrink: 0;
}
.hope-avatar__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}
.hope-avatar__text {
  font-size: inherit;
  line-height: 1;
}
.hope-avatar__status {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 30%;
  height: 30%;
  border-radius: 50%;
  border: 2px solid white;
}

/* Sizes */
.hope-avatar--xs { width: 24px; height: 24px; font-size: 10px; }
.hope-avatar--sm { width: 32px; height: 32px; font-size: 12px; }
.hope-avatar--md { width: 40px; height: 40px; font-size: 14px; }
.hope-avatar--lg { width: 48px; height: 48px; font-size: 16px; }
.hope-avatar--xl { width: 64px; height: 64px; font-size: 20px; }

/* Status indicators */
.hope-avatar[data-status="online"] .hope-avatar__status { background: var(--hope-success); }
.hope-avatar[data-status="busy"] .hope-avatar__status { background: var(--hope-error); }
.hope-avatar[data-status="away"] .hope-avatar__status { background: var(--hope-warning); }
</style>
