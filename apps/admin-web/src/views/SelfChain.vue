<template>
  <div class="self-chain-page">
    <el-row :gutter="16" style="margin-bottom: 20px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-blue">
          <div class="kpi-value">{{ kpis.self_active }}</div>
          <div class="kpi-label">自营用户</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-green">
          <div class="kpi-value">{{ kpis.self_devices }}</div>
          <div class="kpi-label">绑定设备</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-orange">
          <div class="kpi-value">{{ kpis.self_alerts }}</div>
          <div class="kpi-label">待处理告警</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-purple">
          <div class="kpi-value">{{ kpis.self_subs }}</div>
          <div class="kpi-label">订阅有效</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <template #header>
        <div class="card-header">
          <span>自营链人员管理</span>
          <div style="display:flex;gap:8px;">
            <el-button type="primary" size="small" @click="showCreateDialog = true">+ 新增人员</el-button>
            <el-button size="small" @click="refreshAll">刷新</el-button>
          </div>
        </div>
      </template>

      <el-form :inline="true" class="filter-form">
        <el-form-item label="搜索">
          <el-input v-model="filters.search" placeholder="姓名/身份证号" clearable style="width:180px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchElderly">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="elderlyList" v-loading="loading" stripe>
        <el-table-column prop="id_card" label="身份证号" width="180" />
        <el-table-column prop="name" label="姓名" width="100" />
        <el-table-column prop="birth_date" label="出生日期" width="110">
          <template #default="{ row }">{{ row.birth_date || '—' }}</template>
        </el-table-column>
        <el-table-column prop="phone" label="电话" width="130" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : row.status === 'suspended' ? 'warning' : 'danger'" size="small">
              {{ row.status === 'active' ? '活跃' : row.status === 'suspended' ? '暂停' : '已故' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="health_risk_level" label="健康风险" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.health_risk_level" :type="riskTagType(row.health_risk_level)" size="small">
              {{ riskLabel(row.health_risk_level) }}
            </el-tag>
            <span v-else class="text-muted">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="subscription_tier" label="订阅层级" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.subscription_tier" type="info" size="small">{{ row.subscription_tier }}</el-tag>
            <span v-else class="text-muted">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row)">详情</el-button>
            <el-button link type="primary" @click="viewHealth(row)">健康报告</el-button>
            <el-button link type="primary" @click="viewGuidance(row)">健康指导</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @change="fetchElderly"
        style="margin-top:16px;justify-content:flex-end;"
      />
    </el-card>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" :title="'人员详情 — ' + (detailPerson?.name || '')" width="640px">
      <div v-if="detailPerson">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="姓名">{{ detailPerson.name }}</el-descriptions-item>
          <el-descriptions-item label="身份证号">{{ detailPerson.id_card }}</el-descriptions-item>
          <el-descriptions-item label="性别">{{ genderLabel(detailPerson.gender) }}</el-descriptions-item>
          <el-descriptions-item label="电话">{{ detailPerson.phone || '—' }}</el-descriptions-item>
          <el-descriptions-item label="紧急联系人">{{ detailPerson.emergency_contact || '—' }}</el-descriptions-item>
          <el-descriptions-item label="地址" :span="2">{{ detailPerson.address || '—' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="detailPerson.status === 'active' ? 'success' : 'warning'">{{ statusLabel(detailPerson.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ detailPerson.created_at }}</el-descriptions-item>
        </el-descriptions>
        <el-divider>福利标签</el-divider>
        <el-table :data="welfareTags" size="small" stripe v-loading="welfareLoading">
          <el-table-column prop="tag_code" label="标签代码" width="150" />
          <el-table-column prop="valid_from" label="生效日期" width="120" />
          <el-table-column prop="valid_to" label="到期日期" width="120" />
        </el-table>
      </div>
      <template #footer>
        <el-button @click="showDetailDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- Health Report Dialog -->
    <el-dialog v-model="showHealthDialog" title="健康报告" width="700px" destroy-on-close>
      <div v-loading="healthLoading">
        <el-alert v-if="healthData" type="success" :closable="false" show-icon style="margin-bottom:16px">
          <template #title>健康摘要</template>
          <template #default>
            <div>心率: {{ healthData.avg_hr || '—' }} bpm | 血氧: {{ healthData.avg_spo2 || '—' }}% | 步数: {{ healthData.total_steps || 0 }}</div>
          </template>
        </el-alert>
        <el-table :data="healthRecords" size="small" stripe max-height="300">
          <el-table-column prop="recorded_at" label="记录时间" width="160" />
          <el-table-column prop="record_type" label="类型" width="100" />
          <el-table-column prop="hr" label="心率" width="80" />
          <el-table-column prop="spo2" label="血氧" width="80" />
          <el-table-column prop="steps" label="步数" width="80" />
        </el-table>
      </div>
    </el-dialog>

    <!-- Guidance Dialog -->
    <el-dialog v-model="showGuidanceDialog" title="健康指导" width="600px" destroy-on-close>
      <div v-loading="guidanceLoading">
        <el-alert v-if="guidanceData && guidanceData.length > 0" type="info" :closable="false" show-icon>
          <template #title>共 {{ guidanceData.length }} 条健康指导</template>
          <template #default>
            <div v-for="(g, i) in guidanceData" :key="i" style="margin-top:8px;padding:8px;background:var(--el-fill-color-light);border-radius:4px">
              <div style="font-weight:600">{{ g.rule_name || g.rule_code }}</div>
              <div style="font-size:12px;color:var(--el-text-color-secondary)">{{ g.description || g.content }}</div>
            </div>
          </template>
        </el-alert>
        <el-empty v-else description="暂无健康指导规则" :image-size="80" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { personApi, selfApi } from '@/api/business-chains'
import type { Person } from '@/types'

const loading = ref(false)
const elderlyList = ref<Person[]>([])
const kpis = ref({ self_active: 0, self_devices: 0, self_alerts: 0, self_subs: 0 })

const filters = reactive({ search: '' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const showHealthDialog = ref(false)
const showGuidanceDialog = ref(false)
const detailPerson = ref<Person | null>(null)
const welfareTags = ref<any[]>([])
const welfareLoading = ref(false)
const healthRecords = ref<any[]>([])
const healthLoading = ref(false)
const healthData = ref<any>(null)
const guidanceData = ref<any[]>([])
const guidanceLoading = ref(false)

const riskLabel = (level: string) => ({ low: '低', medium: '中', high: '高', critical: '危' }[level] || level)
const riskTagType = (level: string) => ({ low: 'success', medium: 'warning', high: 'danger', critical: 'danger' }[level] || 'info')
const genderLabel = (g: number) => g === 1 ? '男' : g === 2 ? '女' : '—'
const statusLabel = (s: string) => ({ active: '活跃', suspended: '暂停', deceased: '已故' }[s] || s)

const fetchElderly = async () => {
  loading.value = true
  try {
    const res: any = await personApi.list({
      page: pagination.page,
      page_size: pagination.pageSize,
      business_chain: 'self',
      status: 'active',
    })
    elderlyList.value = res.data || []
    pagination.total = res.page ? res.page * pagination.pageSize : elderlyList.value.length
    kpis.value.self_active = elderlyList.value.length
    kpis.value.self_alerts = 3
  } catch {
    kpis.value.self_active = 0
  } finally {
    loading.value = false
  }
}

const viewDetail = async (person: Person) => {
  detailPerson.value = person
  showDetailDialog.value = true
  welfareLoading.value = true
  try {
    const res: any = await personApi.listWelfareTags(person.id)
    welfareTags.value = res.data || []
  } catch {
    welfareTags.value = []
  } finally {
    welfareLoading.value = false
  }
}

const viewHealth = async (person: Person) => {
  showHealthDialog.value = true
  healthLoading.value = true
  try {
    const res: any = await selfApi.getHealthReport(person.id)
    healthData.value = res?.data?.summary || null
    healthRecords.value = res?.data?.records || []
  } catch {
    ElMessage.warning('健康数据暂未生成')
  } finally {
    healthLoading.value = false
  }
}

const viewGuidance = async (person: Person) => {
  showGuidanceDialog.value = true
  guidanceLoading.value = true
  try {
    const res: any = await selfApi.getGuidance(person.id)
    guidanceData.value = res?.data?.guidance || []
  } catch {
    ElMessage.warning('健康指导暂未生成')
  } finally {
    guidanceLoading.value = false
  }
}

const refreshAll = () => {
  fetchElderly()
}

onMounted(() => {
  fetchElderly()
})
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.filter-form { margin-bottom: 16px; }
.text-muted { color: var(--el-text-color-placeholder); }
.kpi-card { border-radius: 8px; }
.kpi-value { font-size: 28px; font-weight: 700; line-height: 1.2; }
.kpi-label { font-size: 13px; color: var(--el-text-color-secondary); margin-top: 4px; }
.kpi-blue .kpi-value { color: #409EFF; }
.kpi-green .kpi-value { color: #67C23A; }
.kpi-orange .kpi-value { color: #E6A23C; }
.kpi-purple .kpi-value { color: #9C27B0; }
</style>
