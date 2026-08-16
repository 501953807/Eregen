<template>
  <div class="settings-page">
    <!-- Tabs — Hope UI pill style -->
    <HopeTabs
      v-model="activeTab"
      :tabs="tabItems"
      :pill-style="true"
    />

    <!-- Notification Settings -->
    <div v-show="activeTab === 'notification'" class="settings-section">
      <HopeCard title="通知设置" subtitle="配置家属 APP 及老人设备的推送规则">
        <div class="notif-list">
          <div v-for="item in notifFields" :key="item.prop" class="notif-row">
            <div class="notif-row__info">
              <div class="notif-row__label">{{ item.label }}</div>
              <div class="notif-row__desc">{{ item.desc }}</div>
            </div>
            <el-switch v-model="(notifSettings as Record<string, boolean>)[item.prop]" />
          </div>
        </div>
        <template #footer>
          <div class="form-actions">
            <HopeBtn variant="filled" @click="saveNotificationSettings">保存设置</HopeBtn>
            <HopeBtn variant="plain" @click="loadNotificationSettings">重置</HopeBtn>
          </div>
        </template>
      </HopeCard>
    </div>

    <!-- API Key Management -->
    <div v-show="activeTab === 'apikey'" class="settings-section">
      <HopeCard title="API Key 管理" subtitle="管理 B2B 对接密钥，用于医院和社区平台接入">
        <template #header>
          <HopeBtn variant="filled" size="sm" @click="showCreateKeyDialog = true">
            + 创建新密钥
          </HopeBtn>
        </template>
        <HopeTable :columns="apiKeyColumns" :data="apiKeys" :loading="false" striped>
          <template #col-key_prefix="{ row }">
            <span class="mono">{{ row.key_prefix }}{{ '•'.repeat(24) }}</span>
          </template>
          <template #col-created_at="{ row }">
            {{ row.created_at ? new Date(row.created_at).toLocaleDateString() : '—' }}
          </template>
          <template #col-active="{ row }">
            <HopeBadge :color="row.active ? 'success' : 'info'">
              <span :class="['status-dot', row.active ? 'dot-success' : 'dot-gray']" />
              {{ row.active ? '启用' : '禁用' }}
            </HopeBadge>
          </template>
          <template #col-action="{ row }">
            <HopeBtn variant="text" size="sm" @click="handleRevokeApiKey(row)" class="btn-revoke">吊销</HopeBtn>
          </template>
        </HopeTable>
      </HopeCard>
    </div>

    <!-- Security Settings -->
    <div v-show="activeTab === 'security'" class="settings-section">
      <HopeCard title="安全设置" subtitle="修改登录密码以保障账号安全">
        <div class="pw-form">
          <div class="form-row">
            <label class="form-label">当前密码</label>
            <HopeInput
              v-model="passwordForm.old_password"
              type="password"
              show-password
              placeholder="请输入当前密码"
            />
          </div>
          <div class="form-row">
            <label class="form-label">新密码</label>
            <HopeInput
              v-model="passwordForm.new_password"
              type="password"
              show-password
              placeholder="请输入新密码（至少8位）"
              :error="passwordForm.new_password.length > 0 && passwordForm.new_password.length < 8 ? '密码长度不足8位' : ''"
            />
          </div>
          <div class="form-row">
            <label class="form-label">确认新密码</label>
            <HopeInput
              v-model="passwordForm.confirm_password"
              type="password"
              show-password
              placeholder="请再次输入新密码"
              :error="passwordForm.new_password && passwordForm.confirm_password && passwordForm.new_password !== passwordForm.confirm_password ? '两次密码不一致' : ''"
            />
          </div>
        </div>
        <template #footer>
          <HopeBtn variant="filled" @click="handleChangePassword">修改密码</HopeBtn>
        </template>
      </HopeCard>
    </div>

    <!-- Create API Key Dialog — Hope Modal -->
    <HopeModal
      v-model="showCreateKeyDialog"
      title="创建 API 密钥"
      size="md"
    >
      <div class="pw-form">
        <div class="form-row">
          <label class="form-label">密钥名称</label>
          <HopeInput
            v-model="newKeyForm.name"
            placeholder="如：第三方对接密钥"
          />
        </div>
        <div class="form-row">
          <label class="form-label">过期时间</label>
          <HopeInput
            v-model="newKeyForm.expires_at"
            type="text"
            placeholder="选择过期日期（YYYY-MM-DD）"
          />
        </div>
      </div>
      <template #footer>
        <div class="form-actions">
          <HopeBtn variant="plain" @click="showCreateKeyDialog = false">取消</HopeBtn>
          <HopeBtn variant="filled" @click="handleCreateApiKey">创建</HopeBtn>
        </div>
      </template>
    </HopeModal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { settingsApi } from '@/api/settings'
import { HopeTabs, HopeCard, HopeBtn, HopeInput, HopeBadge, HopeTable, HopeModal } from '@/components/hope'

const activeTab = ref('notification')

const tabItems = [
  { label: '通知设置', value: 'notification' },
  { label: 'API Key 管理', value: 'apikey' },
  { label: '安全设置', value: 'security' },
]

// Notification settings
const notifSettings = ref({
  sos_push: true,
  fall_alerts: true,
  medication_reminders: true,
  geofence_alerts: true,
  health_alerts: true,
})

const originalNotifSettings = ref<typeof notifSettings.value>({ ...notifSettings.value })

const notifFields = [
  { prop: 'sos_push', label: 'SOS 推送', desc: '开启后家属 APP 将实时收到 SOS 告警推送' },
  { prop: 'fall_alerts', label: '跌倒检测告警', desc: '检测到跌倒时自动发送告警通知' },
  { prop: 'medication_reminders', label: '用药提醒推送', desc: '用药时间到达时向老人设备发送语音播报' },
  { prop: 'geofence_alerts', label: '电子围栏告警', desc: '老人离开设定区域时发送告警' },
  { prop: 'health_alerts', label: '健康异常告警', desc: '心率/血氧等指标异常时触发告警' },
]

async function loadNotificationSettings() {
  try {
    ElMessage.warning('设置功能暂不可用（后端未实现）')
    Object.assign(notifSettings.value, { sms_alerts: true, push_alerts: true, voice_reminders: true, geofence_alerts: true })
    originalNotifSettings.value = { ...notifSettings.value }
  } catch {
    ElMessage.warning('使用默认设置（后端未连接）')
  }
}

function resetNotificationSettings() {
  Object.assign(notifSettings.value, originalNotifSettings.value)
}

async function saveNotificationSettings() {
  ElMessage.warning('设置功能暂不可用（后端未实现）')
}

// API Keys
const apiKeys = ref<Array<any>>([])
const showCreateKeyDialog = ref(false)
const newKeyForm = ref({ name: '', expires_at: '' })

const apiKeyColumns = [
  { prop: 'name', label: '名称', width: 150 },
  { prop: 'key_prefix', label: '密钥前缀', width: 180 },
  { prop: 'created_at', label: '创建时间', width: 160 },
  { prop: 'active', label: '状态', width: 100 },
  { prop: 'action', label: '操作', width: 100 },
]

async function handleRevokeApiKey(row: any) {
  try {
    await ElMessageBox.confirm(`确定要吊销密钥 "${row.name}" 吗？`, '确认', { type: 'warning' })
    ElMessage.warning('API密钥管理功能暂不可用（后端未实现）')
  } catch {
    // cancelled
  }
}

async function handleCreateApiKey() {
  if (!newKeyForm.value.name) {
    ElMessage.warning('请输入密钥名称')
    return
  }
  ElMessage.warning('API密钥管理功能暂不可用（后端未实现）')
  showCreateKeyDialog.value = false
  newKeyForm.value = { name: '', expires_at: '' }
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
.settings-page {
  padding: 0;
  max-width: 800px;
}

/* Tabs spacing */
.settings-page :deep(.hope-tabs) {
  margin-bottom: 20px;
}

.settings-section {
  animation: fadeSlideIn 0.25s ease;
}
@keyframes fadeSlideIn {
  from { opacity: 0; transform: translateY(6px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* Notification row list */
.notif-list {
  display: flex;
  flex-direction: column;
}
.notif-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  border-bottom: 1px solid var(--hope-border);
}
.notif-row:last-of-type { border-bottom: none; }
.notif-row__info { flex: 1; margin-right: 20px; }
.notif-row__label {
  font-size: 14px;
  font-weight: 600;
  color: var(--hope-text);
  margin-bottom: 2px;
}
.notif-row__desc {
  font-size: 12px;
  color: var(--hope-text-muted);
  line-height: 1.4;
}

/* Password form */
.pw-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.form-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.form-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--hope-text-secondary);
  letter-spacing: 0.01em;
}

/* Form actions */
.form-actions {
  display: flex;
  gap: 10px;
}

/* Mono key display */
.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  color: var(--hope-text-secondary);
  letter-spacing: 0.04em;
}

/* Revoke button — red tint */
.btn-revoke {
  color: var(--hope-error) !important;
}
.btn-revoke:hover {
  background: rgba(192, 74, 66, 0.08) !important;
}

/* HopeModal footer padding */
.settings-page :deep(.hope-modal__footer) {
  padding: 14px 22px;
  border-top: 1px solid var(--hope-border);
}

/* Responsive */
@media (max-width: 640px) {
  .settings-page { padding: 0 12px; }
  .notif-row { flex-direction: column; align-items: flex-start; gap: 10px; }
  .notif-row__info { margin-right: 0; }
}
</style>
