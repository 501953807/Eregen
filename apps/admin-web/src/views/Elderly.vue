<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElNotification } from 'element-plus'
import { elderlyApi } from '@/api/elderly'
import { handleApiError, handleApiSuccess } from '@/utils/error'
import ElderlyTable from './Elderly/ElderlyTable.vue'
import ElderDetailDialog from './Elderly/ElderDetailDialog.vue'
import ElderEditDialog from './Elderly/ElderEditDialog.vue'
import WelfareStats from './Elderly/WelfareStats.vue'

const activePage = ref('elderly')
const pageTitles: Record<string, string> = {
  elderly: '老人档案管理',
  welfare: '福利标签管理',
  signin: '签到总览',
  pharmacy: '药房发药记录',
  minzheng: '民政数据导入',
  stats: '统计看板',
}

function switchPage(page: string) { activePage.value = page }

// Elderly table state
const loading = ref({ elderly: false })
const elderlySearch = ref('')
const page = ref(1)
const pageSize = ref(20)
const kpis = ref({ total: 482, wearable: 312, welfareTags: 9, todaySignin: 28, pendingReview: 5, alerts: 3 })

interface ElderlyRow {
  id: string; name: string; id_card?: string; birth_date?: string; gender?: string
  emergency_contact?: string; welfare_tags: { code: string; name: string; issuer?: string; start_date?: string; end_date?: string }[]
  status: string; address?: string; wearable_id?: string; wearable_online?: boolean
}

const elderlyList = ref<ElderlyRow[]>([
  { id: '1', name: '张秀兰', id_card: '510101195001011234', gender: '女', birth_date: '1950-01-01', emergency_contact: '张明（子）', welfare_tags: [{ code: 'orphan', name: '孤寡' }, { code: 'poverty_1', name: '特困一级' }, { code: 'disability_2', name: '残疾二级' }], status: '正常' },
  { id: '2', name: '李建国', id_card: '510101194805055678', gender: '男', birth_date: '1948-05-05', emergency_contact: '李华（女）', welfare_tags: [{ code: 'special_disease', name: '特病门诊' }, { code: 'bus_discount', name: '公交优惠' }], status: '正常' },
  { id: '3', name: '王秀英', id_card: '510101195503127890', gender: '女', birth_date: '1955-03-12', emergency_contact: '王芳（女）', welfare_tags: [{ code: 'medical_assist', name: '医疗救助' }], status: '停用' },
  { id: '4', name: '赵德柱', id_card: '510101194208153456', gender: '男', birth_date: '1942-08-15', emergency_contact: '赵强（子）', welfare_tags: [{ code: 'orphan', name: '孤寡' }, { code: 'poverty_1', name: '特困一级' }, { code: 'special_disease', name: '特病门诊' }], status: '正常' },
  { id: '5', name: '刘美华', id_card: '510101195812256789', gender: '女', birth_date: '1958-12-25', emergency_contact: '刘晓（女）', welfare_tags: [{ code: 'disability_3', name: '残疾三级' }, { code: 'bus_discount', name: '公交优惠' }], status: '正常' },
])

const filteredElderly = computed(() => {
  if (!elderlySearch.value) return elderlyList.value
  const q = elderlySearch.value.toLowerCase()
  return elderlyList.value.filter(e => e.name.toLowerCase().includes(q) || (e.id_card && e.id_card.includes(q)))
})

// Detail dialog
const showDetailDialog = ref(false)
const detailElder = ref<ElderlyRow | null>(null)
function openDetail(row: ElderlyRow) { detailElder.value = row; showDetailDialog.value = true }

// Edit dialog
const showEditDialog = ref(false)
const editingElder = ref<ElderlyRow | null>(null)
const editForm = ref({ name: '', id_card: '', birth_date: '', gender: '', emergency_contact: '', address: '', status: '正常' })

function openEdit(row: ElderlyRow) {
  editingElder.value = row
  editForm.value = { name: row.name || '', id_card: row.id_card || '', birth_date: row.birth_date || '', gender: row.gender || '', emergency_contact: row.emergency_contact || '', address: row.address || '', status: row.status || '正常' }
  showEditDialog.value = true
  elderlyApi.detail(row.id).then((res: any) => {
    if (res.data) {
      editForm.value.name = res.data.name || editForm.value.name
      editForm.value.id_card = res.data.id_card || editForm.value.id_card
      editForm.value.birth_date = res.data.birth_date || editForm.value.birth_date
      editForm.value.gender = res.data.gender || editForm.value.gender
      editForm.value.emergency_contact = res.data.emergency_contact || editForm.value.emergency_contact
      editForm.value.address = res.data.address || editForm.value.address
      editForm.value.status = res.data.status || '正常'
    }
  }).catch(() => {})
}

async function saveElderly(form: typeof editForm.value) {
  if (!editingElder.value) return
  try {
    await elderlyApi.update(editingElder.value.id, form as any)
    handleApiSuccess('档案保存成功')
    const index = elderlyList.value.findIndex(e => e.id === editingElder.value?.id)
    if (index >= 0) {
      elderlyList.value[index] = { ...elderlyList.value[index], ...form }
    }
    showEditDialog.value = false
    editingElder.value = null
  } catch (e) {
    handleApiError(e, '保存失败，请重试')
  }
}

function addWelfareTag() {
  ElNotification.info({ title: '提示', message: '添加福利标签功能正在开发中', duration: 3000 })
}

async function toggleWelfare(row: any, enabled: boolean) {
  row.enabled = enabled
  handleApiSuccess(`福利标签"${enabled ? '已启用' : '已禁用'}"`)
}

function viewBoundElders(row: any) {
  const bound = elderlyList.value.filter(elder => elder.welfare_tags?.some(tag => tag.code === row.code))
  let msg = `标签"${row.name}"共绑定 ${bound.length} 位老人：\n`
  bound.forEach(e => { msg += `• ${e.name} (${(e.id_card || '').slice(-4)}\n` })
  ElNotification.info({ title: '结果', message: msg, duration: 5000, showClose: true })
}

onMounted(async () => {
  loading.value.elderly = true
  try {
    const res = await elderlyApi.list()
    if (res.data?.data) {
      elderlyList.value = res.data.data.map((item: any) => ({
        id: item.id || '', name: item.name || '', id_card: item.id_card || '',
        birth_date: item.birth_date || '', gender: item.gender || '',
        emergency_contact: item.emergency_contact || '', address: item.address || '',
        welfare_tags: item.welfare_tags || [], status: item.status || '正常',
        wearable_id: item.wearable_id || '', wearable_online: item.wearable_online || false,
      }))
      kpis.value.total = elderlyList.value.length
      kpis.value.wearable = elderlyList.value.filter(e => e.wearable_online).length
      kpis.value.welfareTags = elderlyList.value.flatMap(e => e.welfare_tags).length
    }
  } catch { /* mock data fallback */ }
  finally { loading.value.elderly = false }
})
</script>

<template>
  <div>
    <div class="page-header">
      <el-breadcrumb separator="/">
        <el-breadcrumb-item>首页</el-breadcrumb-item>
        <el-breadcrumb-item>社区老人专区</el-breadcrumb-item>
        <el-breadcrumb-item>{{ pageTitles[activePage] }}</el-breadcrumb-item>
      </el-breadcrumb>
      <h2 class="page-title">{{ pageTitles[activePage] }}</h2>
    </div>

    <el-tabs v-model="activePage" type="border-card" @change="switchPage">
      <el-tab-pane label="老人管理" name="elderly">
        <ElderlyTable
          :rows="filteredElderly"
          :loading="loading.elderly"
          :total="kpis.total"
          :page="page"
          :page-size="pageSize"
          :kpis="kpis"
          @update:search="elderlySearch = $event"
          @add="showEditDialog = true"
          @detail="openDetail"
          @edit="openEdit"
          @page-change="page = $event"
        />
      </el-tab-pane>
      <el-tab-pane label="福利标签" name="welfare">
        <WelfareStats active-page="welfare" @toggle-welfare="toggleWelfare" @view-bound="viewBoundElders" />
      </el-tab-pane>
      <el-tab-pane label="签到总览" name="signin">
        <WelfareStats active-page="signin" />
      </el-tab-pane>
      <el-tab-pane label="药房发药" name="pharmacy">
        <WelfareStats active-page="pharmacy" />
      </el-tab-pane>
      <el-tab-pane label="民政数据" name="minzheng">
        <WelfareStats active-page="minzheng" />
      </el-tab-pane>
      <el-tab-pane label="统计看板" name="stats">
        <WelfareStats active-page="stats" />
      </el-tab-pane>
    </el-tabs>

    <ElderDetailDialog v-model="showDetailDialog" :row="detailElder" @edit="openEdit" />
    <ElderEditDialog v-model="showEditDialog" :row="editingElder" :initial-form="editForm" @save="saveElderly" @add-tag="addWelfareTag" />
  </div>
</template>

<style scoped>
.elderly-page { padding: 0; }
.elderly-page :deep(.el-card) {
  border-radius: 12px !important;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.06) !important;
  transition: all var(--duration-normal) var(--easing-out);
}
.elderly-page :deep(.el-card:hover) {
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.06), 0 12px 32px rgba(0,0,0,0.08) !important;
  transform: translateY(-1px);
}
.page-header { margin-bottom: 20px; }
.page-title { font-size: 22px; font-weight: 800; color: var(--el-text-color-primary); margin: 8px 0 0; }
</style>
