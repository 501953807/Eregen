<template>
  <div>
    <!-- Rule Engine Status -->
    <el-card shadow="never" class="content-panel rule-panel">
      <template #header>
        <div class="panel-header">
          <span class="panel-title">规则引擎状态</span>
        </div>
      </template>
      <div class="rule-list">
        <div v-for="rule in ruleStatusList" :key="rule.code" class="rule-item" :class="'risk-' + rule.riskLevel">
          <div>
            <div class="rule-name">{{ rule.name }}</div>
            <div class="rule-desc">{{ rule.desc }}</div>
          </div>
          <span class="rule-trigger-count" :style="{ color: rule.triggerColor }">{{ rule.triggerText }}</span>
        </div>
      </div>
    </el-card>

    <!-- Department Distribution -->
    <el-card shadow="never" class="content-panel dept-panel" style="margin-top: 20px;">
      <template #header>
        <div class="panel-header">
          <span class="panel-title">今日科室分布</span>
        </div>
      </template>
      <div class="dept-list">
        <div v-for="dept in departmentStats" :key="dept.name" class="dept-item">
          <div class="dept-name">{{ dept.name }}</div>
          <div class="dept-bar-wrap"><div class="dept-bar" :style="{ width: dept.barWidth + '%' }"></div></div>
          <div class="dept-count">{{ dept.count }} 人</div>
        </div>
      </div>
    </el-card>

    <!-- Rule Config Table -->
    <el-table :data="ruleConfigs" v-loading="loading" stripe style="margin-top: 20px;">
      <el-table-column prop="code" label="规则代码" width="100"><template #default="{ row }"><span class="mono">{{ row.code }}</span></template></el-table-column>
      <el-table-column prop="name" label="规则名称" width="160" />
      <el-table-column prop="enabled" label="启用" width="80">
        <template #default="{ row }"><el-switch v-model="row.enabled" @change="v => $emit('update-rule', row)" /></template>
      </el-table-column>
      <el-table-column prop="config" label="配置">
        <template #default="{ row }"><el-button size="small" @click="$emit('edit-rule', row)">编辑</el-button></template>
      </el-table-column>
    </el-table>

    <!-- Rule Edit Dialog -->
    <el-dialog v-model="showEdit" title="编辑规则配置" width="600px" destroy-on-close>
      <el-form :model="editing" label-width="100px">
        <el-form-item label="规则代码"><el-input v-model="editing.code" disabled /></el-form-item>
        <el-form-item label="配置(JSON)">
          <el-input v-model="configJson" type="textarea" :rows="10" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEdit = false">取消</el-button>
        <el-button type="primary" @click="$emit('save-rule')">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { RuleConfig } from '@/api/regulatory'

defineProps<{
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

function editRuleConfig(row: RuleConfig) {
  editing.value = { ...row }
  configJson.value = JSON.stringify(row.config || {}, null, 2)
  showEdit.value = true
}

defineExpose({ editRuleConfig })
</script>

<style scoped>
.rule-list { display: flex; flex-direction: column; gap: 8px; }
.rule-item { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; border-radius: 6px; font-size: 13px; }
.risk-high { background: #FEF2F2; }
.risk-med { background: #FFFBEB; }
.risk-low { background: #F0FDF4; }
.rule-name { font-weight: 600; }
.rule-desc { font-size: 11px; color: var(--el-text-color-secondary); margin-top: 2px; }
.rule-trigger-count { font-size: 12px; font-weight: 600; }
.dept-list { display: flex; flex-direction: column; gap: 8px; }
.dept-item { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.dept-name { width: 80px; flex-shrink: 0; }
.dept-bar-wrap { flex: 1; height: 10px; background: var(--el-fill-color-lighter); border-radius: 3px; overflow: hidden; }
.dept-bar { height: 100%; background: linear-gradient(90deg, #165DFF, #9B8ED8); border-radius: 3px; transition: width 0.3s; }
.dept-count { width: 60px; text-align: right; color: var(--el-text-color-secondary); font-size: 12px; }
.mono { font-family: 'SF Mono', 'Consolas', monospace; font-size: 12px; }
.panel-header { display: flex; justify-content: space-between; align-items: center; }
.panel-title { font-size: 15px; font-weight: 700; color: var(--el-text-color-primary); border-left: 3px solid #165DFF; padding-left: 8px; }
</style>
