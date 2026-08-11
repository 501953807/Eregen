<template>
  <div>
    <!-- Welfare Tags -->
    <template v-if="activeTab === 'welfare'">
      <el-row :gutter="16" style="margin-bottom: 16px;">
        <el-col :span="3"><el-button type="primary" @click="$emit('add-tag')">＋ 新增标签</el-button></el-col>
        <el-col :span="3"><el-button>批量分配</el-button></el-col>
      </el-row>
      <el-table :data="welfareTags" v-loading="loading.welfare" stripe>
        <el-table-column prop="tag_code" label="标签代码" width="160"><template #default="{ row }"><span class="mono">{{ row.tag_code }}</span></template></el-table-column>
        <el-table-column prop="tag_name" label="标签名称" width="120" />
        <el-table-column prop="issuer" label="发放机构" width="100" />
        <el-table-column prop="renewal_period_days" label="Renewal 周期" width="110"><template #default="{ row }">{{ row.renewal_period_days }} 天</template></el-table-column>
        <el-table-column prop="benefit_amount" label="补助金额" width="100"><template #default="{ row }">{{ row.benefit_amount > 0 ? '¥' + row.benefit_amount : '¥0' }}</template></el-table-column>
        <el-table-column label="绑定老人" width="90" align="center"><template #default="{ row }"><span style="text-align:center;display:inline-block;width:100%;">{{ countByTag(row.tag_code) }}</span></template></el-table-column>
        <el-table-column label="启用" width="60" align="center"><template #default="{ row }"><span class="enabled-dot"></span></template></el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="$emit('view-tag-elders', row.tag_code, row.tag_name)">查看绑定</el-button>
            <el-button size="small" link>编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <!-- Sign-in -->
    <template v-if="activeTab === 'signin'">
      <el-row :gutter="16" style="margin-bottom: 16px;">
        <el-col :span="5"><el-date-picker v-model="signinPeriod" type="month" placeholder="选择月份" value-format="YYYY-MM" style="width: 100%;" /></el-col>
        <el-col :span="5"><el-select v-model="signinHospital" placeholder="医院筛选" clearable style="width: 100%;"><el-option label="全部医院" value="" /><el-option label="社区医院 A" value="hospital-a" /><el-option label="社区医院 B" value="hospital-b" /><el-option label="社区医院 C" value="hospital-c" /></el-select></el-col>
        <el-col :span="3"><el-button type="primary" @click="loadSigninRecords">🔍 查询</el-button></el-col>
      </el-row>
      <el-card shadow="never" style="margin-bottom: 20px;">
        <template #header><span class="section-title">近 7 天签到趋势</span></template>
        <div class="bar-chart">
          <div class="bar-col" v-for="(day, i) in weekSigninData" :key="i">
            <div class="bar-value">{{ day.count }}</div>
            <div class="bar" :style="{ height: (day.count / maxWeekCount * 100) + 'px', minHeight: '4px' }"></div>
            <div class="bar-label">{{ day.label }}</div>
          </div>
        </div>
      </el-card>
      <el-table :data="signinRecords" v-loading="loading.signin" stripe>
        <el-table-column prop="elder_name" label="老人姓名" width="100"><template #default="{ row }"><strong>{{ row.elder_name || row.elder_id }}</strong></template></el-table-column>
        <el-table-column label="身份证号" width="190"><template #default="{ row }"><span class="mono">{{ (row.elder_id || '').slice(-4) }}</span></template></el-table-column>
        <el-table-column prop="hospital_id" label="医院" width="120" />
        <el-table-column prop="signin_time" label="签到时间" width="180" />
        <el-table-column prop="activated_tags" label="激活标签" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="t in parseTags(row.activated_tags)" :key="t" size="small" style="margin-right: 4px;">{{ t }}</el-tag>
            <span v-if="!parseTags(row.activated_tags)?.length" style="color:var(--el-text-color-placeholder);">—</span>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <span class="status-badge" :class="row.is_welfare_signin ? 'badge-success' : 'badge-primary'">
              <span class="status-dot" :class="row.is_welfare_signin ? 'dot-success' : 'dot-primary'"></span>
              {{ row.is_welfare_signin ? '福利签到' : '医保签到' }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <!-- Pharmacy -->
    <template v-if="activeTab === 'pharmacy'">
      <el-row :gutter="16" style="margin-bottom: 16px;">
        <el-col :span="3"><el-button type="primary">＋ 手动发药</el-button></el-col>
        <el-col :span="5"><el-date-picker v-model="pharmacyMonth" type="month" placeholder="选择月份" value-format="YYYY-MM" style="width: 100%;" /></el-col>
        <el-col :span="6"><el-input v-model="pharmacySearch" placeholder="搜索老人姓名 / 药品名" clearable /></el-col>
      </el-row>
      <el-table :data="pharmacyLogs" v-loading="loading.pharmacy" stripe>
        <el-table-column label="日期" width="70"><template #default="{ row }">{{ formatDate(row.created_at) }}</template></el-table-column>
        <el-table-column prop="elder_name" label="老人姓名" width="100"><template #default="{ row }"><strong>{{ row.elder_name || row.elder_id }}</strong></template></el-table-column>
        <el-table-column prop="hospital_id" label="医院" width="110" />
        <el-table-column prop="items" label="药品清单" min-width="180"><template #default="{ row }">{{ parseItems(row.items).join('、') }}</template></el-table-column>
        <el-table-column label="金额" width="80"><template #default="{ row }">¥{{ row.total_cost?.toFixed(2) || '0.00' }}</template></el-table-column>
        <el-table-column prop="pharmacist_id" label="药师/护士" width="100" />
        <el-table-column label="签到状态" width="90">
          <template #default="{ row }">
            <span class="status-badge" :class="row.signed_in ? 'badge-success' : 'badge-warning'">
              <span class="status-dot" :class="row.signed_in ? 'dot-success' : 'dot-warning'"></span>{{ row.signed_in ? '已签到' : '未签到' }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <!-- Minzheng -->
    <template v-if="activeTab === 'minzheng'">
      <el-row :gutter="16" style="margin-bottom: 20px;">
        <el-col :span="12">
          <el-card shadow="never">
            <template #header><span class="section-title">上传 CSV / XLSX 文件</span></template>
            <div class="upload-zone" @click="triggerFileUpload">
              <div style="font-size:36px;margin-bottom:8px;">📁</div>
              <p>点击或拖拽文件到此处上传</p>
              <p style="font-size:11px;margin-top:4px;color:var(--el-text-color-placeholder);">支持民政局标准模板或自定义模板</p>
            </div>
            <input ref="fileInput" type="file" accept=".csv,.xlsx" style="display:none" @change="handleFileUpload" />
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="never">
            <template #header><span class="section-title">CSV 字段说明</span></template>
            <table class="template-table">
              <thead><tr><th>列名</th><th>必填</th><th>说明</th><th>示例</th></tr></thead>
              <tbody>
                <tr><td>姓名</td><td><el-tag type="danger" size="small">是</el-tag></td><td>老人姓名</td><td>张秀兰</td></tr>
                <tr><td>身份证号</td><td><el-tag type="danger" size="small">是</el-tag></td><td>18 位身份证号码</td><td>510101195001011234</td></tr>
                <tr><td>福利类型</td><td><el-tag type="danger" size="small">是</el-tag></td><td>orphan / poverty_level_1 / ...</td><td>特困</td></tr>
                <tr><td>认定等级</td><td><el-tag type="danger" size="small">是</el-tag></td><td>1 / 2 / 3</td><td>一级</td></tr>
                <tr><td>有效期开始</td><td><el-tag type="danger" size="small">是</el-tag></td><td>YYYY-MM-DD</td><td>2025-01-01</td></tr>
                <tr><td>有效期结束</td><td><el-tag type="danger" size="small">是</el-tag></td><td>YYYY-MM-DD</td><td>2028-12-31</td></tr>
                <tr><td>备注</td><td><el-tag type="info" size="small">否</el-tag></td><td>额外信息</td><td>肢体残疾</td></tr>
              </tbody>
            </table>
          </el-card>
        </el-col>
      </el-row>
      <el-table :data="minzhengSyncs" v-loading="loading.minzheng" stripe>
        <el-table-column prop="source" label="数据来源" width="140" />
        <el-table-column prop="filename" label="文件名" min-width="140"><template #default="{ row }">{{ row.filename || '—' }}</template></el-table-column>
        <el-table-column prop="imported_count" label="导入数" width="80" align="center" />
        <el-table-column prop="matched_count" label="匹配数" width="80" align="center" />
        <el-table-column prop="pending_review_count" label="待审核" width="80" align="center" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <span class="status-badge" :class="minzhengStatusClass(row.status)">
              <span class="status-dot" :class="minzhengStatusDot(row.status)"></span>{{ statusLabel(row.status) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
      </el-table>
    </template>

    <!-- Payments -->
    <template v-if="activeTab === 'payments'">
      <el-row :gutter="16" style="margin-bottom: 16px;">
        <el-col :span="5"><el-select v-model="paymentPeriod" placeholder="选择月份" value-format="YYYY-MM" style="width: 100%;"><el-option v-for="m in paymentPeriods" :key="m" :label="m" :value="m" /></el-select></el-col>
        <el-col :span="3"><el-button type="primary" @click="$emit('execute-payment')">执行发放</el-button></el-col>
      </el-row>
      <el-table :data="batchPayments" v-loading="loading.payments" stripe>
        <el-table-column prop="batch_id" label="批次号" width="170"><template #default="{ row }"><span class="mono">{{ row.batch_id }}</span></template></el-table-column>
        <el-table-column prop="period" label="月份" width="100" />
        <el-table-column prop="pay_type" label="发放类型" width="100" />
        <el-table-column prop="amount" label="金额" width="100"><template #default="{ row }">{{ row.amount > 0 ? '¥' + row.amount : '—' }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <span class="status-badge" :class="paymentStatusClass(row.status)">
              <span class="status-dot" :class="paymentStatusDot(row.status)"></span>{{ statusLabel(row.status) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="failure_reason" label="失败原因" min-width="150"><template #default="{ row }">{{ row.failure_reason || '—' }}</template></el-table-column>
        <el-table-column prop="executed_at" label="执行时间" width="180" />
      </el-table>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import dayjs from 'dayjs'

const props = defineProps<{ activeTab: string; elders: any[]; loading: Record<string, boolean> }>()
const emit = defineEmits<{ 'view-tag-elders': [code: string, name: string]; 'execute-payment': []; 'file-upload': [file: File]; 'add-tag': [] }>()

// Welfare
const welfareTags = ref<any[]>([])
function countByTag(code: string): number { return props.elders.filter(e => (e as any)._welfareTags?.some((t: any) => t.tag_code === code)).length }

// Signin
const signinPeriod = ref(dayjs().format('YYYY-MM'))
const signinHospital = ref('')
const signinRecords = ref<any[]>([])
const weekSigninData = ref<{ label: string; count: number }[]>([])
const maxWeekCount = ref(1)

function generateWeekData() {
  const days = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  const counts = [42, 56, 48, 41, 62, 28, 19]
  weekSigninData.value = days.map((label, i) => ({ label, count: counts[i] }))
  maxWeekCount.value = Math.max(...counts)
}

// Pharmacy
const pharmacyMonth = ref(dayjs().format('YYYY-MM'))
const pharmacySearch = ref('')
const pharmacyLogs = ref<any[]>([])

// Payments
const paymentPeriod = ref(dayjs().format('YYYY-MM'))
const batchPayments = ref<any[]>([])
const paymentPeriods = ref<string[]>([dayjs().format('YYYY-MM'), dayjs().subtract(1, 'month').format('YYYY-MM')])

// Minzheng
const minzhengSyncs = ref<any[]>([])
const fileInput = ref<HTMLInputElement | null>(null)

// Helpers
function parseTags(json: string | undefined): string[] {
  if (!json) return []
  try { return JSON.parse(json) } catch { return [] }
}
function parseItems(json: string | undefined): string[] {
  if (!json) return []
  try { return JSON.parse(json) } catch { return [json] }
}
function formatDate(ts: string | undefined): string { if (!ts) return '—'; return ts.slice(5) }
function statusLabel(status: string): string {
  const map: Record<string, string> = { active: '正常', inactive: '离线', retired: '已退役', success: '成功', failed: '失败', pending: '待处理', retrying: '重试中', completed: '完成', processing: '处理中' }
  return map[status] || status
}
function minzhengStatusClass(s: string): string { if (s === 'completed') return 'badge-success'; if (s === 'failed') return 'badge-danger'; return 'badge-warning' }
function minzhengStatusDot(s: string): string { if (s === 'completed') return 'dot-success'; if (s === 'failed') return 'dot-danger'; return 'dot-warning' }
function paymentStatusClass(s: string): string { if (s === 'success') return 'badge-success'; if (s === 'failed') return 'badge-danger'; if (s === 'retrying') return 'badge-warning'; return 'badge-info' }
function paymentStatusDot(s: string): string { if (s === 'success') return 'dot-success'; if (s === 'failed') return 'dot-danger'; if (s === 'retrying') return 'dot-warning'; return 'dot-info' }
function triggerFileUpload() { fileInput.value?.click() }
function handleFileUpload(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) { emit('file-upload', file); target.value = '' }
}
async function loadSigninRecords() {
  // Placeholder: would fetch from API
  signinRecords.value = []
}

onMounted(() => { generateWeekData() })
</script>

<style scoped>
.section-title { font-size: 15px; font-weight: 700; }
.bar-chart { display: flex; align-items: flex-end; gap: 16px; padding: 16px 0 8px; height: 120px; }
.bar-col { display: flex; flex-direction: column; align-items: center; flex: 1; }
.bar { width: 36px; background: linear-gradient(180deg, #5C8D73, #7BAF8C); border-radius: 3px 3px 0 0; transition: height 0.3s; }
.bar-label { font-size: 11px; color: var(--el-text-color-secondary); margin-top: 6px; }
.bar-value { font-size: 12px; font-weight: 600; color: var(--el-text-color-primary); margin-bottom: 4px; }
.upload-zone { border: 2px dashed var(--el-border-color-light); border-radius: 4px; padding: 32px; text-align: center; color: var(--el-text-color-placeholder); cursor: pointer; transition: border-color 0.2s; }
.upload-zone:hover { border-color: #5C8D73; color: #5C8D73; }
.template-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.template-table th, .template-table td { padding: 8px 12px; border: 1px solid var(--el-border-color-light); text-align: left; }
.template-table th { background: #fafafa; font-weight: 600; color: var(--el-text-color-primary); }
.enabled-dot { width: 8px; height: 8px; border-radius: 50%; background: #16A34A; display: inline-block; }
.status-badge { display: inline-flex; align-items: center; gap: 5px; padding: 3px 10px; border-radius: 8px; font-size: 12px; font-weight: 600; }
.badge-success { background: #F0FDF4; color: #16A34A; }
.badge-danger { background: #FEF2F2; color: #DC2626; }
.badge-warning { background: #FFFBEB; color: #D97706; }
.badge-primary { background: #DDEBE1; color: #47745C; }
.badge-gray { background: #F3F4F6; color: #6B7280; }
.badge-info { background: #F8FAFC; color: #94A3B8; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; }
.dot-success { background: #16A34A; }
.dot-danger { background: #DC2626; }
.dot-warning { background: #D97706; }
.dot-primary { background: #5C8D73; }
.dot-gray { background: #6B7280; }
.dot-info { background: #94A3B8; }
.mono { font-family: 'SF Mono', 'Consolas', monospace; font-size: 12px; }
</style>
