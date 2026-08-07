<template>
  <div class="empty-state" v-if="isShowing">
    <div class="empty-state-icon">{{ icon }}</div>
    <div class="empty-state-title">{{ title }}</div>
    <div class="empty-state-desc">{{ description }}</div>
    <button
      v-if="hasAction"
      class="btn btn-primary"
      @click="onAction"
    >{{ actionText }}</button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'

// Entry animation
const isShowing = ref(false)
let timeout: number | null = null

onMounted(() => {
  timeout = window.setTimeout(() => { isShowing.value = true; }, 100)
})

onBeforeUnmount(() => {
  if (timeout !== null) clearTimeout(timeout)
})

defineProps<{
  title?: string
  description?: string
  icon?: string
  hasAction?: boolean
  actionText?: string
}>()

const emit = defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()

const onAction = (event: MouseEvent) => {
  emit('click', event)
}
</script>

<style scoped>
.empty-state {
  text-align: center;
  padding: 60px 24px;
  color: var(--el-text-color-secondary);
  background: white;
  border-radius: var(--el-border-radius-md);
  box-shadow: var(--el-shadow-sm);
  margin: 24px 0;
}

.empty-state-icon {
  font-size: 64px;
  margin-bottom: 20px;
  opacity: 0.7;
  animation: icon-fade-in 0.8s ease-out forwards;
}

.empty-state-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 12px;
  animation: title-fade-in 0.9s ease-out forwards;
}

.empty-state-desc {
  font-size: 15px;
  max-width: 400px;
  margin: 0 auto 28px;
  line-height: 1.6;
  animation: desc-fade-in 1s ease-out forwards;
}

.btn {
  display: inline-flex; align-items: center; justify-content: center;
  border-radius: var(--el-border-radius-md); font-weight: 600; cursor: pointer;
  border: none; transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  padding: 12px 28px; font-size: 15px;
}

.btn-primary {
  background: var(--el-color-primary); color: white;
}

.btn-primary:hover {
  background: var(--el-color-primary-dark);
  transform: translateY(-2px);
  box-shadow: var(--el-shadow-lg);
}

@keyframes icon-fade-in { from { opacity: 0; transform: scale(0.8) rotate(-10deg); } to { opacity: 0.7; transform: scale(1) rotate(0); } }
@keyframes title-fade-in { from { opacity: 0; transform: translateY(15px); } to { opacity: 1; transform: translateY(0); } }
@keyframes desc-fade-in { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
</style>