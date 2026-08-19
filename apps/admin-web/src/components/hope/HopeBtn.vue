<template>
  <button
    class="hope-btn"
    :class="[
      `hope-btn--${variant}`,
      `hope-btn--${size}`,
      { 'hope-btn--disabled': disabled, 'hope-btn--icon': iconOnly, 'hope-btn--loading': loading }
    ]"
    :disabled="disabled || loading"
    @click="!disabled && !loading && $emit('click', $event)"
  >
    <span v-if="loading" class="hope-btn__spinner">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
        <path d="M21 12a9 9 0 11-6.219-8.56" stroke-dasharray="50" stroke-dashoffset="10">
          <animate attributeName="stroke-dashoffset" values="50;0" dur="0.6s" repeatCount="indefinite"/>
        </path>
      </svg>
    </span>
    <slot name="icon" v-if="!loading && $slots.icon" />
    <el-icon v-if="!loading && icon && !$slots.icon" :size="iconSize">
      <component :is="icon" />
    </el-icon>
    <span v-if="$slots.default || label" class="hope-btn__label">
      <slot />
    </span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Component } from 'vue'

const props = withDefaults(defineProps<{
  variant?: 'filled' | 'outlined' | 'text' | 'ghost' | 'plain' | 'success' | 'warning' | 'error' | 'info' | 'accent'
  size?: 'sm' | 'md' | 'lg' | 'icon'
  icon?: Component
  iconOnly?: boolean
  label?: string
  disabled?: boolean
  loading?: boolean
}>(), {
  variant: 'filled',
  size: 'md',
  iconOnly: false,
  disabled: false,
  loading: false,
})

const iconSize = computed(() => props.size === 'sm' ? 14 : props.size === 'lg' ? 18 : 16)

defineEmits<{ click: [e: MouseEvent] }>()
</script>

<style scoped>
.hope-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1.5px solid transparent;
  border-radius: var(--hope-radius-sm);
  font-family: inherit;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  outline: none;
  position: relative;
  overflow: hidden;
}

.hope-btn::after {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at var(--ripple-x, 50%) var(--ripple-y, 50%), rgba(255,255,255,0.15) 0%, transparent 60%);
  opacity: 0;
  transition: opacity 0.3s;
}

.hope-btn:hover::after { opacity: 1; }
.hope-btn:focus-visible { box-shadow: var(--hope-shadow-input-focus) !important; }
.hope-btn:active { transform: scale(0.97); }
.hope-btn:disabled { opacity: 0.4; cursor: not-allowed; transform: none; }

/* Variant: filled */
.hope-btn--filled {
  background: var(--hope-primary-gradient);
  color: #fff;
  border-color: transparent;
  box-shadow: var(--hope-shadow-primary);
}
.hope-btn--filled:hover:not(:disabled) {
  box-shadow: var(--hope-shadow-primary-hover);
  transform: translateY(-1px);
}

/* Variant: outlined */
.hope-btn--outlined {
  background: transparent;
  color: var(--hope-primary);
  border-color: var(--hope-primary);
}
.hope-btn--outlined:hover:not(:disabled) {
  background: var(--hope-primary-lighter);
}

/* Variant: text */
.hope-btn--text {
  background: transparent;
  color: var(--hope-primary);
  border-color: transparent;
  padding: 6px 12px;
}
.hope-btn--text:hover:not(:disabled) { background: rgba(var(--hope-primary-rgb), 0.06); }

/* Variant: ghost */
.hope-btn--ghost {
  background: transparent;
  color: var(--hope-text);
  border-color: transparent;
}
.hope-btn--ghost:hover:not(:disabled) { background: var(--hope-bg); color: var(--hope-text); }

/* Variant: plain */
.hope-btn--plain {
  background: var(--hope-bg);
  color: var(--hope-text-secondary);
  border-color: var(--hope-border);
}
.hope-btn--plain:hover:not(:disabled) { background: var(--hope-border); }

/* Semantic variants */
.hope-btn--success { background: var(--hope-success); color: #fff; border-color: transparent; box-shadow: 0 2px 8px rgba(var(--hope-success-rgb), 0.2); }
.hope-btn--success:hover:not(:disabled) { background: #158a45; transform: translateY(-1px); }
.hope-btn--warning { background: var(--hope-warning); color: #000; border-color: transparent; }
.hope-btn--warning:hover:not(:disabled) { background: #e09430; transform: translateY(-1px); }
.hope-btn--error   { background: var(--hope-danger); color: #fff; border-color: transparent; box-shadow: 0 2px 8px rgba(var(--hope-error-rgb), 0.2); }
.hope-btn--error:hover:not(:disabled) { background: #a82a1d; transform: translateY(-1px); }
.hope-btn--info    { background: var(--hope-info); color: #fff; border-color: transparent; }
.hope-btn--info:hover:not(:disabled) { background: #068a90; transform: translateY(-1px); }
.hope-btn--accent  { background: var(--hope-accent); color: #fff; border-color: transparent; }
.hope-btn--accent:hover:not(:disabled) { background: var(--hope-accent-dark); transform: translateY(-1px); }

/* Sizes */
.hope-btn--sm { padding: 6px 14px; font-size: 12px; border-radius: var(--hope-radius-sm); }
.hope-btn--md { padding: 9px 20px; font-size: 14px; }
.hope-btn--lg { padding: 14px 28px; font-size: 16px; border-radius: var(--hope-radius-lg); }
.hope-btn--icon { padding: 10px; border-radius: 50%; min-width: 36px; min-height: 36px; }
.hope-btn--icon.hope-btn--sm { min-width: 32px; min-height: 32px; padding: 6px; border-radius: var(--hope-radius-md); }

/* Loading */
.hope-btn__spinner {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.hope-btn__spinner svg {
  animation: hope-spin 0.8s linear infinite;
}
@keyframes hope-spin { to { transform: rotate(360deg); } }
</style>
