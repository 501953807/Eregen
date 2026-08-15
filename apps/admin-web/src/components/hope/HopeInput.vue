<template>
  <div class="hope-input-wrap" :class="{ 'hope-input-wrap--error': error, 'hope-input-wrap--success': success, 'hope-input-wrap--focused': focused }">
    <div v-if="$slots.prefix" class="hope-input-prefix">
      <slot name="prefix" />
    </div>
    <input
      ref="inputRef"
      class="hope-input"
      :class="[`hope-input--${size}`, { 'hope-input--password': type === 'password' }]"
      :type="showPasswordVisible ? 'text' : type"
      :placeholder="placeholder"
      :disabled="disabled"
      :value="modelValue"
      :maxlength="maxlength"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      @focus="focused = true"
      @blur="focused = false"
    />
    <div v-if="$slots.suffix || showPasswordToggle" class="hope-input-suffix">
      <button v-if="showPasswordToggle && type === 'password'" type="button" class="hope-input-pw-toggle" @click="showPasswordVisible = !showPasswordVisible">
        <svg v-if="!showPasswordVisible" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
        <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
      </button>
      <slot name="suffix" />
    </div>
    <div v-if="error" class="hope-input-msg hope-input-msg--error">{{ error }}</div>
    <div v-else-if="success" class="hope-input-msg hope-input-msg--success">{{ success }}</div>
    <div v-else-if="helper" class="hope-input-msg hope-input-msg--helper">{{ helper }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
  size?: 'sm' | 'md' | 'lg'
  type?: 'text' | 'password' | 'email' | 'tel' | 'number'
  error?: string
  success?: string
  helper?: string
  disabled?: boolean
  maxlength?: number | string
  showPassword?: boolean
}>(), {
  modelValue: '',
  size: 'md',
  type: 'text',
  error: '',
  success: '',
  helper: '',
  disabled: false,
  showPassword: false,
})

defineEmits<{ 'update:modelValue': [v: string] }>()

const focused = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)
const showPasswordVisible = ref(props.showPassword)

const showPasswordToggle = computed(() => props.type === 'password')
</script>

<style scoped>
.hope-input-wrap {
  position: relative;
  width: 100%;
}

.hope-input {
  width: 100%;
  height: 42px;
  padding: 0 14px;
  border: 1.5px solid var(--hope-border);
  border-radius: var(--hope-radius-md);
  font-size: 14px;
  font-family: inherit;
  color: var(--hope-text);
  background: var(--hope-surface);
  box-shadow: var(--hope-shadow-sm);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  outline: none;
  box-sizing: border-box;
}

.hope-input-wrap--focused .hope-input {
  border-color: var(--hope-primary);
  box-shadow: var(--hope-shadow-input-focus);
}

.hope-input-wrap--error .hope-input {
  border-color: var(--hope-error);
  box-shadow: 0 0 0 3px rgba(192,74,66,0.1), var(--hope-shadow-sm);
}

.hope-input-wrap--success .hope-input {
  border-color: var(--hope-success);
  box-shadow: 0 0 0 3px rgba(86,202,0,0.1), var(--hope-shadow-sm);
}

.hope-input::placeholder { color: var(--hope-text-muted); }
.hope-input:disabled { background: var(--hope-bg); color: var(--hope-text-muted); cursor: not-allowed; }

.hope-input--sm { height: 34px; font-size: 13px; border-radius: var(--hope-radius-sm); }
.hope-input--lg { height: 48px; font-size: 15px; border-radius: var(--hope-radius-lg); }

.hope-input-prefix,
.hope-input-suffix {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  color: var(--hope-text-muted);
  pointer-events: none;
}

.hope-input-prefix { left: 12px; }
.hope-input-suffix { right: 12px; pointer-events: auto; }

.hope-input-pw-toggle {
  background: none;
  border: none;
  cursor: pointer;
  padding: 2px;
  color: var(--hope-text-muted);
  display: flex;
  transition: color 0.2s;
}
.hope-input-pw-toggle:hover { color: var(--hope-primary); }

.hope-input-msg {
  font-size: 12px;
  margin-top: 4px;
  padding-left: 2px;
}
.hope-input-msg--error { color: var(--hope-error); }
.hope-input-msg--success { color: var(--hope-success); }
.hope-input-msg--helper { color: var(--hope-text-muted); }
</style>
