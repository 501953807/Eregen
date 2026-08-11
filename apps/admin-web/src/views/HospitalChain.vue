<template>
  <div class="hospital-chain-page">
    <el-card>
      <template #header>
        <span>住院链管理</span>
      </template>
      <el-alert type="warning" :closable="false" show-icon>
        <template #title>住院患者管理</template>
        <template #default>管理住院患者的腕带绑定、巡检查房、用药执行</template>
      </el-alert>
      <el-table :data="patients" v-loading="loading" stripe>
        <el-table-column prop="admission_no" label="入院号" />
        <el-table-column prop="name" label="姓名" />
        <el-table-column prop="department" label="科室" />
        <el-table-column prop="bed_number" label="床位" />
        <el-table-column prop="status" label="状态" />
        <el-table-column label="操作">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { hospitalApi } from '@/api/business-chains'

const loading = ref(false)
const patients = ref<any[]>([])

const fetchPatients = async () => {
  loading.value = true
  try {
    const res: any = await hospitalApi.listPatients()
    if (res?.data) {
      patients.value = res.data
    }
  } catch (e) {
    console.error('Failed to fetch patients:', e)
  } finally {
    loading.value = false
  }
}

const viewDetail = (row: any) => {
  console.log('View detail:', row)
}

onMounted(() => {
  fetchPatients()
})
</script>
