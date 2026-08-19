<template>
  <div>
    <!-- Welfare -->
    <template v-if="activePage === 'welfare'">
      <el-row :gutter="12" style="margin-bottom: 20px;">
        <el-col :span="6"><el-card shadow="never" class="kpi-card"><div class="kpi-value">{{ welfareKpis.valid }}</div><div class="kpi-label">有效标签</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="never" class="kpi-card kpi-warning"><div class="kpi-value">{{ welfareKpis.expiring }}</div><div class="kpi-label">本月到期</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="never" class="kpi-card kpi-success"><div class="kpi-value">{{ welfareKpis.newIssued }}</div><div class="kpi-label">本月新发</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="never" class="kpi-card"><div class="kpi-value">{{ welfareKpis.revoked }}</div><div class="kpi-label">本月撤销</div></el-card></el-col>
      </el-row>
      <div class="filter-bar">
        <el-button type="primary">＋ 新增标签</el-button>
        <el-button>批量分配</el-button>
      </div>
      <el-card shadow="never" class="table-card">
        <el-table :data="welfareList" stripe>
          <el-table-column prop="code" label="标签代码" width="150">
            <template #default="{ row }"><span class="mono">{{ row.code }}</span></template>
          </el-table-column>
          <el-table-column prop="name" label="标签名称" width="120" />
          <el-table-column prop="issuer" label="发放机构" width="100" />
          <el-table-column label="Renewal 周期" width="100">
            <template #default="{ row }">{{ row.renewal_days }} 天</template>
          </el-table-column>
          <el-table-column label="补助金额" width="100">
            <template #default="{ row }">¥{{ row.subsidy_amount }}</template>
          </el-table-column>
          <el-table-column label="绑定老人" width="90" align="center">
            <template #default="{ row }"><strong>{{ row.bound_count }}</strong></template>
          </el-table-column>
          <el-table-column label="启用" width="80" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.enabled" @change="(_v: any) => ($emit as any)('toggle-welfare', row, _v)" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="(_v: any) => $emit('view-bound', row)">查看绑定</el-button>
              <el-button link type="primary" size="small">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>

    <!-- Signin -->
    <template v-if="activePage === 'signin'">
      <el-row :gutter="12" style="margin-bottom: 20px;">
        <el-col :span="4"><el-card shadow="never" class="kpi-card"><div class="kpi-value">856</div><div class="kpi-label">本月签到</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="never" class="kpi-card kpi-success"><div class="kpi-value">234</div><div class="kpi-label">本月首次</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="never" class="kpi-card kpi-danger"><div class="kpi-value">3</div><div class="kpi-label">跨院重复</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="never" class="kpi-card"><div class="kpi-value">189</div><div class="kpi-label">医保签到</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="never" class="kpi-card kpi-success"><div class="kpi-value">667</div><div class="kpi-label">福利签到</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="never" class="kpi-card kpi-warning"><div class="kpi-value">2</div><div class="kpi-label">异常</div></el-card></el-col>
      </el-row>
      <div class="filter-bar">
        <el-date-picker v-model="signinMonth" type="month" placeholder="月份选择" value-format="YYYY-MM" />
        <el-select v-model="signinHospital" placeholder="全部医院" style="width: 140px;">
          <el-option label="社区医院 A" value="A" />
          <el-option label="社区医院 B" value="B" />
          <el-option label="社区医院 C" value="C" />
        </el-select>
        <el-button type="primary">查询</el-button>
      </div>
      <el-card shadow="never" class="chart-card">
        <template #header><span class="panel-title">近 7 天签到趋势</span></template>
        <div class="bar-chart">
          <div v-for="(d, i) in signinTrend" :key="i" class="bar-col">
            <div class="bar-value">{{ d.count }}</div>
            <div class="bar" :style="{ height: (d.count / 62 * 100) + 'px' }"></div>
            <div class="bar-label">{{ d.day }}</div>
          </div>
        </div>
      </el-card>
      <el-card shadow="never" class="table-card" style="margin-top: 20px;">
        <el-table :data="signinRecords" stripe>
          <el-table-column prop="name" label="老人姓名" width="100" />
          <el-table-column label="身份证号" width="140">
            <template #default="{ row }"><span class="mono">{{ row.id_card }}</span></template>
          </el-table-column>
          <el-table-column prop="hospital" label="医院" width="120" />
          <el-table-column prop="signin_time" label="签到时间" width="170" />
          <el-table-column label="激活标签" min-width="180">
            <template #default="{ row }">
              <el-tag v-for="t in (row.tags || [])" :key="t" size="small" style="margin-right: 4px;">{{ t }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="类型" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="row.type === '福利签到' ? 'badge-success' : 'badge-primary'">
                <span class="status-dot" :class="row.type === '福利签到' ? 'dot-success' : 'dot-primary'"></span>{{ row.type }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>

    <!-- Pharmacy -->
    <template v-if="activePage === 'pharmacy'">
      <el-row :gutter="12" style="margin-bottom: 20px;">
        <el-col :span="6"><el-card shadow="never" class="kpi-card"><div class="kpi-value">34</div><div class="kpi-label">今日发药</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="never" class="kpi-card kpi-success"><div class="kpi-value">512</div><div class="kpi-label">本月发药</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="never" class="kpi-card"><div class="kpi-value">28</div><div class="kpi-label">药品种类</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="never" class="kpi-card kpi-warning"><div class="kpi-value">¥12,450</div><div class="kpi-label">总金额</div></el-card></el-col>
      </el-row>
      <div class="filter-bar">
        <el-button type="primary">＋ 手动发药</el-button>
        <el-date-picker v-model="pharmacyMonth" type="month" placeholder="月份" value-format="YYYY-MM" />
        <el-input v-model="pharmacySearch" placeholder="搜索老人姓名 / 药品名" clearable style="width: 220px;">
          <template #prefix><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg></template>
        </el-input>
      </div>
      <el-card shadow="never" class="table-card">
        <el-table :data="pharmacyRecords" stripe>
          <el-table-column label="日期" width="80"><template #default="{ row }">{{ row.date }}</template></el-table-column>
          <el-table-column prop="name" label="老人姓名" width="100" />
          <el-table-column prop="hospital" label="医院" width="110" />
          <el-table-column prop="medications" label="药品清单" min-width="180" show-overflow-tooltip />
          <el-table-column label="金额" width="80"><template #default="{ row }">¥{{ row.amount }}</template></el-table-column>
          <el-table-column prop="staff" label="药师/护士" width="90" />
          <el-table-column label="签到状态" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="row.signed_in ? 'badge-success' : 'badge-warning'">
                <span class="status-dot" :class="row.signed_in ? 'dot-success' : 'dot-warning'"></span>{{ row.signed_in ? '已签到' : '未签到' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-wrapper">
          <el-pagination background layout="total, prev, pager, next" :total="512" :current-page="1" :page-size="20" />
        </div>
      </el-card>
    </template>

    <!-- Minzheng -->
    <template v-if="activePage === 'minzheng'">
      <el-row :gutter="12" style="margin-bottom: 20px;">
        <el-col :span="6"><el-card shadow="never" class="kpi-card"><div class="kpi-value">12</div><div class="kpi-label">导入批次</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="never" class="kpi-card kpi-success"><div class="kpi-value">1,234</div><div class="kpi-label">总导入</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="never" class="kpi-card"><div class="kpi-value">1,198</div><div class="kpi-label">已匹配</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="never" class="kpi-card kpi-warning"><div class="kpi-value">36</div><div class="kpi-label">待审核</div></el-card></el-col>
      </el-row>
      <div class="filter-bar">
        <el-upload action="#" :auto-upload="false" :show-file-list="false" class="upload-zone">
          <div class="upload-inner">
            <div class="upload-icon">📁</div>
            <p>点击或拖拽 CSV/XLSX 文件到此处上传</p>
            <p style="font-size:11px;margin-top:4px;color:var(--hope-text-muted);">支持民政局标准模板或自定义模板</p>
          </div>
        </el-upload>
        <el-button>📥 下载 CSV 模板</el-button>
      </div>
      <el-card shadow="never" class="table-card" style="margin-bottom: 20px;">
        <template #header><span class="panel-title">CSV 字段说明</span></template>
        <el-table :data="csvTemplateFields" stripe size="small">
          <el-table-column prop="field" label="列名" width="120" />
          <el-table-column label="必填" width="80">
            <template #default="{ row }">
              <span class="status-badge" :class="row.required ? 'badge-danger' : 'badge-gray'">
                <span class="status-dot" :class="row.required ? 'dot-danger' : 'dot-gray'"></span>{{ row.required ? '是' : '否' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="desc" label="说明" />
          <el-table-column prop="example" label="示例" width="180">
            <template #default="{ row }"><span class="mono">{{ row.example }}</span></template>
          </el-table-column>
        </el-table>
      </el-card>
      <el-card shadow="never" class="table-card">
        <el-table :data="importRecords" stripe>
          <el-table-column prop="source" label="数据来源" width="120" />
          <el-table-column prop="filename" label="文件名" width="140" />
          <el-table-column label="导入数" width="80"><template #default="{ row }"><strong>{{ row.imported }}</strong></template></el-table-column>
          <el-table-column label="匹配数" width="80"><template #default="{ row }">{{ row.matched }}</template></el-table-column>
          <el-table-column label="待审核" width="80"><template #default="{ row }"><strong :style="{ color: row.pending > 0 ? 'var(--hope-error)' : '' }">{{ row.pending }}</strong></template></el-table-column>
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="row.status === '完成' ? 'badge-success' : 'badge-warning'">
                <span class="status-dot" :class="row.status === '完成' ? 'dot-success' : 'dot-warning'"></span>{{ row.status }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="170" />
        </el-table>
      </el-card>
    </template>

    <!-- Stats -->
    <template v-if="activePage === 'stats'">
      <div class="filter-bar" style="margin-bottom: 20px;">
        <el-date-picker v-model="statsMonth" type="month" placeholder="月份" value-format="YYYY-MM" />
        <el-select v-model="statsHospital" placeholder="全部医院" style="width: 140px;">
          <el-option label="社区医院 A" value="A" />
          <el-option label="社区医院 B" value="B" />
          <el-option label="社区医院 C" value="C" />
        </el-select>
        <el-button type="primary">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg>刷新
        </el-button>
      </div>
      <el-row :gutter="16" style="margin-bottom: 16px;">
        <el-col :span="8">
          <el-card shadow="never" class="stat-box">
            <template #header><span class="panel-title">登记老人总数</span></template>
            <div class="stat-center"><div class="stat-big-num">482</div><div style="font-size:12px;color:var(--hope-text-muted);">在线腕带 312 · 离线 170</div></div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="never" class="stat-box">
            <template #header><span class="panel-title">福利标签分布</span></template>
            <div class="h-bars">
              <div v-for="w in welfareDist" :key="w.code" class="h-bar-row">
                <span class="h-bar-label">{{ w.label }}</span>
                <div class="h-bar-track"><div class="h-bar-fill" :style="{ width: Math.min(w.pct, 100) + '%', background: w.color }"></div></div>
                <span class="h-bar-val">{{ w.count }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="never" class="stat-box">
            <template #header><span class="panel-title">签到活跃度</span></template>
            <div class="activity-stats">
              <div class="act-row"><span>本月签到率</span><el-progress :percentage="87" :color="'var(--hope-success)'" :stroke-width="8" /><strong>87%</strong></div>
              <div class="act-row"><span>连续签到≥3月</span><el-progress :percentage="68" :color="'var(--hope-success)'" :stroke-width="8" /><strong>68%</strong></div>
              <div class="act-row"><span>本月首次签到</span><span style="font-weight:600;color:var(--hope-warning);">234 人</span></div>
              <div class="act-row"><span>跨院重复</span><span style="font-weight:600;color:var(--hope-error);">3 人次</span></div>
            </div>
          </el-card>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="8">
          <el-card shadow="never" class="stat-box">
            <template #header><span class="panel-title">医院分布</span></template>
            <div class="h-bars">
              <div v-for="h in hospitalDist" :key="h.name" class="h-bar-row">
                <span class="h-bar-label">{{ h.name }}</span>
                <div class="h-bar-track"><div class="h-bar-fill" :style="{ width: h.pct + '%', background: h.color }"></div></div>
                <span class="h-bar-val">{{ h.count }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="never" class="stat-box">
            <template #header><span class="panel-title">补助发放统计</span></template>
            <div class="payment-stats">
              <div class="pay-total">¥45,200</div>
              <div style="font-size:12px;color:var(--hope-text-muted);margin-bottom:12px;">本月发放总额</div>
              <div class="pay-metrics">
                <div class="pay-metric"><div style="font-size:18px;font-weight:600;color:var(--hope-success);">92%</div><div style="font-size:11px;color:var(--hope-text-muted);">成功率</div></div>
                <div class="pay-metric"><div style="font-size:18px;font-weight:600;color:var(--hope-error);">8</div><div style="font-size:11px;color:var(--hope-text-muted);">失败笔数</div></div>
                <div class="pay-metric"><div style="font-size:18px;font-weight:600;color:var(--hope-warning);">12</div><div style="font-size:11px;color:var(--hope-text-muted);">待发笔数</div></div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="never" class="stat-box">
            <template #header><span class="panel-title">规则引擎告警</span></template>
            <div class="alert-list">
              <div v-for="a in ruleAlerts" :key="a.code" class="alert-item">
                <span class="status-badge" :class="a.tagType === 'danger' ? 'badge-danger' : a.tagType === 'warning' ? 'badge-warning' : 'badge-gray'">
                  <span class="status-dot" :class="a.tagType === 'danger' ? 'dot-danger' : a.tagType === 'warning' ? 'dot-warning' : 'dot-gray'"></span>{{ a.code }}</span>
                <span>{{ a.desc }}</span>
                <strong>{{ a.count }}</strong>
              </div>
            </div>
            <div style="text-align:center;margin-top:12px;"><el-button size="small">查看全部告警 →</el-button></div>
          </el-card>
        </el-col>
      </el-row>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ activePage: string }>()
const emit = defineEmits<{ 'toggle-welfare': [row: any, enabled: boolean]; 'view-bound': [row: any] }>()

// Welfare
const welfareKpis = ref({ valid: 9, expiring: 3, newIssued: 12, revoked: 2 })
const welfareList = ref([
  { code: 'orphan', name: '孤寡老人', issuer: '民政局', renewal_days: 365, subsidy_amount: 0, bound_count: 12, enabled: true },
  { code: 'poverty_level_1', name: '特困一级', issuer: '民政局', renewal_days: 365, subsidy_amount: 800, bound_count: 28, enabled: true },
  { code: 'poverty_level_2', name: '特困二级', issuer: '民政局', renewal_days: 365, subsidy_amount: 500, bound_count: 15, enabled: true },
  { code: 'disability_1', name: '残疾一级', issuer: '残联', renewal_days: 365, subsidy_amount: 600, bound_count: 22, enabled: true },
  { code: 'disability_2', name: '残疾二级', issuer: '残联', renewal_days: 365, subsidy_amount: 400, bound_count: 35, enabled: true },
  { code: 'disability_3', name: '残疾三级', issuer: '残联', renewal_days: 365, subsidy_amount: 200, bound_count: 42, enabled: true },
  { code: 'special_disease', name: '特病门诊', issuer: '医保局', renewal_days: 180, subsidy_amount: 0, bound_count: 89, enabled: true },
  { code: 'bus_discount', name: '公交优惠', issuer: '交通局', renewal_days: 30, subsidy_amount: 0, bound_count: 156, enabled: true },
  { code: 'medical_assist', name: '医疗救助', issuer: '民政局', renewal_days: 365, subsidy_amount: 1000, bound_count: 67, enabled: true },
])

// Signin
const signinMonth = ref('2026-07')
const signinHospital = ref('')
const signinTrend = [
  { day: '周一', count: 42 }, { day: '周二', count: 56 }, { day: '周三', count: 48 },
  { day: '周四', count: 41 }, { day: '周五', count: 62 }, { day: '周六', count: 28 }, { day: '周日', count: 19 },
]
const signinRecords = ref([
  { name: '张秀兰', id_card: '...1234', hospital: '社区医院 A', signin_time: '2026-07-23 10:30', tags: ['孤寡', '特困一级', '残疾二级'], type: '福利签到' },
  { name: '李建国', id_card: '...5678', hospital: '社区医院 B', signin_time: '2026-07-23 09:15', tags: ['特病门诊', '公交优惠'], type: '医保签到' },
  { name: '王秀英', id_card: '...7890', hospital: '社区医院 A', signin_time: '2026-07-22 14:20', tags: ['医疗救助'], type: '福利签到' },
  { name: '赵德柱', id_card: '...3456', hospital: '社区医院 C', signin_time: '2026-07-22 11:05', tags: ['孤寡', '特困一级', '特病门诊'], type: '福利签到' },
  { name: '刘美华', id_card: '...6789', hospital: '社区医院 A', signin_time: '2026-07-21 16:40', tags: ['残疾三级', '公交优惠'], type: '福利签到' },
])

// Pharmacy
const pharmacyMonth = ref('2026-07')
const pharmacySearch = ref('')
const pharmacyRecords = ref([
  { date: '07-23', name: '张秀兰', hospital: '社区医院 A', medications: '氨氯地平、二甲双胍', amount: '45.50', staff: '张护士', signed_in: true },
  { date: '07-23', name: '李建国', hospital: '社区医院 B', medications: '阿司匹林肠溶片', amount: '12.00', staff: '李药师', signed_in: true },
  { date: '07-22', name: '王秀英', hospital: '社区医院 A', medications: '硝苯地平缓释片', amount: '28.00', staff: '张护士', signed_in: false },
  { date: '07-22', name: '赵德柱', hospital: '社区医院 C', medications: '格列本脲、二甲双胍', amount: '67.30', staff: '王药师', signed_in: true },
  { date: '07-21', name: '刘美华', hospital: '社区医院 A', medications: '氨氯地平', amount: '18.50', staff: '张护士', signed_in: true },
])

// Minzheng
const csvTemplateFields = [
  { field: '姓名', required: true, desc: '老人姓名', example: '张秀兰' },
  { field: '身份证号', required: true, desc: '18位身份证号码', example: '510101195001011234' },
  { field: '福利类型', required: true, desc: 'orphan/poverty_level_1/disability_2/...', example: '特困' },
  { field: '认定等级', required: true, desc: '1/2/3', example: '一级' },
  { field: '有效期开始', required: true, desc: 'YYYY-MM-DD', example: '2025-01-01' },
  { field: '有效期结束', required: true, desc: 'YYYY-MM-DD', example: '2028-12-31' },
  { field: '备注', required: false, desc: '额外信息', example: '肢体残疾' },
]
const importRecords = ref([
  { source: 'XX区民政局', filename: '202607.csv', imported: 234, matched: 230, pending: 4, status: '完成', created_at: '2026-07-23 10:00' },
  { source: 'XX街道办', filename: '7月数据.xlsx', imported: 156, matched: 155, pending: 1, status: '完成', created_at: '2026-07-20 14:30' },
  { source: 'XX区残联', filename: '残疾人补贴.csv', imported: 89, matched: 85, pending: 4, status: '完成', created_at: '2026-07-18 09:15' },
  { source: 'XX市民政局', filename: '特困人员汇总.csv', imported: 312, matched: 298, pending: 14, status: '处理中', created_at: '2026-07-24 08:00' },
])

// Stats
const statsMonth = ref('2026-07')
const statsHospital = ref('')
const welfareDist = [
  { code: 'orphan', label: '孤寡', count: 12, pct: 12, color: 'var(--hope-error)' },
  { code: 'poverty_1', label: '特困一', count: 28, pct: 28, color: 'var(--hope-warning)' },
  { code: 'poverty_2', label: '特困二', count: 15, pct: 15, color: '#F59E0B' },
  { code: 'disability_1', label: '残疾一', count: 22, pct: 22, color: 'var(--hope-primary)' },
  { code: 'disability_2', label: '残疾二', count: 35, pct: 35, color: '#5B7EE8' },
  { code: 'disability_3', label: '残疾三', count: 42, pct: 42, color: '#8BA8F5' },
  { code: 'special_disease', label: '特病', count: 89, pct: 89, color: 'var(--hope-accent)' },
  { code: 'bus_discount', label: '公交', count: 156, pct: 100, color: 'var(--hope-success)' },
  { code: 'medical_assist', label: '医疗', count: 67, pct: 67, color: 'var(--hope-info)' },
]
const hospitalDist = [
  { name: '社区医院A', count: 234, pct: 100, color: 'var(--hope-primary)' },
  { name: '社区医院B', count: 156, pct: 67, color: 'var(--hope-accent)' },
  { name: '社区医院C', count: 92, pct: 39, color: '#A88AFF' },
]
const ruleAlerts = [
  { code: 'R_C01', desc: '重复领取', count: 3, tagType: 'danger' },
  { code: 'R_C02', desc: '冒领嫌疑', count: 1, tagType: 'danger' },
  { code: 'R_C03', desc: '异常高频', count: 2, tagType: 'warning' },
  { code: 'R_C04', desc: '僵尸账户', count: 5, tagType: 'info' },
  { code: 'R_C05', desc: '补助未到账', count: 1, tagType: 'warning' },
]
</script>

<style scoped>
.mono { font-family: 'SF Mono', 'Consolas', monospace; font-size: 12px; color: var(--hope-text-muted); }
.filter-bar { display: flex; gap: 12px; align-items: center; margin-bottom: 16px; flex-wrap: wrap; }
.table-card { margin-bottom: 20px; }
.pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 16px; }
.panel-title { font-size: 15px; font-weight: 700; color: var(--hope-text); border-left: 3px solid var(--hope-primary); padding-left: 8px; }
.upload-zone { border: 2px dashed var(--hope-border); border-radius: var(--hope-radius-lg); padding: 32px; text-align: center; cursor: pointer; background: var(--hope-surface-light); }
.upload-inner p { font-size: 13px; color: var(--hope-text-secondary); }
.upload-icon { font-size: 36px; margin-bottom: 8px; }
.bar-chart { display: flex; align-items: flex-end; gap: 16px; padding: 16px 0; height: 130px; }
.bar-col { display: flex; flex-direction: column; align-items: center; flex: 1; }
.bar { width: 32px; background: linear-gradient(180deg, var(--hope-primary), var(--hope-accent)); border-radius: 4px 4px 0 0; min-height: 4px; transition: height 0.3s; }
.bar-label { font-size: 11px; color: var(--hope-text-muted); margin-top: 6px; }
.bar-value { font-size: 12px; font-weight: 600; color: var(--hope-text); margin-bottom: 4px; }
.h-bars { display: flex; flex-direction: column; gap: 6px; }
.h-bar-row { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.h-bar-label { width: 60px; text-align: right; color: var(--hope-text-secondary); flex-shrink: 0; font-size: 12px; }
.h-bar-track { flex: 1; height: 14px; background: var(--hope-surface-light); border-radius: 3px; overflow: hidden; }
.h-bar-fill { height: 100%; border-radius: 3px; transition: width 0.3s; }
.h-bar-val { width: 40px; color: var(--hope-text-muted); font-size: 12px; flex-shrink: 0; text-align: right; }
.stat-box :deep(.el-card__body) { padding: 16px; }
.stat-center { text-align: center; padding: 12px 0; }
.stat-big-num { font-size: 36px; font-weight: 800; background: linear-gradient(135deg, var(--hope-primary), var(--hope-accent)); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
.activity-stats { display: flex; flex-direction: column; gap: 12px; }
.act-row { display: flex; align-items: center; justify-content: space-between; font-size: 13px; gap: 8px; }
.act-row span:first-child { color: var(--hope-text-secondary); white-space: nowrap; }
.payment-stats { text-align: center; padding: 8px 0; }
.pay-total { font-size: 24px; font-weight: 800; color: var(--hope-primary); }
.pay-metrics { display: flex; justify-content: space-around; }
.pay-metric { text-align: center; }
.alert-list { display: flex; flex-direction: column; gap: 8px; }
.alert-item { display: flex; align-items: center; gap: 8px; padding: 6px 0; font-size: 13px; border-bottom: 1px solid var(--hope-border); }
.alert-item:last-child { border-bottom: none; }
.status-badge { display: inline-flex; align-items: center; gap: 5px; padding: 3px 10px; border-radius: var(--hope-radius-pill); font-size: 12px; font-weight: 600; }
.badge-success { background: var(--hope-success-light); color: var(--hope-success); }
.badge-danger { background: var(--hope-error-light); color: var(--hope-error); }
.badge-warning { background: var(--hope-warning-light); color: #926C0E; }
.badge-primary { background: rgba(var(--hope-primary-rgb), 0.12); color: var(--hope-primary); }
.badge-gray { background: var(--hope-surface-light); color: var(--hope-text-muted); }
.status-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; }
.dot-success { background: var(--hope-success); }
.dot-danger { background: var(--hope-error); }
.dot-warning { background: var(--hope-warning); }
.dot-primary { background: var(--hope-primary); }
.dot-gray { background: var(--hope-text-muted); }
.kpi-card :deep(.el-card__body) { padding: 18px; display: flex; flex-direction: column; align-items: center; text-align: center; border-radius: 14px; }
.kpi-value { font-size: 32px; font-weight: 800; letter-spacing: -0.03em; line-height: 1; margin-bottom: 4px; }
.kpi-label { font-size: 12px; color: var(--hope-text-muted); margin-top: 6px; font-weight: 600; }
</style>
