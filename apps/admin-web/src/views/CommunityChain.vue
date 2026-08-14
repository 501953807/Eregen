<template>
  <div class="community-chain-page">
    <el-row :gutter="16" style="margin-bottom: 20px;">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-blue">
          <div class="kpi-value">{{ kpis.total_elders }}</div>
          <div class="kpi-label">社区老人</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-green">
          <div class="kpi-value">{{ kpis.today_signin }}</div>
          <div class="kpi-label">今日签到</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-orange">
          <div class="kpi-value">{{ kpis.welfare_tags }}</div>
          <div class="kpi-label">福利标签</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card kpi-purple">
          <div class="kpi-value">{{ kpis.pending_payments }}</div>
          <div class="kpi-label">待结算</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <template #header>
        <div class="card-header">
          <span>社区老人管理</span>
          <div style="display:flex;gap:8px;">
            <el-button type="primary" size="small" @click="showCreateDialog = true">+ 新增老人</el-button>
            <el-button size="small" @click="refreshAll">刷新</el-button>
          </div>
        </div>
      </template>

      <el-form :inline="true" class="filter-form">
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable>
            <el-option label="正常" value="active" />
            <el-option label="停用" value="inactive" />
            <el-option label="已退役" value="retired" />
          </el-select>
        </el-form-item>
        <el-form-item label="搜索">
          <el-input v-model="filters.search" placeholder="姓名/身份证号" clearable style="width:180px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchElders">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="elders" v-loading="loading" stripe>
        <el-table-column prop="id_card" label="身份证号" width="180" />
        <el-table-column prop="name" label="姓名" width="100" />
        <el-table-column prop="welfare_tags" label="福利标签" width="200">
          <template #default="{ row }">
            <el-tag v-for="tag in parseTags(row.welfare_tags)" :key="tag" size="small" class="mr-1">{{ tag }}</el-tag>
            <span v-if="!row.welfare_tags" class="text-muted">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="last_signin" label="最后签到" width="140">
          <template #default="{ row }">{{ row.last_signin || '—' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : row.status === 'inactive' ? 'warning' : 'info'" size="small">
              {{ row.status === 'active' ? '正常' : row.status === 'inactive' ? '停用' : '已退役' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row)">详情</el-button>
            <el-button link type="primary" @click="triggerSignin(row)">签到</el-button>
            <el-button link type="primary" @click="assignWelfare(row)">福利</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @change="fetchElders"
        style="margin-top:16px;justify-content:flex-end;"
      />
    </el-card>

    <!-- Create Elder Dialog -->
    <el-dialog v-model="showCreateDialog" title="新增社区老人" width="520px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="身份证号" required>
          <el-input v-model="createForm.id_card" />
        </el-form-item>
        <el-form-item label="姓名" required>
          <el-input v-model="createForm.name" />
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="createForm.phone" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="createForm.address" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createElder" :loading="createLoading">创建</el-button>
      </template>
    </el-dialog>

    <!-- Welfare Dialog -->
    <el-dialog v-model="showWelfareDialog" title="福利标签管理" width="480px">
      <div v-if="currentElder">
        <div style="margin-bottom:12px;font-weight:600">{{ currentElder.name }} — 福利标签</div>
        <el-table :data="currentWelfareTags" size="small" stripe>
          <el-table-column prop="tag_code" label="标签代码" width="150" />
          <el-table-column prop="valid_from" label="生效日期" width="120" />
          <el-table-column prop="valid_to" label="到期日期" width="120" />
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button link type="danger" size="small" @click="revokeTag(row.tag_code)">撤销</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="showWelfareDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { communityApi, personApi } from '@/api/business-chains'

interface CommunityElder {
  id: string
  id_card: string
  name: string
  phone?: string
  address?: string
  welfare_tags?: string[]
  status: string
  last_signin?: string
  created_at: string
}

const loading = ref(false)
const elders = ref<CommunityElder[]>([])
const kpis = ref({ total_elders: 0, today_signin: 0, welfare_tags: 0, pending_payments: 0 })

const filters = reactive({ status: '', search: '' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const showCreateDialog = ref(false)
const createLoading = ref(false)
const createForm = reactive({ id_card: '', name: '', phone: '', address: '' })

const showWelfareDialog = ref(false)
const currentElder = ref<CommunityElder | null>(null)
const currentWelfareTags = ref<any[]>([])

const parseTags = (json: string | null | undefined): string[] => {
  if (!json) return []
  try { return JSON.parse(json) } catch { return [] }
}

const fetchElders = async () => {
  loading.value = true
  try {
    const res: any = await communityApi.listElders({
      page: pagination.page,
      page_size: pagination.pageSize,
      status: filters.status,
    })
    elders.value = res.data || []
    pagination.total = elders.value.length
    kpis.value.total_elders = elders.value.length
    kpis.value.today_signin = elders.value.filter((e: CommunityElder) => e.last_signin).length
    kpis.value.pending_payments = 5
  } catch {
    elders.value = []
  } finally {
    loading.value = false
  }
}

const createElder = async () => {
  if (!createForm.id_card || !createForm.name) {
    ElMessage.warning('请填写身份证号和姓名')
    return
  }
  createLoading.value = true
  try {
    await communityApi.createElder(createForm)
    ElMessage.success('创建成功')
    showCreateDialog.value = false
    Object.assign(createForm, { id_card: '', name: '', phone: '', address: '' })
    await fetchElders()
  } catch (e: any) {
    ElMessage.error(e.message || '创建失败')
  } finally {
    createLoading.value = false
  }
}

const triggerSignin = (row: CommunityElder) => {
  communityApi.signin(row.id, { type: 'welfare' })
    .then(() => ElMessage.success('签到成功'))
    .catch(() => ElMessage.error('签到失败'))
}

const viewDetail = (row: CommunityElder) => {
  ElMessage.info(`查看 ${row.name} 详情`)
}

const assignWelfare = (row: CommunityElder) => {
  currentElder.value = row
  currentWelfareTags.value = []
  showWelfareDialog.value = true
  communityApi.getStats(row.id).then((res: any) => {
    currentWelfareTags.value = res.data?.welfare_tags || []
  }).catch(() => {})
}

const revokeTag = (tagCode: string) => {
  if (!currentElder.value) return
  personApi.revokeWelfareTag(currentElder.value.id, tagCode)
    .then(() => {
      ElMessage.success('已撤销')
      currentWelfareTags.value = currentWelfareTags.value.filter((t: any) => t.tag_code !== tagCode)
    })
    .catch(() => ElMessage.error('撤销失败'))
}

const refreshAll = () => fetchElders()

onMounted(() => {
  fetchElders()
})
</script>

<style scoped>
.community-chain-page { padding: 0; }
.community-chain-page :deep(.el-card) {
  border-radius: 12px !important;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.06) !important;
  transition: all var(--duration-normal) var(--easing-out);
}
.community-chain-page :deep(.el-card:hover) {
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.06), 0 12px 32px rgba(0,0,0,0.08) !important;
  transform: translateY(-1px);
}
.card-header { display: flex; justify-content: space-between; align-items: center; }
.filter-form { margin-bottom: 16px; }
.mr-1 { margin-right: 4px; }
.text-muted { color: var(--el-text-color-placeholder); }
.kpi-card {
  position: relative;
  overflow: hidden;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}
.kpi-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: radial-gradient(ellipse at top left, rgba(255,255,255,0.6) 0%, transparent 60%);
  pointer-events: none;
}
.kpi-card:hover { transform: translateY(-3px); }
.kpi-card :deep(.el-card__body) { padding: 18px; display: flex; flex-direction: column; align-items: center; text-align: center; border-radius: 14px; }
.kpi-value { font-size: 32px; font-weight: 800; letter-spacing: -0.03em; line-height: 1; margin-bottom: 4px; }
.kpi-label { font-size: 12px; color: var(--el-text-color-secondary); margin-top: 6px; font-weight: 600; }
.kpi-blue .kpi-value { color: #5C8D73; }
.kpi-green .kpi-value { color: #6FAF8F; }
.kpi-orange .kpi-value { color: #D9A441; }
.kpi-purple .kpi-value { color: #7BAF8C; }
</style>
