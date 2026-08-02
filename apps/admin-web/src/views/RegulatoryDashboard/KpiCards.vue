<template>
  <el-row :gutter="12" style="margin-bottom: 16px;">
    <el-col :span="6">
      <el-card shadow="hover" class="kpi-card kpi-blue">
        <div class="kpi-icon-wrap">🏥</div>
        <div class="kpi-value">{{ overview.total_patients }}</div>
        <div class="kpi-label">在院患者总数</div>
        <div class="kpi-trend trend-up">↑ {{ todayAdmissions }} 今日入院</div>
      </el-card>
    </el-col>
    <el-col :span="6">
      <el-card shadow="hover" class="kpi-card kpi-green">
        <div class="kpi-icon-wrap">💍</div>
        <div class="kpi-value">{{ overview.wearable_count }}</div>
        <div class="kpi-label">佩戴腕带设备</div>
        <div class="kpi-trend trend-down">↓ {{ offlineDevices }} 离线</div>
      </el-card>
    </el-col>
    <el-col :span="6">
      <el-card shadow="hover" class="kpi-card kpi-danger">
        <div class="kpi-icon-wrap">⚠️</div>
        <div class="kpi-value">{{ overview.today_alerts }}</div>
        <div class="kpi-label">今日异常告警</div>
        <div class="kpi-trend trend-down">↑ {{ fenceViolations }} 越界</div>
      </el-card>
    </el-col>
    <el-col :span="6">
      <el-card shadow="hover" class="kpi-card kpi-purple">
        <div class="kpi-icon-wrap">⚙️</div>
        <div class="kpi-value">{{ overview.rule_triggers }}</div>
        <div class="kpi-label">规则引擎触发</div>
        <div class="kpi-trend">自动处理率 {{ autoHandleRate }}%</div>
      </el-card>
    </el-col>
  </el-row>
</template>

<script setup lang="ts">
const props = defineProps<{
  overview: { total_patients: number; wearable_count: number; today_alerts: number; rule_triggers: number }
  todayAdmissions: number
  offlineDevices: number
  fenceViolations: number
  autoHandleRate: number
}>()
</script>

<style scoped>
.kpi-card :deep(.el-card__body) { padding: 18px; display: flex; flex-direction: column; align-items: center; text-align: center; border-radius: 14px; }
.kpi-value { font-size: 28px; font-weight: 800; line-height: 1.2; }
.kpi-label { font-size: 12px; color: var(--el-text-color-secondary); margin-top: 6px; font-weight: 600; }
.kpi-icon-wrap { font-size: 24px; margin-bottom: 4px; }
.kpi-trend { font-size: 11px; margin-top: 4px; }
.trend-up { color: #16A34A; }
.trend-down { color: #EF4444; }
.kpi-blue .kpi-value { color: #2563EB; }
.kpi-green .kpi-value { color: #16A34A; }
.kpi-danger .kpi-value { color: #EF4444; }
.kpi-purple .kpi-value { color: #7C3AED; }
</style>
