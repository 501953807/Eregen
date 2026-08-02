<template>
  <div class="skeleton-wrapper">
    <!-- Loading skeleton bars of varying widths to mimic real content -->
    <div class="skeleton-bar large"></div>
    <div class="skeleton-bar medium"></div>
    <div class="skeleton-bar small"></div>
    <div v-if="showCard" class="skeleton-card"></div>
    <div v-if="showList" class="skeleton-list">
      <div class="skeleton-row"></div>
      <div class="skeleton-row"></div>
      <div class="skeleton-row"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  showCard?: boolean;
  showList?: boolean;
}>()
</script>

<style scoped>
.skeleton-wrapper {
  background: var(--el-bg-color-light);
  border-radius: var(--el-border-radius-md);
  padding: 24px;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

@keyframes skeleton-pulse {
  0%, 100% { opacity: 0.6; }
  50% { opacity: 1; }
}

.skeleton-bar {
  height: 8px;
  background: linear-gradient(90deg, var(--el-border-color) 30%, var(--el-border-color-light) 50%, var(--el-border-color) 70%);
  background-size: 200% 100%;
  animation: skeleton-move 1.5s ease-in-out infinite;
  border-radius: 4px;
  margin-bottom: 16px;
}

.skeleton-bar.large { width: 100%; }
.skeleton-bar.medium { width: 60%; }
.skeleton-bar.small { width: 30%; }

.skeleton-card {
  width: 100%;
  height: 200px;
  background: linear-gradient(90deg, var(--el-border-color) 30%, var(--el-border-color-light) 50%, var(--el-border-color) 70%);
  background-size: 200% 100%;
  animation: skeleton-move 1.5s ease-in-out infinite;
  border-radius: var(--el-border-radius-lg);
  margin-bottom: 20px;
}

.skeleton-list { display: flex; flex-direction: column; gap: 12px; }

.skeleton-row {
  height: 48px;
  background: linear-gradient(90deg, var(--el-border-color) 30%, var(--el-border-color-light) 50%, var(--el-border-color) 70%);
  background-size: 200% 100%;
  animation: skeleton-move 1.5s ease-in-out infinite;
  border-radius: var(--el-border-radius-sm);
  display: flex; align-items: center; padding: 0 16px;
}

.skeleton-row:nth-child(1) { width: 40%; }
.skeleton-row:nth-child(2) { width: 70%; }
.skeleton-row:nth-child(3) { width: 50%; }

@keyframes skeleton-move {
  0% { transform: translateX(-200%); }
  100% { transform: translateX(200%); }
}
</style>