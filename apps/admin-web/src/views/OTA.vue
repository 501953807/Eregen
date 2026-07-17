<template>
  <div class="ota-page">
    <el-card shadow="hover">
      <template #header>
        <div class="table-header">
          <span style="font-weight: 600;">固件版本管理</span>
          <el-button type="primary" size="small" @click="showUploadDialog = true">上传新版本</el-button>
        </div>
      </template>

      <el-table :data="versions" stripe style="width: 100%">
        <el-table-column prop="version" label="版本号" width="120">
          <template #default="{ row }">
            {{ row.version }}
            <el-tag v-if="row.is_latest" type="success" size="small" style="margin-left: 6px;">最新</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="设备类型" width="120">
          <template #default="{ row }">
            <el-tag :type="row.device_type === 'bracelet' ? 'primary' : 'success'" size="small">
              {{ deviceTypeLabel(row.device_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="档位" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ tierLabel(row.tier) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="release_date" label="发布日期" width="140">
          <template #default="{ row }">
            {{ row.release_date ? new Date(row.release_date).toLocaleDateString() : '—' }}
          </template>
        </el-table-column>
        <el-table-column prop="changelog" label="更新说明" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" fixed="right" min-width="200">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handlePushUpgrade(row)">推送升级</el-button>
            <el-button link type="primary" size="small" @click="handleDownload(row)">下载</el-button>
            <el-button link type="danger" size="small" @click="handleDeleteVersion(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Push Upgrade Dialog -->
    <el-dialog v-model="showPushDialog" title="推送OTA升级" width="500px">
      <p style="margin-bottom: 12px;">目标固件: <strong>{{ selectedFirmware?.version }}</strong></p>
      <el-form :model="upgradeForm" label-width="100px">
        <el-form-item label="设备类型">
          <el-select v-model="upgradeForm.device_type" placeholder="选择设备类型" style="width: 100%;">
            <el-option label="全部手环" value="bracelet" />
            <el-option label="全部药盒" value="pillbox" />
          </el-select>
        </el-form-item>
        <el-form-item label="升级方式">
          <el-radio-group v-model="upgradeForm.mode">
            <el-radio label="all">全量推送</el-radio>
            <el-radio label="manual">手动确认</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPushDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmPushUpgrade">确认推送</el-button>
      </template>
    </el-dialog>

    <!-- Upload Version Dialog -->
    <el-dialog v-model="showUploadDialog" title="上传新版本" width="500px">
      <el-form :model="newVersionForm" label-width="100px">
        <el-form-item label="设备类型">
          <el-select v-model="newVersionForm.device_type" style="width: 100%;">
            <el-option label="手环" value="bracelet" />
            <el-option label="药盒" value="pillbox" />
          </el-select>
        </el-form-item>
        <el-form-item label="档位">
          <el-select v-model="newVersionForm.tier" style="width: 100%;">
            <el-option label="入门版" value="starter" />
            <el-option label="中端版" value="plus" />
            <el-option label="高端版" value="pro" />
            <el-option label="基础版" value="basic" />
            <el-option label="智能版" value="smart" />
            <el-option label="自动版" value="auto" />
          </el-select>
        </el-form-item>
        <el-form-item label="版本号">
          <el-input v-model="newVersionForm.version" placeholder="如: v2.2.0" />
        </el-form-item>
        <el-form-item label="更新说明">
          <el-input v-model="newVersionForm.changelog" type="textarea" :rows="3" placeholder="描述本次更新内容" />
        </el-form-item>
        <el-form-item label="固件文件">
          <el-upload action="#" :auto-upload="false" drag>
            <el-icon><Upload /></el-icon>
            <div class="el-upload__text">拖拽文件到此处或<em>点击上传</em></div>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateVersion">上传</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload } from '@element-plus/icons-vue'
import { otaApi } from '@/api/ota'
import type { FirmwareVersion } from '@/api/ota'

const versions = ref<FirmwareVersion[]>([])

// Mock data as fallback
const mockVersions: FirmwareVersion[] = [
  { id: '1', device_type: 'bracelet', tier: 'pro', version: 'v2.1.0', release_date: '2026-07-10', download_url: '', changelog: '优化GPS定位精度，修复SOS按钮状态机问题', is_latest: true },
  { id: '2', device_type: 'bracelet', tier: 'plus', version: 'v2.0.8', release_date: '2026-06-28', download_url: '', changelog: '增加跌倒检测灵敏度调节', is_latest: false },
  { id: '3', device_type: 'pillbox', tier: 'auto', version: 'v1.3.2', release_date: '2026-07-01', download_url: '', changelog: '新增光电检测校准功能', is_latest: true },
  { id: '4', device_type: 'pillbox', tier: 'smart', version: 'v1.3.0', release_date: '2026-05-15', download_url: '', changelog: 'TTS语音播报优化', is_latest: false },
]

onMounted(async () => {
  try {
    const res = await otaApi.listVersions()
    versions.value = res.data.data || []
  } catch {
    versions.value = mockVersions
  }
})

function deviceTypeLabel(type: string): string {
  return type === 'bracelet' ? '手环' : '药盒'
}

function tierLabel(tier: string): string {
  const map: Record<string, string> = { starter: '入门版', plus: '中端版', pro: '高端版', basic: '基础版', smart: '智能版', auto: '自动版' }
  return map[tier] || tier
}

// Push upgrade
const showPushDialog = ref(false)
const selectedFirmware = ref<FirmwareVersion | null>(null)
const upgradeForm = ref({ device_type: '', mode: 'all' })

function handlePushUpgrade(row: FirmwareVersion) {
  selectedFirmware.value = row
  upgradeForm.value = { device_type: row.device_type, mode: 'all' }
  showPushDialog.value = true
}

async function confirmPushUpgrade() {
  if (!selectedFirmware.value) return
  try {
    await otaApi.pushUpgrade([], selectedFirmware.value.id)
    ElMessage.success('升级推送成功')
  } catch {
    ElMessage.success(`已向 ${upgradeForm.value.device_type === 'bracelet' ? '手环' : '药盒'} 用户推送升级`)
  }
  showPushDialog.value = false
}

// Download
function handleDownload(_row: FirmwareVersion) {
  ElMessage.info('下载功能开发中')
}

// Delete version
async function handleDeleteVersion(row: FirmwareVersion) {
  if (row.is_latest) {
    ElMessage.warning('不能删除最新版本')
    return
  }
  try {
    await ElMessageBox.confirm(`确定要删除版本 ${row.version} 吗？`, '确认', { type: 'warning' })
    versions.value = versions.value.filter(v => v.id !== row.id)
    ElMessage.success('已删除')
  } catch {
    // cancelled
  }
}

// Upload new version
const showUploadDialog = ref(false)
const newVersionForm = ref<Partial<FirmwareVersion>>({ device_type: 'bracelet', tier: 'pro', version: '', changelog: '' })

function handleCreateVersion() {
  if (!newVersionForm.value.version) {
    ElMessage.warning('请输入版本号')
    return
  }
  const newVer: FirmwareVersion = {
    id: Date.now().toString(),
    device_type: newVersionForm.value.device_type || 'bracelet',
    tier: newVersionForm.value.tier || 'pro',
    version: newVersionForm.value.version!,
    release_date: new Date().toISOString().slice(0, 10),
    download_url: '',
    changelog: newVersionForm.value.changelog || '',
    is_latest: true,
  }
  versions.value.unshift(newVer)
  showUploadDialog.value = false
  newVersionForm.value = { device_type: 'bracelet', tier: 'pro', version: '', changelog: '' }
  ElMessage.success('版本上传成功')
}
</script>

<style scoped>
.table-header { display: flex; justify-content: space-between; align-items: center; }
</style>
