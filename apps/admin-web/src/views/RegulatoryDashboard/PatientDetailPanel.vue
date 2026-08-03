<template>
  <div>
    <div class="side-panel-overlay" :class="{ show: visible }" @click="$emit('close')" />
    <div class="side-panel" :class="{ open: visible }">
      <div class="panel-header">
        <span class="panel-title">患者详情 — {{ patient?.name || '' }}</span>
        <button class="panel-close" @click="$emit('close')">&#10005;</button>
      </div>
      <div class="panel-body" v-if="patient">
        <div class="patient-detail-header">
          <div class="patient-avatar-large avatar-blue">{{ patient.name?.[0] || '?' }}</div>
          <div>
            <div class="patient-detail-name">{{ patient.name }}</div>
            <div class="patient-detail-id">住院号: <span class="mono">{{ patient.admission_no }}</span></div>
          </div>
        </div>

        <div class="info-section">
          <div class="section-title">基本信息</div>
          <div class="panel-row"><span class="panel-label">科室</span><span class="panel-value">{{ patient.department }}</span></div>
          <div class="panel-row"><span class="panel-label">床号</span><span class="panel-value">{{ patient.bed_number }}</span></div>
          <div class="panel-row"><span class="panel-label">最后核验</span><span class="panel-value">{{ patient.last_verify || '未核验' }}</span></div>
          <div class="panel-row"><span class="panel-label">围栏状态</span>
            <span class="panel-value">
              <span class="status-badge" :class="patient.fence_status === 'inside' ? 'badge-success' : 'badge-danger'">
                <span class="status-dot" :class="patient.fence_status === 'inside' ? 'dot-success' : 'dot-danger'"></span>
                {{ patient.fence_status === 'inside' ? '在院内' : '已越界' }}
              </span>
            </span>
          </div>
        </div>

        <div class="info-section" v-if="patient.alert_tags?.length">
          <div class="section-title">告警标签</div>
          <div class="alert-tags-wrap">
            <el-tag v-for="tag in patient.alert_tags" :key="tag" size="small" type="warning" effect="light" style="margin: 2px 4px 2px 0;">{{ tag }}</el-tag>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{ modelValue: boolean; patient: any }>()
const emit = defineEmits<{ 'update:modelValue': [v: boolean] }>()

const visible = ref(false)
watch(() => props.modelValue, v => { visible.value = v })
watch(visible, v => { if (!v) emit('update:modelValue', false) })
</script>

<style scoped>
.side-panel-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 200; display: none; }
.side-panel-overlay.show { display: block; }
.side-panel { position: fixed; top: 0; right: -520px; bottom: 0; width: 520px; background: white; z-index: 201; transition: right 0.3s ease; overflow-y: auto; box-shadow: -10px 0 40px rgba(0,0,0,0.1); }
.side-panel.open { right: 0; }
.panel-header { padding: 20px 24px; border-bottom: 1px solid var(--el-border-color-light); display: flex; align-items: center; justify-content: space-between; position: sticky; top: 0; background: white; z-index: 1; }
.panel-title { font-size: 15px; font-weight: 700; }
.panel-close { width: 32px; height: 32px; border-radius: 8px; border: none; background: var(--el-fill-color-light); cursor: pointer; font-size: 18px; display: flex; align-items: center; justify-content: center; transition: background 0.15s; }
.panel-close:hover { background: var(--el-border-color-light); }
.panel-body { padding: 20px 24px; }
.patient-detail-header { display: flex; align-items: center; gap: 14px; margin-bottom: 20px; padding-bottom: 16px; border-bottom: 1px solid var(--el-border-color-light); }
.patient-avatar-large { width: 48px; height: 48px; border-radius: 50%; background: #DBEAFE; color: #165DFF; display: flex; align-items: center; justify-content: center; font-size: 20px; font-weight: 700; flex-shrink: 0; }
.patient-detail-name { font-size: 17px; font-weight: 700; color: var(--el-text-color-primary); }
.patient-detail-id { font-size: 12px; color: var(--el-text-color-secondary); margin-top: 2px; }
.info-section { margin-bottom: 20px; }
.section-title { font-size: 13px; font-weight: 700; color: var(--el-text-color-regular); text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 10px; padding-bottom: 6px; border-bottom: 1px solid var(--el-border-color-light); }
.panel-row { display: flex; justify-content: space-between; align-items: center; padding: 6px 0; }
.panel-label { font-size: 13px; color: var(--el-text-color-secondary); font-weight: 500; }
.panel-value { font-size: 13px; color: var(--el-text-color-primary); font-weight: 600; }
.status-badge { display: inline-flex; align-items: center; gap: 5px; padding: 3px 10px; border-radius: 8px; font-size: 12px; font-weight: 600; }
.badge-success { background: #F0FDF4; color: #16A34A; }
.badge-danger { background: #FEF2F2; color: #DC2626; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; }
.dot-success { background: #16A34A; }
.dot-danger { background: #DC2626; }
.mono { font-family: 'SF Mono', 'Consolas', monospace; font-size: 12px; }
</style>
