<template>
  <button
    class="hope-btn"
    :class="[
      `hope-btn--${variant}`,
      `hope-btn--${size}`,
      { 'hope-btn--disabled': disabled, 'hope-btn--icon': iconOnly }
    ]"
    :disabled="disabled"
    :style="{ width: iconOnly && size === 'sm' ? undefined : undefined }"
    @click="!disabled && $emit('click', $event)"
  >
    <slot name="icon" v-if="$slots.icon" />
    <el-icon v-if="icon && !$slots.icon" :size="size === 'sm' ? 14 : size === 'lg' ? 18 : 16">
      <component :is="icon" />
    </el-icon>
    <span v-if="$slots.default || label">{{ label || $slots.default?.() }}</span>
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
}>(), {
  variant: 'filled',
  size: 'md',
  iconOnly: false,
  disabled: false,
})

defineEmits<{ click: [e: MouseEvent] }>()
</script>
