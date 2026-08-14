<template>
  <div class="hope-chart-wrapper">
    <div class="hope-chart-wrapper__header">
      <div>
        <div class="hope-chart-wrapper__title">{{ title }}</div>
        <div v-if="subtitle" class="hope-chart-wrapper__subtitle">{{ subtitle }}</div>
      </div>
      <div class="hope-chart-wrapper__controls">
        <slot name="controls" />
      </div>
    </div>
    <div ref="chartRef" class="hope-chart" :class="`hope-chart--${size}`"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps<{
  option: echarts.EChartsOption
  title?: string
  subtitle?: string
  size?: 'sm' | 'md' | 'lg'
}>()

const chartRef = ref<HTMLElement | null>(null)
let chart: echarts.ECharts | null = null

function initChart() {
  if (!chartRef.value) return
  chart = echarts.init(chartRef.value)
  chart.setOption(props.option)
}

watch(() => props.option, (newOpt) => {
  chart?.setOption(newOpt, true)
}, { deep: true })

onMounted(() => {
  initChart()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chart?.dispose()
})

function handleResize() {
  chart?.resize()
}
</script>
