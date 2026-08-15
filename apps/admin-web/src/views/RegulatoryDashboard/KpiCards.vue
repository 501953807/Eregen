<template>
  <div class="kpi-grid">
    <HopeStatCard
      value="overview.total_patients"
      label="在院患者总数"
      icon-color="info"
      gradient="linear-gradient(135deg, #079aa2, #0ea5e9)"
    >
      <template #icon>
        <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg></el-icon>
      </template>
      <template #trend>
        <span class="hope-stat-card__trend hope-stat-card__trend-up">↑ {{ todayAdmissions }} 今日入院</span>
      </template>
    </HopeStatCard>

    <HopeStatCard
      value="overview.wearable_count"
      label="佩戴腕带设备"
      icon-color="success"
      gradient="linear-gradient(135deg, #22c55e, #16a34a)"
    >
      <template #icon>
        <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg></el-icon>
      </template>
      <template #trend>
        <span class="hope-stat-card__trend hope-stat-card__trend-down">↓ {{ offlineDevices }} 离线</span>
      </template>
    </HopeStatCard>

    <HopeStatCard
      value="overview.today_alerts"
      label="今日异常告警"
      icon-color="warning"
      gradient="linear-gradient(135deg, #f59e0b, #d97706)"
    >
      <template #icon>
        <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0"/></svg></el-icon>
      </template>
      <template #trend>
        <span class="hope-stat-card__trend hope-stat-card__trend-down">↑ {{ fenceViolations }} 越界</span>
      </template>
    </HopeStatCard>

    <HopeStatCard
      :value="overview.rule_triggers"
      label="规则引擎触发"
      icon-color="primary"
      gradient="linear-gradient(135deg, #3a57e8, #6f42c1)"
    >
      <template #icon>
        <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z"/></svg></el-icon>
      </template>
      <template #trend>
        <span class="hope-stat-card__trend hope-stat-card__trend-up">自动处理 {{ autoHandleRate }}%</span>
      </template>
    </HopeStatCard>
  </div>
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
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 4px;
}
@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 640px) {
  .kpi-grid { grid-template-columns: 1fr; }
}
</style>
