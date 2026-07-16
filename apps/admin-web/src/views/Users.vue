<template>
  <div class="users-page">
    <!-- Tabs for user types -->
    <el-card shadow="hover" style="margin-bottom: 20px;">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="家属用户" name="family">
          <div class="tab-toolbar">
            <el-input v-model="familySearch" placeholder="搜索用户名/手机号" clearable style="width: 240px;" prefix-icon="Search" />
            <el-button type="primary" style="margin-left: 12px;">查询</el-button>
            <el-button type="success">添加用户</el-button>
          </div>
          <el-table :data="familyUsers" stripe style="width: 100%; margin-top: 16px;">
            <el-table-column prop="name" label="姓名" width="120" />
            <el-table-column prop="phone" label="手机号" width="140" />
            <el-table-column prop="elderlyCount" label="关联老人" width="100" />
            <el-table-column prop="plan" label="订阅套餐" width="120">
              <template #default="{ row }">
                <el-tag :type="row.planTag" size="small">{{ row.plan }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="registered" label="注册时间" width="160" />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.statusTag" size="small">{{ row.status }}</el-tag>
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
            <el-button type="primary" style="margin-left: 12px;">查询</el-button>
          </div>
          <el-table :data="elderlyUsers" stripe style="width: 100%; margin-top: 16px;">
            <el-table-column prop="name" label="姓名" width="120" />
            <el-table-column prop="age" label="年龄" width="80" />
            <el-table-column prop="gender" label="性别" width="80" />
            <el-table-column prop="device" label="设备" width="120" />
            <el-table-column prop="healthLevel" label="健康等级" width="120">
              <template #default="{ row }">
                <el-tag :type="row.healthTag" size="small">{{ row.healthLevel }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="family" label="绑定家属" width="120" />
            <el-table-column prop="registered" label="注册日期" width="160" />
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
            <el-button type="success">添加机构</el-button>
          </div>
          <el-table :data="institutions" stripe style="width: 100%; margin-top: 16px;">
            <el-table-column prop="name" label="机构名称" width="200" />
            <el-table-column prop="type" label="类型" width="120" />
            <el-table-column prop="beds" label="床位" width="80" />
            <el-table-column prop="devices" label="设备数" width="100" />
            <el-table-column prop="users" label="用户数" width="100" />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.statusTag" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" fixed="right" min-width="120">
              <template #default="{ row }">
                <el-button link type="primary" size="small">管理</el-button>
                <el-button link type="primary" size="small">设置</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const activeTab = ref('family')
const familySearch = ref('')
const elderlySearch = ref('')

interface FamilyUser {
  name: string
  phone: string
  elderlyCount: number
  plan: string
  planTag: 'primary' | 'success' | 'warning'
  registered: string
  status: string
  statusTag: 'success' | 'danger'
}

const familyUsers: FamilyUser[] = [
  { name: '张伟', phone: '138****5678', elderlyCount: 1, plan: '专业版', planTag: 'success', registered: '2025-03-15', status: '正常', statusTag: 'success' },
  { name: '李芳', phone: '139****1234', elderlyCount: 2, plan: '基础版', planTag: 'primary', registered: '2025-05-20', status: '正常', statusTag: 'success' },
  { name: '王磊', phone: '137****9876', elderlyCount: 1, plan: '专业版', planTag: 'success', registered: '2025-08-10', status: '禁用', statusTag: 'danger' },
  { name: '赵敏', phone: '136****4321', elderlyCount: 3, plan: '企业版', planTag: 'warning', registered: '2025-01-05', status: '正常', statusTag: 'success' },
]

interface ElderlyUser {
  name: string
  age: number
  gender: string
  device: string
  healthLevel: string
  healthTag: 'success' | 'warning' | 'danger'
  family: string
  registered: string
}

const elderlyUsers: ElderlyUser[] = [
  { name: '张建国', age: 78, gender: '男', device: 'BR-0042', healthLevel: '低风险', healthTag: 'success', family: '张伟', registered: '2025-03-20' },
  { name: '李秀英', age: 82, gender: '女', device: 'BR-0017', healthLevel: '中风险', healthTag: 'warning', family: '李芳', registered: '2025-06-01' },
  { name: '王德明', age: 75, gender: '男', device: 'BR-0089', healthLevel: '高风险', healthTag: 'danger', family: '王磊', registered: '2025-09-15' },
]

interface Institution {
  name: string
  type: string
  beds: number
  devices: number
  users: number
  status: string
  statusTag: 'success' | 'info'
}

const institutions: Institution[] = [
  { name: '阳光养老院', type: '养老院', beds: 200, devices: 45, users: 45, status: '运营中', statusTag: 'success' },
  { name: '康健社区中心', type: '社区中心', beds: 50, devices: 12, users: 12, status: '运营中', statusTag: 'success' },
  { name: '夕阳红护理院', type: '护理院', beds: 150, devices: 30, users: 30, status: '待接入', statusTag: 'info' },
]
</script>

<style scoped>
.tab-toolbar { display: flex; align-items: center; }
:deep(.el-tabs--border-card) { border: none; box-shadow: none; }
:deep(.el-tabs--border-card > .el-tabs__header) { background: #fafafa; border-bottom: 1px solid #e8e8e8; }
</style>
