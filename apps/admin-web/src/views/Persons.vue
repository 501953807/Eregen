<template>
  <div class="persons-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>统一人员管理</span>
          <el-button type="primary" @click="showCreateDialog = true">新增人员</el-button>
        </div>
      </template>
      
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="业务链">
          <el-select v-model="filters.business_chain" placeholder="全部" clearable>
            <el-option label="自营链" value="self" />
            <el-option label="住院链" value="hospital" />
            <el-option label="社区链" value="community" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable>
            <el-option label="活跃" value="active" />
            <el-option label="暂停" value="suspended" />
            <el-option label="已故" value="deceased" />
          </el-select>
        </el-form-item>
        <el-form-item label="搜索">
          <el-input v-model="filters.search" placeholder="姓名/身份证号" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchPersons">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="persons" v-loading="loading" stripe>
        <el-table-column prop="id_card" label="身份证号" width="180" />
        <el-table-column prop="name" label="姓名" width="100" />
        <el-table-column prop="gender" label="性别" width="60">
          <template #default="{ row }">
            {{ row.gender === 1 ? '男' : row.gender === 2 ? '女' : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="电话" width="120" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : row.status === 'suspended' ? 'warning' : 'danger'">
              {{ row.status === 'active' ? '活跃' : row.status === 'suspended' ? '暂停' : '已故' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="业务链" width="150">
          <template #default="{ row }">
            <el-tag v-for="chain in getChains(row.id)" :key="chain" size="small" class="mr-1">
              {{ chain === 'self' ? '自营' : chain === 'hospital' ? '住院' : '社区' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" fixed="right" width="150">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row)">详情</el-button>
            <el-button link type="primary" @click="editPerson(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @change="fetchPersons"
      />
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="showCreateDialog" :title="editingPerson ? '编辑人员' : '新增人员'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="身份证号" required>
          <el-input v-model="form.id_card" :disabled="!!editingPerson" />
        </el-form-item>
        <el-form-item label="姓名" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="性别">
          <el-select v-model="form.gender">
            <el-option label="男" :value="1" />
            <el-option label="女" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="出生日期">
          <el-date-picker v-model="form.birth_date" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="紧急联系人">
          <el-input v-model="form.emergency_contact" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="form.address" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="savePerson" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { personApi } from '@/api/business-chains'
import type { Person } from '@/types'

const loading = ref(false)
const saving = ref(false)
const persons = ref<Person[]>([])
const showCreateDialog = ref(false)
const editingPerson = ref<Person | null>(null)

const filters = reactive({
  business_chain: '',
  status: '',
  search: '',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const form = reactive({
  id_card: '',
  name: '',
  gender: 0 as 0 | 1 | 2,
  birth_date: '',
  phone: '',
  emergency_contact: '',
  address: '',
})

const chainMap: Record<string, string> = {
  self: '自营',
  hospital: '住院',
  community: '社区',
}

const getChains = (personId: string) => {
  // In production, this would fetch from person_profiles
  return ['self'] as string[]
}

const fetchPersons = async () => {
  loading.value = true
  try {
    const res = await personApi.list({
      page: pagination.page,
      page_size: pagination.pageSize,
      business_chain: filters.business_chain,
      status: filters.status,
    })
    if (res.data) {
      persons.value = (res as unknown as { data: Person[]; page: number; page_size: number }).data
    }
  } catch (e) {
    console.error('Failed to fetch persons:', e)
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.business_chain = ''
  filters.status = ''
  filters.search = ''
  pagination.page = 1
  fetchPersons()
}

const viewDetail = (person: Person) => {
  // Navigate to detail page
  console.log('View detail:', person)
}

const editPerson = (person: Person) => {
  editingPerson.value = person
  form.id_card = person.id_card
  form.name = person.name
  form.gender = person.gender
  form.birth_date = person.birth_date || ''
  form.phone = person.phone || ''
  form.emergency_contact = person.emergency_contact || ''
  form.address = person.address || ''
  showCreateDialog.value = true
}

const savePerson = async () => {
  if (!form.id_card || !form.name) {
    ElMessage.warning('请填写必填项')
    return
  }
  saving.value = true
  try {
    if (editingPerson.value) {
      await personApi.update(editingPerson.value.id, form)
      ElMessage.success('更新成功')
    } else {
      await personApi.create(form)
      ElMessage.success('创建成功')
    }
    showCreateDialog.value = false
    editingPerson.value = null
    resetForm()
    fetchPersons()
  } catch (e) {
    console.error('Failed to save person:', e)
  } finally {
    saving.value = false
  }
}

const resetForm = () => {
  form.id_card = ''
  form.name = ''
  form.gender = 0
  form.birth_date = ''
  form.phone = ''
  form.emergency_contact = ''
  form.address = ''
}

onMounted(() => {
  fetchPersons()
})
</script>

<style scoped>
.persons-page { padding: 0; }
.persons-page :deep(.el-card) {
  border-radius: 12px !important;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 1px 3px rgba(0,0,0,0.04), 0 4px 16px rgba(0,0,0,0.06) !important;
  transition: all var(--duration-normal) var(--easing-out);
}
.persons-page :deep(.el-card:hover) {
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.8), 0 2px 8px rgba(0,0,0,0.04), 0 4px 12px rgba(0,0,0,0.06), 0 12px 32px rgba(0,0,0,0.08) !important;
  transform: translateY(-1px);
}
.card-header { display: flex; justify-content: space-between; align-items: center; }
.filter-form { margin-bottom: 16px; }
.mr-1 { margin-right: 4px; }

/* Responsive */
@media (max-width: 768px) {
  .persons-page :deep(.el-table) { font-size: 12px; }
  .persons-page :deep(.el-table th),
  .persons-page :deep(.el-table td) { padding: 6px 4px; }
}
</style>
