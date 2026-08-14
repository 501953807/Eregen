<template>
  <div class="hope-tabs" :class="{ 'hope-tabs--pills': pillStyle }">
    <div class="hope-tabs__list" ref="listRef">
      <div v-if="animated" class="hope-tabs__indicator" ref="indicatorRef"></div>
      <button
        v-for="(tab, i) in tabs"
        :key="tab.value"
        class="hope-tab"
        :class="{ active: modelValue === tab.value }"
        :disabled="tab.disabled"
        @click="!tab.disabled && selectTab(tab.value)"
      >
        {{ tab.label }}
        <span v-if="tab.badge" class="hope-tab__badge">{{ tab.badge }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

interface TabItem {
  label: string
  value: string | number
  badge?: number
  disabled?: boolean
}

const props = defineProps<{
  modelValue: string | number
  tabs: TabItem[]
  animated?: boolean
  pillStyle?: boolean
}>()

const emit = defineEmits<{ 'update:modelValue': [v: string | number] }>()
const listRef = ref<HTMLElement | null>(null)
const indicatorRef = ref<HTMLElement | null>(null)

function selectTab(value: string | number) {
  emit('update:modelValue', value)
  if (props.animated && listRef.value && indicatorRef.value) {
    const activeEl = listRef.value.querySelector('.hope-tab.active') as HTMLElement
    if (activeEl) {
      indicatorRef.value.style.left = activeEl.offsetLeft + 'px'
      indicatorRef.value.style.width = activeEl.offsetWidth + 'px'
    }
  }
}

onMounted(() => selectTab(props.modelValue))
watch(() => props.modelValue, () => selectTab(props.modelValue))
</script>
