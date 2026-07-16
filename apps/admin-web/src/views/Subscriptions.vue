<template>
  <div class="subscriptions-page">
    <!-- Overview Stats -->
    <el-row :gutter="20" style="margin-bottom: 24px;">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value">2,156</div>
            <div class="stat-label">订阅总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #67C23A;">1,842</div>
            <div class="stat-label">活跃订阅</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #E6A23C;">128</div>
            <div class="stat-label">即将到期</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-value" style="color: #F56C6C;">186</div>
            <div class="stat-label">已过期</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Subscription Type Breakdown -->
    <el-card shadow="hover" style="margin-bottom: 20px;">
      <template #header><span style="font-weight: 600;">订阅类型分布</span></template>
      <el-row :gutter="20">
        <el-col :span="8" v-for="item in subTypes" :key="item.name">
          <div class="sub-type-card">
            <div class="sub-type-header">
              <span class="sub-type-name">{{ item.name }}</span>
              <el-tag :type="item.tagType" size="small">{{ item.price }}</el-tag>
            </div>
            <div class="sub-type-count">{{ item.count }} 用户</div>
            <el-progress
              :percentage="Math.round((item.count / 2156) * 100)"
              :stroke-width="8"
              :color="item.color"
              style="margin-top: 8px;"
            />
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- Renewal Records Table -->
    <el-card shadow="hover">
      <template #header>
        <div class="table-header">
          <span style="font-weight: 600;">续费记录</span>
          <el-button size="small">导出报表</el-button>
        </div>
      </template>
      <el-table :data="renewals" stripe style="width: 100%">
        <el-table-column prop="user" label="家属用户" width="120" />
        <el-table-column prop="elderly" label="关联老人" width="120" />
        <el-table-column prop="plan" label="套餐" width="100">
          <template #default="{ row }">
            <el-tag :type="row.planTag" size="small">{{ row.plan }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="amount" label="金额" width="100">
          <template #default="{ row }">¥{{ row.amount }}</template>
        </el-table-column>
        <el-table-column prop="method" label="支付方式" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.statusTag" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="date" label="日期" width="160" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button link type="primary" size="small">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="display: flex; justify-content: flex-end; margin-top: 16px;">
        <el-pagination background layout="prev, pager, next" :total="500" :page-size="20" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface SubType {
  name: string
  price: string
  count: number
  tagType: 'primary' | 'success' | 'warning'
  color: string[]
}

const subTypes: SubType[] = [
  { name: '基础版', price: '¥19/月', count: 1200, tagType: 'primary', color: ['#4A90D9', '#4A90D9'] },
  { name: '专业版', price: '¥39/月', count: 642, tagType: 'success', color: ['#67C23A', '#67C23A'] },
  { name: '企业版', price: '¥99/月', count: 314, tagType: 'warning', color: ['#E6A23C', '#E6A23C'] },
]

interface Renewal {
  user: string
  elderly: string
  plan: string
  planTag: 'primary' | 'success' | 'warning'
  amount: string
  method: string
  status: string
  statusTag: 'success' | 'warning' | 'danger'
  date: string
}

const renewals: Renewal[] = [
  { user: '张伟', elderly: '张建国', plan: '专业版', planTag: 'success', amount: '39.00', method: '微信支付', status: '成功', statusTag: 'success', date: '2026-07-15' },
  { user: '李芳', elderly: '李秀英', plan: '基础版', planTag: 'primary', amount: '19.00', method: '支付宝', status: '成功', statusTag: 'success', date: '2026-07-14' },
  { user: '王磊', elderly: '王德明', plan: '专业版', planTag: 'success', amount: '39.00', method: '微信支付', status: '失败', statusTag: 'danger', date: '2026-07-14' },
  { user: '赵敏', elderly: '赵淑华', plan: '企业版', planTag: 'warning', amount: '99.00', method: '银行转账', status: '成功', statusTag: 'success', date: '2026-07-13' },
  { user: '陈刚', elderly: '陈志强', plan: '基础版', planTag: 'primary', amount: '19.00', method: '微信支付', status: '已取消', statusTag: 'warning', date: '2026-07-12' },
]
</script>

<style scoped>
.stat-card :deep(.el-card__body) { padding: 20px; display: flex; align-items: center; justify-content: space-between; }
.stat-content { flex: 1; }
.stat-value { font-size: 32px; font-weight: 700; color: #303133; }
.stat-label { font-size: 13px; color: #909399; margin-top: 4px; }
.sub-type-card { padding: 16px; background: #fafafa; border-radius: 8px; }
.sub-type-header { display: flex; justify-content: space-between; align-items: center; }
.sub-type-name { font-weight: 600; font-size: 15px; }
.sub-type-count { font-size: 13px; color: #909399; margin-top: 4px; }
.table-header { display: flex; justify-content: space-between; align-items: center; }
</style>
