<template>
  <Teleport to="body">
    <Transition name="hope-modal">
      <div v-if="modelValue" class="hope-modal__overlay" @click.self="handleOverlayClick">
        <div class="hope-modal__container" :class="`hope-modal--${size}`">
          <div class="hope-modal__header">
            <div v-if="$slots.header" class="hope-modal__header-content">
              <slot name="header" />
            </div>
            <h3 v-else class="hope-modal__title">{{ title }}</h3>
            <button v-if="closable" class="hope-modal__close" @click="handleClose">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
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
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: boolean
  title?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  closable?: boolean
}>(), {
  title: '',
  size: 'md',
  closable: true,
})

const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

function handleOverlayClick() {
  if (props.closable) {
    emit('update:modelValue', false)
  }
}

function handleClose() {
  emit('update:modelValue', false)
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.closable) {
    emit('update:modelValue', false)
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
})
</script>

<style scoped>
.hope-modal__overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}
.hope-modal__container {
  background: var(--hope-surface);
  border-radius: var(--hope-radius-lg);
  box-shadow: var(--hope-shadow-lg);
  width: 100%;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.hope-modal--sm { max-width: 400px; }
.hope-modal--md { max-width: 600px; }
.hope-modal--lg { max-width: 800px; }
.hope-modal--xl { max-width: 1000px; }

.hope-modal__header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--hope-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.hope-modal__title {
  font-size: 18px;
  font-weight: 600;
  color: var(--hope-text);
  margin: 0;
}
.hope-modal__close {
  width: 32px;
  height: 32px;
  border-radius: var(--hope-radius-sm);
  border: none;
  background: var(--hope-surface-light);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--hope-text-muted);
  transition: all 0.2s;
}
.hope-modal__close:hover {
  background: var(--hope-danger-light);
  color: var(--hope-danger);
}
.hope-modal__body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}
.hope-modal__footer {
  padding: 16px 24px;
  border-top: 1px solid var(--hope-border);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* Transitions */
.hope-modal-enter-active,
.hope-modal-leave-active {
  transition: opacity 0.2s ease;
}
.hope-modal-enter-from,
.hope-modal-leave-to {
  opacity: 0;
}
.hope-modal-enter-active .hope-modal__container,
.hope-modal-leave-active .hope-modal__container {
  transition: transform 0.2s ease;
}
.hope-modal-enter-from .hope-modal__container,
.hope-modal-leave-to .hope-modal__container {
  transform: scale(0.95);
}
</style>
