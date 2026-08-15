<template>
  <!-- Rule Engine Status -->
  <HopeCard title="规则引擎状态" style="margin-bottom: 16px;">
    <div class="rule-list">
      <div v-for="rule in ruleStatusList" :key="rule.code" class="rule-item" :class="'risk-' + rule.riskLevel">
        <div class="rule-info">
          <div class="rule-name">{{ rule.name }}</div>
          <div class="rule-desc">{{ rule.desc }}</div>
        </div>
        <span class="rule-trigger" :style="{ color: rule.triggerColor }">{{ rule.triggerText }}</span>
      </div>
    </div>
  </HopeCard>

  <!-- Department Distribution -->
  <HopeCard title="今日科室分布" style="margin-bottom: 16px;">
    <div class="dept-list">
      <div v-for="dept in departmentStats" :key="dept.name" class="dept-item">
        <div class="dept-name">{{ dept.name }}</div>
        <div class="dept-bar-wrap">
          <div class="dept-bar" :style="{ width: dept.barWidth + '%' }"></div>
        </div>
        <div class="dept-count">{{ dept.count }} 人</div>
      </div>
    </div>
  </HopeCard>

  <!-- Rule Config Table -->
  <HopeCard title="规则配置列表" subtitle="管理各项监管规则的开关与阈值">
    <template #header>
      <div class="rule-table-header">
        <div>
          <div class="hope-content-card__title">规则配置列表</div>
          <div class="hope-content-card__subtitle">管理各项监管规则的开关与阈值</div>
        </div>
      </div>
    </template>
    <HopeTable
      :columns="ruleColumns"
      :data="ruleConfigs"
      :loading="loading"
      :compact="true"
      class="rule-table"
    >
      <template #col-code="{ row }">
        <span class="mono">{{ row.code }}</span>
      </template>
      <template #col-name="{ row }">
        <span class="rule-name-cell">{{ row.name }}</span>
      </template>
      <template #col-enabled="{ row }">
        <el-switch v-model="row.enabled" @change="(v: boolean) => $emit('update-rule', row)" />
      </template>
      <template #col-config="{ row }">
        <HopeBtn variant="text" size="sm" @click="$emit('edit-rule', row)">编辑</HopeBtn>
      </template>
    </HopeTable>
  </HopeCard>

  <!-- Rule Edit Dialog -->
  <el-dialog v-model="showEdit" title="编辑规则配置" width="600px" destroy-on-close>
    <el-form :model="editing" label-width="100px">
      <el-form-item label="规则代码"><el-input v-model="editing.code" disabled /></el-form-item>
      <el-form-item label="配置(JSON)">
        <el-input v-model="configJson" type="textarea" :rows="10" />
      </el-form-item>
    </el-form>
    <template #footer>
      <HopeBtn variant="plain" size="sm" @click="showEdit = false">取消</HopeBtn>
      <HopeBtn variant="filled" size="sm" @click="$emit('save-rule')">保存</HopeBtn>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { HopeCard, HopeTable, HopeBtn } from '@/components/hope'
import type { RuleConfig } from '@/api/regulatory'

const props = defineProps<{
  ruleStatusList: any[]
  departmentStats: { name: string; count: number; barWidth: number }[]
  ruleConfigs: RuleConfig[]
  loading: boolean
}>()

const emit = defineEmits<{
  'update-rule': [row: RuleConfig]
  'edit-rule': [row: RuleConfig]
  'save-rule': []
}>()

const showEdit = ref(false)
const editing = ref<Partial<RuleConfig>>({})
const configJson = ref('')

const ruleColumns = [
  { prop: 'code', label: '规则代码', width: '100px' },
  { prop: 'name', label: '规则名称', width: '160px' },
  { prop: 'enabled', label: '启用', width: '80px' },
  { prop: 'config', label: '配置', width: '80px' },
]

function editRuleConfig(row: RuleConfig) {
  editing.value = { ...row }
  configJson.value = JSON.stringify(row.config || {}, null, 2)
  showEdit.value = true
}

defineExpose({ editRuleConfig })
</script>

<style scoped>
.rule-list { display: flex; flex-direction: column; gap: 8px; }
.rule-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 12px 14px; border-radius: var(--hope-radius-md);
  font-size: 13px; transition: background 0.15s;
}
.risk-high { background: rgba(192,50,33,0.06); }
.risk-med  { background: rgba(250,169,56,0.06); }
.risk-low  { background: rgba(26,160,83,0.06); }
.rule-info { display: flex; flex-direction: column; gap: 2px; }
.rule-name { font-weight: 600; color: var(--hope-text); }
.rule-desc { font-size: 12px; color: var(--hope-text-muted); }
.rule-trigger { font-size: 12px; font-weight: 600; white-space: nowrap; }
.dept-list { display: flex; flex-direction: column; gap: 10px; }
.dept-item { display: flex; align-items: center; gap: 10px; font-size: 13px; }
.dept-name { width: 80px; flex-shrink: 0; color: var(--hope-text-secondary); font-weight: 500; }
.dept-bar-wrap { flex: 1; height: 8px; background: var(--hope-bg); border-radius: 4px; overflow: hidden; }
.dept-bar { height: 100%; background: linear-gradient(90deg, var(--hope-primary), var(--hope-accent)); border-radius: 4px; transition: width 0.4s ease; }
.dept-count { width: 60px; text-align: right; color: var(--hope-text-muted); font-size: 12px; }
.mono { font-family: 'SF Mono', 'Consolas', monospace; font-size: 12px; color: var(--hope-text-secondary); }
.rule-name-cell { font-size: 13px; color: var(--hope-text); font-weight: 500; }
.rule-table-header { display: flex; justify-content: space-between; align-items: flex-start; }
</style>
