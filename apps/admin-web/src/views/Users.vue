<template>
  <div class="users-page">
    <!-- Tabs for user types -->
    <el-card shadow="hover" style="margin-bottom: 20px;">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="家属用户" name="family">
          <div class="tab-toolbar">
            <el-input v-model="familySearch" placeholder="搜索用户名/手机号" clearable style="width: 240px;" prefix-icon="Search" />
            <el-button type="primary" style="margin-left: 12px;" @click="handleFamilySearch">查询</el-button>
            <el-button type="success" @click="handleAddFamily">添加用户</el-button>
          </div>
          <el-table v-loading="usersStore.loading" :data="usersStore.familyUsers" stripe style="width: 100%; margin-top: 16px;">
            <el-table-column prop="name" label="姓名" width="120" />
            <el-table-column prop="phone" label="手机号" width="140" />
            <el-table-column prop="role" label="角色" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ roleLabel(row.role) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="注册时间" width="160">
              <template #default="{ row }">
                {{ row.created_at ? new Date(row.created_at).toLocaleDateString() : '—' }}
              </template>
            </el-table-column>
            <el-table-column label="操作" fixed="right" min-width="160">
              <template #default="{ row }">
                <el-button link type="primary" size="small">编辑</el-button>
                <el-button link type="primary" size="small">权限</el-button>
                <el-button link type="danger" size="small">禁用</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="老人档案" name="elderly">
          <div class="tab-toolbar">
            <el-input v-model="elderlySearch" placeholder="搜索姓名/设备ID" clearable style="width: 240px;" prefix-icon="Search" />
            <el-button type="primary" style="margin-left: 12px;" @click="handleElderlySearch">查询</el-button>
          </div>
          <el-table v-loading="usersStore.loading" :data="elderlyDisplayData" stripe style="width: 100%; margin-top: 16px;">
            <el-table-column prop="name" label="姓名" width="120" />
            <el-table-column label="年龄" width="80">
              <template #default="{ row }">
                {{ row.birth_date ? calculateAge(row.birth_date) : '—' }}
              </template>
            </el-table-column>
            <el-table-column prop="user_id" label="关联用户" width="120" />
            <el-table-column label="健康等级" width="120">
              <template #default="{ row }">
                <el-tag v-if="row.health_tiers?.length" size="small">{{ row.health_tiers[0] }}</el-tag>
                <span v-else>—</span>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="注册日期" width="160">
              <template #default="{ row }">
                {{ row.created_at ? new Date(row.created_at).toLocaleDateString() : '—' }}
              </template>
            </el-table-column>
            <el-table-column label="操作" fixed="right" min-width="120">
              <template #default="{ row }">
                <el-button link type="primary" size="small">查看详情</el-button>
                <el-button link type="primary" size="small">编辑</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="机构管理" name="institution">
          <div class="tab-toolbar">
            <el-button type="success" @click="handleAddInstitution">添加机构</el-button>
          </div>
          <el-empty description="机构管理功能开发中" :image-size="100" />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useUsersStore } from '@/stores/users'
import type { User, ElderlyProfile } from '@/types'

const usersStore = useUsersStore()
const activeTab = ref('family')
const familySearch = ref('')
const elderlySearch = ref('')

function calculateAge(birthDate: string): number {
  const today = new Date()
  const birth = new Date(birthDate)
  let age = today.getFullYear() - birth.getFullYear()
  if (today.getMonth() < birth.getMonth() || (today.getMonth() === birth.getMonth() && today.getDate() < birth.getDate())) age--
  return age
}

function roleLabel(role: string): string {
  const map: Record<string, string> = { family: '家属', elderly: '老人', institution: '机构', admin: '管理员' }
  return map[role] || role
}

const elderlyDisplayData = ref<Array<ElderlyProfile & { created_at?: string; updated_at?: string }>>([])

async function handleFamilySearch() {
  await usersStore.fetchFamily({ page_size: 50 })
  if (familySearch.value) {
    ElMessage.info(`搜索: ${familySearch.value}`)
  }
}

async function handleElderlySearch() {
  await usersStore.fetchElderly({ page_size: 50 })
  elderlyDisplayData.value = usersStore.elderlyProfiles as any
  if (elderlySearch.value) {
    ElMessage.info(`搜索: ${elderlySearch.value}`)
  }
}

function handleAddFamily() {
  ElMessage.info('添加家属用户功能开发中')
}

function handleAddInstitution() {
  ElMessage.info('添加机构功能开发中')
}

onMounted(async () => {
  await usersStore.fetchFamily({ page_size: 50 })
})
</script>

<style scoped>
.tab-toolbar { display: flex; align-items: center; }
:deep(.el-tabs--border-card) { border: none; box-shadow: none; }
:deep(.el-tabs--border-card > .el-tabs__header) { background: #fafafa; border-bottom: 1px solid #e8e8e8; }
</style>
