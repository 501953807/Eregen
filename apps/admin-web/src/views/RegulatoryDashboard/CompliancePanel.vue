<template>
  <div>
    <div class="compliance-header">
      <el-row :gutter="12" style="margin-bottom: 0;">
        <el-col :span="10">
          <el-date-picker v-model="dateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" style="width: 100%;" />
        </el-col>
        <el-col :span="4">
          <HopeBtn variant="filled" size="sm" @click="$emit('generate')">生成报表</HopeBtn>
        </el-col>
      </el-row>
    </div>

    <div v-if="report" class="report-body">
      <!-- Summary Cards -->
      <div class="summary-grid">
        <div class="summary-item">
          <div class="summary-value">{{ report.summary.total_patients_period }}</div>
          <div class="summary-label">期间患者总数</div>
        </div>
        <div class="summary-item">
          <div class="summary-value">{{ report.summary.avg_stay_days }}</div>
          <div class="summary-label">平均住院天数</div>
        </div>
        <div class="summary-item">
          <div class="summary-value" :class="{ 'is-high': report.summary.compliance_rate >= 95, 'is-mid': report.summary.compliance_rate >= 80 }">
            {{ report.summary.compliance_rate }}%
          </div>
          <div class="summary-label">合规率</div>
        </div>
        <div class="summary-item">
          <div class="summary-value" :class="{ 'is-danger': report.summary.fence_violations > 0 }">{{ report.summary.fence_violations }}</div>
          <div class="summary-label">围栏违规</div>
        </div>
        <div class="summary-item">
          <div class="summary-value" :class="{ 'is-danger': report.summary.no_verify_alerts > 0 }">{{ report.summary.no_verify_alerts }}</div>
          <div class="summary-label">未核验告警</div>
        </div>
        <div class="summary-item">
          <div class="summary-value" :class="{ 'is-danger': report.summary.expense_anomalies > 0 }">{{ report.summary.expense_anomalies }}</div>
          <div class="summary-label">费用异常</div>
        </div>
      </div>

      <!-- Department Breakdown -->
      <div class="dept-breakdown-title">科室合规率</div>
      <HopeTable
        :columns="deptColumns"
        :data="report.department_breakdown"
        :compact="true"
        class="dept-table"
      >
        <template #col-compliance_rate="{ row }">
          <div class="progress-wrap">
            <div class="progress-bar" :class="{ 'progress-ok': row.compliance_rate >= 95, 'progress-warn': row.compliance_rate >= 80, 'progress-bad': row.compliance_rate < 80 }" :style="{ width: row.compliance_rate + '%' }"></div>
            <span class="progress-text">{{ row.compliance_rate }}%</span>
          </div>
        </template>
      </HopeTable>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { HopeBtn, HopeTable } from '@/components/hope'
import type { ComplianceReport } from '@/api/regulatory'

defineProps<{ report: ComplianceReport | null }>()
defineEmits<{ generate: [] }>()

const dateRange = ref<[Date, Date] | null>(null)

const deptColumns = [
  { prop: 'name', label: '科室' },
  { prop: 'total_patients', label: '患者数' },
  { prop: 'alerts', label: '告警数' },
  { prop: 'compliance_rate', label: '合规率' },
]
</script>

<style scoped>
.compliance-header { margin-bottom: 16px; }
.report-body { padding-top: 4px; }
.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 20px;
}
.summary-item {
  background: var(--hope-bg);
  border-radius: var(--hope-radius-md);
  padding: 14px 16px;
  text-align: center;
  border: 1px solid var(--hope-border);
}
.summary-value {
  font-size: 26px;
  font-weight: 800;
  color: var(--hope-text);
  letter-spacing: -0.02em;
  line-height: 1.1;
}
.summary-value.is-high  { color: #1aa053; }
.summary-value.is-mid   { color: #FAA938; }
.summary-value.is-danger { color: #c03221; }
.summary-label {
  font-size: 12px;
  color: var(--hope-text-muted);
  margin-top: 4px;
  font-weight: 500;
}
.dept-breakdown-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--hope-text);
  margin-bottom: 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--hope-border);
}
.progress-wrap {
  position: relative;
  height: 8px;
  background: var(--hope-border);
  border-radius: 4px;
  overflow: hidden;
  display: flex;
  align-items: center;
}
.progress-bar {
  height: 100%;
  border-radius: 4px;
  transition: width 0.4s ease;
}
.progress-ok   { background: #1aa053; }
.progress-warn { background: #FAA938; }
.progress-bad  { background: #c03221; }
.progress-text {
  position: absolute;
  right: 6px;
  font-size: 11px;
  font-weight: 600;
  color: var(--hope-text-secondary);
}
@media (max-width: 768px) {
  .summary-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 480px) {
  .summary-grid { grid-template-columns: 1fr; }
}
</style>
