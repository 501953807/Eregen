<template>
  <div>
    <div class="side-panel-overlay" :class="{ show: visible }" @click="$emit('close')" />
    <div class="side-panel" :class="{ open: visible }">
      <div class="panel-header">
        <span class="panel-title">老人档案详情</span>
        <button class="panel-close" @click="$emit('close')">&#10005;</button>
      </div>
      <div class="panel-body" v-if="row">
        <div class="patient-detail-header">
          <div class="patient-avatar-large" :class="row.gender === 1 ? 'avatar-blue' : 'avatar-pink'">{{ row.name?.[0] || '?' }}</div>
          <div>
            <div class="patient-detail-name">{{ row.name }}</div>
            <div class="patient-detail-id">
              <span class="mono">{{ row.id_card }}</span>
              <span class="status-badge" :class="row.status === 'active' ? 'badge-success' : 'badge-gray'" style="margin-left: 8px;">
                <span class="status-dot" :class="row.status === 'active' ? 'dot-success' : 'dot-gray'"></span>
                {{ row.status === 'active' ? '正常' : '停用' }}
              </span>
            </div>
          </div>
        </div>

        <div class="info-section">
          <div class="section-title">基本信息</div>
          <div class="panel-row"><span class="panel-label">性别</span><span class="panel-value">{{ row.gender === 1 ? '男' : row.gender === 2 ? '女' : '-' }}</span></div>
          <div class="panel-row"><span class="panel-label">年龄</span><span class="panel-value">{{ row.age || '—' }} 岁</span></div>
          <div class="panel-row"><span class="panel-label">地址</span><span class="panel-value">{{ row.address || '—' }}</span></div>
          <div class="panel-row"><span class="panel-label">紧急联系人</span><span class="panel-value">{{ row.emergency_contact || '—' }}</span></div>
        </div>

        <div class="info-section">
          <div class="section-title">福利标签</div>
          <div v-loading="loading">
            <div v-for="tag in tags" :key="tag.tag_code" class="welfare-tag-row">
              <el-tag :type="welfareTagType(tag.tag_code)" size="small" effect="light">{{ tag.tag_name }}</el-tag>
              <span class="welfare-tag-issuer">{{ tag.issuer }}</span>
              <span class="welfare-tag-dates">{{ tag.valid_from }} ~ {{ tag.valid_to }}</span>
              <span class="welfare-tag-status" :class="{ expired: !isTagValid(tag) }">{{ isTagValid(tag) ? '有效' : '过期' }}</span>
            </div>
            <div v-if="!tags.length" style="color:var(--el-text-color-placeholder);font-size:13px;">暂无福利标签</div>
          </div>
        </div>

        <div class="info-section">
          <div class="section-title">最近签到记录</div>
          <div v-for="rec in history" :key="rec.signin_time" class="signin-record">
            <div class="signin-time">{{ rec.signin_time }}</div>
            <div class="signin-meta">
              <span class="signin-hospital">{{ rec.hospital_id }}</span>
              <span class="status-badge" :class="rec.is_welfare_signin ? 'badge-success' : 'badge-primary'" style="font-size:11px;padding:1px 6px;">
                {{ rec.is_welfare_signin ? '福利' : '医保' }}
              </span>
            </div>
            <div class="signin-tags" v-if="rec.activated_tags">
              <el-tag v-for="t in parseTags(rec.activated_tags)" :key="t" size="small" style="margin: 2px 4px 2px 0;">{{ t }}</el-tag>
            </div>
          </div>
          <div v-if="!history.length" style="color:var(--el-text-color-placeholder);font-size:13px;">暂无签到记录</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import dayjs from 'dayjs'
import type { CommunityElder } from '@/api/community'

const props = defineProps<{ modelValue: boolean; row: CommunityElder | null }>()
const emit = defineEmits<{ 'update:modelValue': [v: boolean] }>()

const visible = ref(false)
const tags = ref<any[]>([])
const history = ref<any[]>([])
const loading = ref(false)

watch(() => props.modelValue, v => {
  visible.value = v
  if (v && props.row) loadDetail(props.row)
})
watch(visible, v => { if (!v) emit('update:modelValue', false) })

async function loadDetail(row: CommunityElder) {
  loading.value = true
  try {
    tags.value = [
      { tag_code: 'orphan', tag_name: '孤寡老人', issuer: '民政局', valid_from: '2025-01-01', valid_to: '2028-12-31' },
      { tag_code: 'poverty_level_1', tag_name: '特困一级', issuer: '民政局', valid_from: '2025-01-01', valid_to: '2028-12-31' },
      { tag_code: 'disability_2', tag_name: '残疾二级', issuer: '残联', valid_from: '2024-06-01', valid_to: '2027-05-31' },
    ]
    history.value = [
      { signin_time: '2026-07-23 10:30', hospital_id: '社区医院 A', is_welfare_signin: true, activated_tags: '["孤寡","特困一级","残疾二级"]' },
      { signin_time: '2026-07-16 09:15', hospital_id: '社区医院 A', is_welfare_signin: true, activated_tags: '["孤寡","特困一级","残疾二级"]' },
    ]
  } finally { loading.value = false }
}

function welfareTagType(code: string): string {
  const map: Record<string, string> = {
    orphan: 'danger', poverty_level_1: 'warning', poverty_level_2: 'warning',
    disability_1: 'primary', disability_2: 'primary', disability_3: 'primary',
    special_disease: 'danger', bus_discount: 'success', medical_assistance: '',
  }
  return map[code] || ''
}

function isTagValid(tag: any): boolean {
  return tag.valid_to && dayjs(tag.valid_to).isAfter(dayjs())
}

function parseTags(json: string | undefined): string[] {
  if (!json) return []
  try { return JSON.parse(json) } catch { return [] }
}
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
.patient-avatar-large { width: 48px; height: 48px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 20px; font-weight: 700; flex-shrink: 0; }
.avatar-blue { background: #DBEAFE; color: #2563EB; }
.avatar-pink { background: #FCE7F3; color: #EC4899; }
.patient-detail-name { font-size: 17px; font-weight: 700; color: var(--el-text-color-primary); }
.patient-detail-id { font-size: 12px; color: var(--el-text-color-secondary); margin-top: 2px; }
.info-section { margin-bottom: 20px; }
.section-title { font-size: 13px; font-weight: 700; color: var(--el-text-color-regular); text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 10px; padding-bottom: 6px; border-bottom: 1px solid var(--el-border-color-light); }
.panel-row { display: flex; justify-content: space-between; align-items: center; padding: 6px 0; }
.panel-label { font-size: 13px; color: var(--el-text-color-secondary); font-weight: 500; }
.panel-value { font-size: 13px; color: var(--el-text-color-primary); font-weight: 600; }
.welfare-tag-row { display: flex; align-items: center; gap: 8px; padding: 4px 0; font-size: 13px; }
.welfare-tag-issuer { color: var(--el-text-color-secondary); font-size: 12px; }
.welfare-tag-dates { color: var(--el-text-color-secondary); font-size: 12px; margin-left: auto; }
.welfare-tag-status { font-size: 11px; font-weight: 600; padding: 1px 6px; border-radius: 4px; background: #F0FDF4; color: #16A34A; }
.welfare-tag-status.expired { background: #FEF2F2; color: #DC2626; }
.signin-record { padding: 8px 0; border-bottom: 1px solid var(--el-border-color-light); }
.signin-record:last-child { border-bottom: none; }
.signin-time { font-size: 14px; font-weight: 600; color: var(--el-text-color-primary); }
.signin-meta { display: flex; align-items: center; gap: 8px; margin-top: 4px; font-size: 12px; color: var(--el-text-color-secondary); }
.signin-hospital { font-weight: 600; }
.signin-tags { margin-top: 4px; }
.status-badge { display: inline-flex; align-items: center; gap: 5px; padding: 3px 10px; border-radius: 8px; font-size: 12px; font-weight: 600; }
.badge-success { background: #F0FDF4; color: #16A34A; }
.badge-primary { background: #EFF6FF; color: #2563EB; }
.badge-gray { background: #F3F4F6; color: #6B7280; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; }
.dot-success { background: #16A34A; }
.dot-gray { background: #6B7280; }
.mono { font-family: 'SF Mono', 'Consolas', monospace; font-size: 12px; }
</style>
