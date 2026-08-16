<template>
  <div class="circle-progress">
    <svg :width="size" :height="size" viewBox="0 0 36 36">
      <circle
        cx="18" cy="18" r="15.5"
        fill="none"
        :stroke="bgColor"
        stroke-width="2.5"
      />
      <circle
        cx="18" cy="18" r="15.5"
        fill="none"
        :stroke="color"
        stroke-width="2.5"
        stroke-linecap="round"
        :stroke-dasharray="`${value} ${100 - value}`"
        transform="rotate(-90 18 18)"
        class="circle-progress__ring"
      />
      <text
        x="18" y="20"
        text-anchor="middle"
        dominant-baseline="middle"
        :fill="color"
        font-size="7"
        font-weight="700"
        font-family="Inter, sans-serif"
      >{{ value }}%</text>
    </svg>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  value: number
  color?: string
  bgColor?: string
  size?: number
}>()
</script>

<style scoped>
.circle-progress {
  display: inline-block;
  vertical-align: middle;
}
.circle-progress__ring {
  transition: stroke-dasharray 0.8s cubic-bezier(0.4, 0, 0.2, 1);
}
</style>
