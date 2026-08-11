<template>
  <div class="community-chain-page">
    <el-card>
      <template #header>
        <span>社区链管理</span>
      </template>
      <el-alert type="success" :closable="false" show-icon>
        <template #title>社区老人管理</template>
        <template #default>管理社区认证老人的福利标签、签到、药品发放</template>
      </el-alert>
      <el-table :data="elders" v-loading="loading" stripe>
        <el-table-column prop="name" label="姓名" />
        <el-table-column prop="id_card" label="身份证号" />
        <el-table-column prop="welfare_tags" label="福利标签" />
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
import { communityApi } from '@/api/business-chains'

const loading = ref(false)
const elders = ref<any[]>([])

const fetchElders = async () => {
  loading.value = true
  try {
    const res: any = await communityApi.listElders()
    if (res?.data) {
      elders.value = res.data
    }
  } catch (e) {
    console.error('Failed to fetch elders:', e)
  } finally {
    loading.value = false
  }
}

const viewDetail = (row: any) => {
  console.log('View detail:', row)
}

onMounted(() => {
  fetchElders()
})
</script>
