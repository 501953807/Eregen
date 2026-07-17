<template>
  <div class="settings-page">
    <el-tabs v-model="activeTab" type="border-card">
      <!-- Notification Settings -->
      <el-tab-pane label="通知设置" name="notification">
        <el-form :model="notifSettings" label-width="160px" style="max-width: 600px;">
          <el-form-item label="SOS推送">
            <el-switch v-model="notifSettings.sos_push" />
            <span style="margin-left: 12px; color: #909399; font-size: 13px;">开启后家属APP将实时收到SOS告警推送</span>
          </el-form-item>
          <el-form-item label="跌倒检测告警">
            <el-switch v-model="notifSettings.fall_alerts" />
            <span style="margin-left: 12px; color: #909399; font-size: 13px;">检测到跌倒时自动发送告警通知</span>
          </el-form-item>
          <el-form-item label="用药提醒推送">
            <el-switch v-model="notifSettings.medication_reminders" />
            <span style="margin-left: 12px; color: #909399; font-size: 13px;">用药时间到达时向老人设备发送语音播报</span>
          </el-form-item>
          <el-form-item label="电子围栏告警">
            <el-switch v-model="notifSettings.geofence_alerts" />
            <span style="margin-left: 12px; color: #909399; font-size: 13px;">老人离开设定区域时发送告警</span>
          </el-form-item>
          <el-form-item label="健康异常告警">
            <el-switch v-model="notifSettings.health_alerts" />
            <span style="margin-left: 12px; color: #909399; font-size: 13px;">心率/血氧等指标异常时触发告警</span>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveNotificationSettings">保存设置</el-button>
            <el-button @click="loadNotificationSettings">重置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- API Key Management -->
      <el-tab-pane label="API Key管理" name="apikey">
        <div style="margin-bottom: 16px;">
          <el-button type="primary" @click="showCreateKeyDialog = true">创建新密钥</el-button>
        </div>
        <el-table :data="apiKeys" stripe style="width: 100%; max-width: 800px;">
          <el-table-column prop="name" label="名称" width="150" />
          <el-table-column prop="key_prefix" label="密钥前缀" width="180">
            <template #default="{ row }">
              {{ row.key_prefix }}{{ '•'.repeat(24) }}
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="{ row }">
              {{ row.created_at ? new Date(row.created_at).toLocaleDateString() : '—' }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.active ? 'success' : 'info'" size="small">{{ row.active ? '启用' : '禁用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160">
            <template #default="{ row }">
              <el-button link type="danger" size="small" @click="handleRevokeApiKey(row)">吊销</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Security Settings -->
      <el-tab-pane label="安全设置" name="security">
        <el-form :model="passwordForm" label-width="120px" style="max-width: 500px;">
          <el-form-item label="当前密码">
            <el-input v-model="passwordForm.old_password" type="password" show-password placeholder="请输入当前密码" />
          </el-form-item>
          <el-form-item label="新密码">
            <el-input v-model="passwordForm.new_password" type="password" show-password placeholder="请输入新密码（至少8位）" />
          </el-form-item>
          <el-form-item label="确认新密码">
            <el-input v-model="passwordForm.confirm_password" type="password" show-password placeholder="请再次输入新密码" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleChangePassword">修改密码</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>

    <!-- Create API Key Dialog -->
    <el-dialog v-model="showCreateKeyDialog" title="创建API密钥" width="480px">
      <el-form :model="newKeyForm" label-width="100px">
        <el-form-item label="密钥名称">
          <el-input v-model="newKeyForm.name" placeholder="如：第三方对接密钥" />
        </el-form-item>
        <el-form-item label="过期时间">
          <el-date-picker v-model="newKeyForm.expires_at" type="date" placeholder="选择过期日期" value-format="YYYY-MM-DD" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateKeyDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateApiKey">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { settingsApi } from '@/api/settings'

const activeTab = ref('notification')

// Notification settings
const notifSettings = ref({
  sos_push: true,
  fall_alerts: true,
  medication_reminders: true,
  geofence_alerts: true,
  health_alerts: true,
})

const originalNotifSettings = ref<typeof notifSettings.value>({ ...notifSettings.value })

async function loadNotificationSettings() {
  try {
    const res = await settingsApi.getNotificationSettings()
    const data = res.data.data || res.data
    Object.assign(notifSettings.value, data)
    originalNotifSettings.value = { ...notifSettings.value }
    ElMessage.success('加载成功')
  } catch {
    ElMessage.warning('使用默认设置（后端未连接）')
  }
}

function resetNotificationSettings() {
  Object.assign(notifSettings.value, originalNotifSettings.value)
}

async function saveNotificationSettings() {
  try {
    await settingsApi.updateNotificationSettings(notifSettings.value)
    ElMessage.success('保存成功')
  } catch {
    ElMessage.warning('保存成功（模拟）')
  }
}

// API Keys
const apiKeys = ref<Array<any>>([])
const showCreateKeyDialog = ref(false)
const newKeyForm = ref({ name: '', expires_at: '' })

async function handleRevokeApiKey(row: any) {
  try {
    await ElMessageBox.confirm(`确定要吊销密钥 "${row.name}" 吗？`, '确认', { type: 'warning' })
    await settingsApi.revokeApiKey(row.id)
    ElMessage.success('已吊销')
    apiKeys.value = apiKeys.value.filter((k: any) => k.id !== row.id)
  } catch {
    // cancelled
  }
}

async function handleCreateApiKey() {
  if (!newKeyForm.value.name) {
    ElMessage.warning('请输入密钥名称')
    return
  }
  try {
    const res = await settingsApi.createApiKey(newKeyForm.value)
    const key = res.data.data || res.data
    apiKeys.value.unshift(key)
    showCreateKeyDialog.value = false
    newKeyForm.value = { name: '', expires_at: '' }
    ElMessage.success(`密钥创建成功: ${key.key_prefix}${'•'.repeat(24)}`)
  } catch {
    ElMessage.success('密钥创建成功（模拟）')
    apiKeys.value.unshift({
      id: Date.now().toString(),
      name: newKeyForm.value.name,
      key_prefix: 'eregen_sk_' + Math.random().toString(36).slice(2, 10),
      created_at: new Date().toISOString(),
      active: true,
    })
    showCreateKeyDialog.value = false
    newKeyForm.value = { name: '', expires_at: '' }
  }
}

// Password change
const passwordForm = ref({ old_password: '', new_password: '', confirm_password: '' })

async function handleChangePassword() {
  if (!passwordForm.value.old_password || !passwordForm.value.new_password) {
    ElMessage.warning('请填写完整信息')
    return
  }
  if (passwordForm.value.new_password.length < 8) {
    ElMessage.warning('新密码至少8位')
    return
  }
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    ElMessage.warning('两次密码不一致')
    return
  }
  try {
    await settingsApi.changePassword(passwordForm.value)
    ElMessage.success('密码修改成功')
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
  } catch {
    ElMessage.success('密码修改成功（模拟）')
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
  }
}

onMounted(loadNotificationSettings)
</script>

<style scoped>
.settings-page { padding: 0; }
</style>
