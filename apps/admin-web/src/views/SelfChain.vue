<template>
  <div class="self-chain-page">
    <el-card>
      <template #header>
        <span>自营链管理</span>
      </template>
      <el-alert type="info" :closable="false" show-icon>
        <template #title>自营链人员管理</template>
        <template #default>管理佩戴Eregen手环/药盒的自营用户</template>
      </el-alert>
      <el-table :data="elderlyList" v-loading="loading" stripe>
        <el-table-column prop="name" label="姓名" />
        <el-table-column prop="tier" label="订阅层级" />
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
import { selfApi } from '@/api/business-chains'

const loading = ref(false)
const elderlyList = ref<any[]>([])

const fetchElderly = async () => {
  loading.value = true
  try {
    const res: any = await selfApi.listElderly()
    if (res?.data) {
      elderlyList.value = res.data
    }
  } catch (e) {
    console.error('Failed to fetch elderly:', e)
  } finally {
    loading.value = false
  }
}

const viewDetail = (row: any) => {
  console.log('View detail:', row)
}

onMounted(() => {
  fetchElderly()
})
</script>
