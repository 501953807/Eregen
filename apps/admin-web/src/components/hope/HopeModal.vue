<template>
  <div class="hope-modal-overlay" :class="{ open: modelValue }" @click.self="close">
    <div class="hope-modal" :class="[`hope-modal--${size}`, className]">
      <div class="hope-modal__header">
        <div>
          <div class="hope-modal__title">{{ title }}</div>
          <div v-if="subtitle" class="hope-modal__subtitle">{{ subtitle }}</div>
        </div>
        <button class="hope-modal__close" @click="close">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>
      <div class="hope-modal__body">
        <slot />
      </div>
      <div v-if="$slots.footer" class="hope-modal__footer">
        <slot name="footer" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'

const props = defineProps<{
  modelValue: boolean
  title?: string
  subtitle?: string
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  className?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [v: boolean] }>()

watch(() => props.modelValue, (v) => {
  if (v) document.body.style.overflow = 'hidden'
  else document.body.style.overflow = ''
})

function close() {
  emit('update:modelValue', false)
}
</script>
