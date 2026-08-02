<template>
  <el-row :gutter="16" style="margin-bottom: 16px;">
    <el-col :span="8">
      <el-date-picker v-model="dateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" style="width: 100%;" />
    </el-col>
    <el-col :span="4">
      <el-button type="primary" @click="$emit('generate')">生成报表</el-button>
    </el-col>
  </el-row>
  <div v-if="report">
    <el-descriptions title="总体概览" :column="3" border class="report-desc">
      <el-descriptions-item label="期间患者总数">{{ report.summary.total_patients_period }}</el-descriptions-item>
      <el-descriptions-item label="平均住院天数">{{ report.summary.avg_stay_days }}</el-descriptions-item>
      <el-descriptions-item label="合规率">{{ report.summary.compliance_rate }}%</el-descriptions-item>
      <el-descriptions-item label="围栏违规">{{ report.summary.fence_violations }}</el-descriptions-item>
      <el-descriptions-item label="未核验告警">{{ report.summary.no_verify_alerts }}</el-descriptions-item>
      <el-descriptions-item label="费用异常">{{ report.summary.expense_anomalies }}</el-descriptions-item>
    </el-descriptions>
    <h4 style="margin-top: 20px;">科室合规率</h4>
    <el-table :data="report.department_breakdown" stripe>
      <el-table-column prop="name" label="科室" width="150" />
      <el-table-column prop="total_patients" label="患者数" width="100" />
      <el-table-column prop="alerts" label="告警数" width="100" />
      <el-table-column prop="compliance_rate" label="合规率" width="100">
        <template #default="{ row }">
          <el-progress :percentage="row.compliance_rate" :format="() => row.compliance_rate + '%'" :stroke-width="10" />
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ComplianceReport } from '@/api/regulatory'

defineProps<{ report: ComplianceReport | null }>()
defineEmits<{ generate: [] }>()

const dateRange = ref<[Date, Date] | null>(null)
</script>
