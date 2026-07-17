<template>
  <div class="institutions-page">
    <el-card shadow="hover">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span style="font-weight: 600; font-size: 16px;">机构管理</span>
          <el-button type="primary" size="default" @click="showDialog = true">
            <el-icon style="margin-right: 4px;"><Plus /></el-icon>新增机构
          </el-button>
        </div>
      </template>

      <!-- Search & Filter -->
      <el-form :inline="true" :model="searchForm" style="margin-bottom: 20px;">
        <el-form-item label="机构名称">
          <el-input v-model="searchForm.name" placeholder="请输入机构名称" clearable style="width: 200px;" />
        </el-form-item>
        <el-form-item label="机构类型">
          <el-select v-model="searchForm.type" placeholder="全部类型" clearable style="width: 140px;">
            <el-option label="医院" value="hospital" />
            <el-option label="社区" value="community" />
            <el-option label="养老院" value="nursing" />
            <el-option label="服务站" value="station" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部状态" clearable style="width: 120px;">
            <el-option label="已激活" value="active" />
            <el-option label="待审核" value="pending" />
            <el-option label="已停用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- Summary Stats -->
      <el-row :gutter="16" style="margin-bottom: 20px;">
        <el-col :span="6">
          <el-statistic title="机构总数" :value="total" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="已激活" :value="activeCount" value-style="color: #67C23A" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="待审核" :value="pendingCount" value-style="color: #E6A23C" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="API密钥生成" :value="apiKeysGenerated" value-style="color: #409EFF" />
        </el-col>
      </el-row>

      <!-- Institution List -->
      <el-table :data="filteredInstitutions" stripe style="width: 100%" v-loading="loading">
        <el-table-column prop="code" label="机构编码" width="130" />
        <el-table-column prop="name" label="机构名称" min-width="200" />
        <el-table-column prop="type" label="类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)" size="small">{{ getTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="region" label="地区" width="140" />
        <el-table-column prop="elderlyLinked" label="关联老人" width="90" align="right" />
        <el-table-column prop="dataPoints" label="数据量" width="100" align="right">
          <template #default="{ row }">{{ formatNumber(row.dataPoints) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : (row.status === 'pending' ? 'warning' : 'info')" size="small">
              {{ row.status === 'active' ? '已激活' : (row.status === 'pending' ? '待审核' : '已停用') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="apiKeyCreated" label="API密钥" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.apiKeyCreated" type="success" size="small" effect="plain">✓</el-tag>
            <el-tag v-else type="info" size="small" effect="plain">—</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="viewDetail(row)">详情</el-button>
            <el-button link type="primary" size="small" @click="toggleApiKey(row)">
              {{ row.apiKeyCreated ? '查看密钥' : '生成密钥' }}
            </el-button>
            <el-button link :type="row.status === 'active' ? 'warning' : 'success'" size="small"
              @click="toggleStatus(row)">
              {{ row.status === 'active' ? '停用' : '启用' }}
            </el-button>
            <el-button link type="danger" size="small" @click="deleteInstitution(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination style="margin-top: 20px; justify-content: center;"
        :current-page="pagination.page" :page-size="pagination.pageSize"
        :total="pagination.total" layout="total, prev, pager, next" @current-change="handlePageChange" />
    </el-card>

    <!-- Add Institution Dialog -->
    <el-dialog v-model="showDialog" title="新增机构" width="500px" destroy-on-close>
      <el-form :model="form" label-width="80px">
        <el-form-item label="机构名称" required>
          <el-input v-model="form.name" placeholder="如：上海市第一中心医院" />
        </el-form-item>
        <el-form-item label="机构编码" required>
          <el-input v-model="form.code" placeholder="如：SH-YXY-001" />
        </el-form-item>
        <el-form-item label="机构类型" required>
          <el-select v-model="form.type" placeholder="请选择类型" style="width: 100%;">
            <el-option label="医院" value="hospital" />
            <el-option label="社区服务中心" value="community" />
            <el-option label="养老院" value="nursing" />
            <el-option label="养老服务站" value="station" />
          </el-select>
        </el-form-item>
        <el-form-item label="所在地区">
          <el-input v-model="form.region" placeholder="如：上海市浦东新区" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="form.contact" placeholder="联系人姓名" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.phone" placeholder="联系电话" />
        </el-form-item>
        <el-form-item label="API密钥">
          <el-switch v-model="form.autoGenerateKey" active-text="自动生成" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAdd">确认添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

interface Institution {
  id: string
  code: string
  name: string
  type: string
  region: string
  contact: string
  phone: string
  status: 'active' | 'pending' | 'inactive'
  elderlyLinked: number
  dataPoints: number
  apiKeyCreated: boolean
}

const loading = ref(false)
const showDialog = ref(false)
const total = ref(128)
const activeCount = ref(96)
const pendingCount = ref(24)
const apiKeysGenerated = ref(82)

const searchForm = ref({ name: '', type: '', status: '' })
const pagination = ref({ page: 1, pageSize: 10, total: 128 })

const form = ref({
  name: '', code: '', type: '', region: '', contact: '', phone: '', autoGenerateKey: true,
})

const institutions = ref<Institution[]>([
  { id: '1', code: 'SH-YXY-001', name: '上海市第一中心医院', type: 'hospital', region: '上海市黄浦区', contact: '张主任', phone: '021-12345678', status: 'active', elderlyLinked: 1250, dataPoints: 45200, apiKeyCreated: true },
  { id: '2', code: 'PD-SQ-001', name: '浦东新区社区服务中心', type: 'community', region: '上海市浦东新区', contact: '李站长', phone: '021-87654321', status: 'active', elderlyLinked: 890, dataPoints: 28300, apiKeyCreated: true },
  { id: '3', code: 'BJ-XIE-001', name: '北京协和医院', type: 'hospital', region: '北京市东城区', contact: '王医生', phone: '010-12345678', status: 'active', elderlyLinked: 780, dataPoints: 24100, apiKeyCreated: true },
  { id: '4', code: 'CY-YLF-001', name: '朝阳区养老服务站', type: 'station', region: '北京市朝阳区', contact: '赵站长', phone: '010-87654321', status: 'pending', elderlyLinked: 420, dataPoints: 12600, apiKeyCreated: false },
  { id: '5', code: 'GZ-YIFU-001', name: '广州医科大学附属第一医院', type: 'hospital', region: '广州市越秀区', contact: '刘主任', phone: '020-12345678', status: 'active', elderlyLinked: 360, dataPoints: 10800, apiKeyCreated: true },
  { id: '6', code: 'SZ-NS-001', name: '深圳市南山区养老院', type: 'nursing', region: '深圳市南山区', contact: '陈院长', phone: '0755-12345678', status: 'inactive', elderlyLinked: 280, dataPoints: 8400, apiKeyCreated: false },
])

const filteredInstitutions = computed(() => {
  let list = institutions.value
  if (searchForm.value.type) list = list.filter(i => i.type === searchForm.value.type)
  if (searchForm.value.status) list = list.filter(i => i.status === searchForm.value.status)
  if (searchForm.value.name) list = list.filter(i => i.name.includes(searchForm.value.name))
  return list
})

function getTypeTagType(type: string): string {
  const map: Record<string, string> = { hospital: '', community: 'success', station: 'warning', nursing: 'info' }
  return map[type] || ''
}

function getTypeLabel(type: string): string {
  const map: Record<string, string> = { hospital: '医院', community: '社区', station: '服务站', nursing: '养老院' }
  return map[type] || type
}

function formatNumber(n: number): string {
  return n >= 10000 ? `${(n / 10000).toFixed(1)}万` : n.toLocaleString()
}

function handleSearch() { /* filter is reactive */ }
function resetSearch() { searchForm.value = { name: '', type: '', status: '' } }
function handlePageChange(_page: number) { /* pagination logic */ }

function viewDetail(row: Institution) {
  ElMessage.info(`查看机构详情: ${row.name}`)
}

function toggleApiKey(row: Institution) {
  if (row.apiKeyCreated) {
    ElMessage.info(`查看机构 ${row.name} 的 API 密钥`)
  } else {
    ElMessageBox.confirm(`为机构 ${row.name} 生成 API 密钥？`, '提示', { type: 'warning' })
      .then(() => {
        row.apiKeyCreated = true
        apiKeysGenerated.value++
        ElMessage.success('API 密钥已生成')
      }).catch(() => {})
  }
}

function toggleStatus(row: Institution) {
  const action = row.status === 'active' ? '停用' : '启用'
  ElMessageBox.confirm(`确定要${action}机构 ${row.name} 吗？`, '提示', { type: 'warning' })
    .then(() => {
      row.status = row.status === 'active' ? 'inactive' : 'active'
      ElMessage.success(`已${action}`)
    }).catch(() => {})
}

function deleteInstitution(row: Institution) {
  ElMessageBox.confirm(`确定要删除机构 ${row.name} 吗？此操作不可恢复。`, '警告', { type: 'error' })
    .then(() => {
      institutions.value = institutions.value.filter(i => i.id !== row.id)
      total.value--
      ElMessage.success('已删除')
    }).catch(() => {})
}

function handleAdd() {
  if (!form.value.name || !form.value.code) {
    ElMessage.warning('请填写必填字段')
    return
  }
  institutions.value.unshift({
    id: String(Date.now()),
    code: form.value.code,
    name: form.value.name,
    type: form.value.type || 'hospital',
    region: form.value.region || '',
    contact: form.value.contact || '',
    phone: form.value.phone || '',
    status: 'pending',
    elderlyLinked: 0,
    dataPoints: 0,
    apiKeyCreated: form.value.autoGenerateKey,
  })
  total.value++
  showDialog.value = false
  ElMessage.success('机构添加成功')
}
</script>

<style scoped>
.institutions-page { padding: 4px 0; }
.el-statistic { text-align: center; }
</style>
