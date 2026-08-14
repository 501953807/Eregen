<template>
  <div ref="fieldRef" class="hope-field" :class="{ focused: focused, 'has-value': hasValue }">
    <input
      ref="inputRef"
      class="hope-input"
      :class="[`hope-input--${size}`, { 'hope-input--error': error }]"
      :placeholder="placeholder"
      :disabled="disabled"
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      @focus="focused = true"
      @blur="focused = false; hasValue = !!($event.target as HTMLInputElement).value"
    />
    <label class="hope-label">{{ label }}</label>
    <span v-if="error" class="hope-error-text">{{ error }}</span>
    <span v-else-if="helper" class="hope-helper">{{ helper }}</span>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'

const props = withDefaults(defineProps<{
  modelValue?: string
  label: string
  placeholder?: string
  size?: 'sm' | 'md' | 'lg'
  error?: string
  helper?: string
  disabled?: boolean
}>(), {
  modelValue: '',
  size: 'md',
  placeholder: '',
  error: '',
  helper: '',
  disabled: false,
})

defineEmits<{ 'update:modelValue': [v: string] }>()

const focused = ref(false)
const hasValue = ref(false)
const fieldRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)

watch(() => props.modelValue, (v) => { hasValue.value = !!v })
</script>
