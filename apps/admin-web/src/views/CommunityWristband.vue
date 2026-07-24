<template>
  <div class="community-page">
    <!-- ===== KPI Cards ===== -->
    <el-row :gutter="12" style="margin-bottom: 16px;">
      <el-col :span="4">
        <el-card shadow="hover" class="kpi-card kpi-blue">
          <div class="kpi-value">{{ stats.total_elders }}</div>
          <div class="kpi-label">登记老人</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="kpi-card kpi-green">
          <div class="kpi-value">{{ stats.active_devices }}</div>
          <div class="kpi-label">在线腕带</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="kpi-card kpi-purple">
          <div class="kpi-value">{{ stats.welfare_tags_count }}</div>
          <div class="kpi-label">福利标签</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="kpi-card kpi-warning">
          <div class="kpi-value">{{ todaySignin }}</div>
          <div class="kpi-label">今日签到</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="kpi-card kpi-danger">
          <div class="kpi-value">{{ minzhengPending }}</div>
          <div class="kpi-label">待审核民政</div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="hover" class="kpi-card kpi-orange">
          <div class="kpi-value">{{ pendingAlerts }}</div>
          <div class="kpi-label">异常告警</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- ===== Tabs ===== -->
    <el-tabs v-model="activeTab" type="border-card">

      <!-- Tab 1: Elder Management -->
      <el-tab-pane label="老人管理" name="elders">
        <el-row :gutter="16" style="margin-bottom: 16px;">
          <el-col :span="3">
            <el-button type="primary" @click="showAddElder = true; editingElder = null; resetElderForm()">新增老人</el-button>
          </el-col>
          <el-col :span="7">
            <el-input v-model="elderSearch" placeholder="搜索姓名 / 身份证号 / 手机号" clearable />
          </el-col>
        </el-row>

        <el-table :data="filteredElders" v-loading="loading.elders" stripe class="elder-table">
          <el-table-column prop="name" label="姓名" width="90">
            <template #default="{ row }">
              <div class="patient-cell">
                <div class="patient-avatar" :class="row.gender === 1 ? 'avatar-blue' : 'avatar-pink'">{{ row.name[0] || '?' }}</div>
                <strong>{{ row.name }}</strong>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="id_card" label="身份证号" width="190">
            <template #default="{ row }"><span class="mono">{{ row.id_card }}</span></template>
          </el-table-column>
          <el-table-column prop="age" label="年龄" width="55" />
          <el-table-column prop="gender" label="性别" width="50">
            <template #default="{ row }">{{ row.gender === 1 ? '男' : row.gender === 2 ? '女' : '-' }}</template>
          </el-table-column>
          <el-table-column prop="emergency_contact" label="紧急联系人" width="120" />
          <el-table-column prop="welfare_tags" label="福利标签" min-width="180">
            <template #default="{ row }">
              <el-tag v-for="tag in (row._welfareTags || [])" :key="tag.tag_code" size="small" :class="'welfare-tag-' + welfareTagClass(tag.tag_code)" effect="light" style="margin-right: 4px;">
                {{ tag.tag_name }}
              </el-tag>
              <span v-if="!row._welfareTags?.length" style="color:var(--el-text-color-placeholder);">无</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="70">
            <template #default="{ row }">
              <span class="status-badge" :class="row.status === 'active' ? 'badge-success' : 'badge-gray'">
                <span class="status-dot" :class="row.status === 'active' ? 'dot-success' : 'dot-gray'"></span>
                {{ row.status === 'active' ? '正常' : '停用' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="primary" link @click="viewElderDetail(row)">详情</el-button>
              <el-button size="small" link @click="editElder(row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Tab 2: Welfare Tags -->
      <el-tab-pane label="福利标签" name="welfare">
        <el-row :gutter="16" style="margin-bottom: 16px;">
          <el-col :span="3">
            <el-button type="primary" @click="showTagDialog = true">＋ 新增标签</el-button>
          </el-col>
          <el-col :span="3">
            <el-button>批量分配</el-button>
          </el-col>
        </el-row>

        <el-table :data="welfareTags" v-loading="loading.welfare" stripe>
          <el-table-column prop="tag_code" label="标签代码" width="160">
            <template #default="{ row }"><span class="mono">{{ row.tag_code }}</span></template>
          </el-table-column>
          <el-table-column prop="tag_name" label="标签名称" width="120" />
          <el-table-column prop="issuer" label="发放机构" width="100" />
          <el-table-column prop="renewal_period_days" label="Renewal 周期" width="110">
            <template #default="{ row }">{{ row.renewal_period_days }} 天</template>
          </el-table-column>
          <el-table-column prop="benefit_amount" label="补助金额" width="100">
            <template #default="{ row }">{{ row.benefit_amount > 0 ? '¥' + row.benefit_amount : '¥0' }}</template>
          </el-table-column>
          <el-table-column label="绑定老人" width="90" align="center">
            <template #default="{ row }">
              <span style="text-align:center;display:inline-block;width:100%;">{{ countByTag(row.tag_code) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="启用" width="60" align="center">
            <template #default="{ row }">
              <span class="enabled-dot"></span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="primary" link @click="viewTagElders(row.tag_code, row.tag_name)">查看绑定</el-button>
              <el-button size="small" link>编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Tab 3: Sign-in Overview -->
      <el-tab-pane label="签到总览" name="signin">
        <el-row :gutter="16" style="margin-bottom: 16px;">
          <el-col :span="5">
            <el-date-picker v-model="signinPeriod" type="month" placeholder="选择月份" value-format="YYYY-MM" style="width: 100%;" />
          </el-col>
          <el-col :span="5">
            <el-select v-model="signinHospital" placeholder="医院筛选" clearable style="width: 100%;">
              <el-option label="全部医院" value="" />
              <el-option label="社区医院 A" value="hospital-a" />
              <el-option label="社区医院 B" value="hospital-b" />
              <el-option label="社区医院 C" value="hospital-c" />
            </el-select>
          </el-col>
          <el-col :span="3">
            <el-button type="primary" @click="loadSigninRecords">🔍 查询</el-button>
          </el-col>
        </el-row>

        <!-- CSS Bar Chart -->
        <el-card shadow="never" style="margin-bottom: 20px;">
          <template #header><span class="section-title">近 7 天签到趋势</span></template>
          <div class="bar-chart">
            <div class="bar-col" v-for="(day, i) in weekSigninData" :key="i">
              <div class="bar-value">{{ day.count }}</div>
              <div class="bar" :style="{ height: (day.count / maxWeekCount * 100) + 'px', minHeight: '4px' }"></div>
              <div class="bar-label">{{ day.label }}</div>
            </div>
          </div>
        </el-card>

        <el-table :data="signinRecords" v-loading="loading.signin" stripe>
          <el-table-column prop="elder_name" label="老人姓名" width="100">
            <template #default="{ row }"><strong>{{ row.elder_name || row.elder_id }}</strong></template>
          </el-table-column>
          <el-table-column label="身份证号" width="190">
            <template #default="{ row }"><span class="mono">{{ (row.elder_id || '').slice(-4) }}</span></template>
          </el-table-column>
          <el-table-column prop="hospital_id" label="医院" width="120" />
          <el-table-column prop="signin_time" label="签到时间" width="180" />
          <el-table-column prop="activated_tags" label="激活标签" min-width="200">
            <template #default="{ row }">
              <el-tag v-for="t in parseTags(row.activated_tags)" :key="t" size="small" style="margin-right: 4px;">{{ t }}</el-tag>
              <span v-if="!parseTags(row.activated_tags)?.length" style="color:var(--el-text-color-placeholder);">—</span>
            </template>
          </el-table-column>
          <el-table-column label="类型" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="row.is_welfare_signin ? 'badge-success' : 'badge-primary'">
                <span class="status-dot" :class="row.is_welfare_signin ? 'dot-success' : 'dot-primary'"></span>
                {{ row.is_welfare_signin ? '福利签到' : '医保签到' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Tab 4: Pharmacy Records -->
      <el-tab-pane label="药房发药" name="pharmacy">
        <el-row :gutter="16" style="margin-bottom: 16px;">
          <el-col :span="3">
            <el-button type="primary">＋ 手动发药</el-button>
          </el-col>
          <el-col :span="5">
            <el-date-picker v-model="pharmacyMonth" type="month" placeholder="选择月份" value-format="YYYY-MM" style="width: 100%;" />
          </el-col>
          <el-col :span="6">
            <el-input v-model="pharmacySearch" placeholder="搜索老人姓名 / 药品名" clearable />
          </el-col>
        </el-row>

        <el-table :data="pharmacyLogs" v-loading="loading.pharmacy" stripe>
          <el-table-column label="日期" width="70">
            <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
          </el-table-column>
          <el-table-column prop="elder_name" label="老人姓名" width="100">
            <template #default="{ row }"><strong>{{ row.elder_name || row.elder_id }}</strong></template>
          </el-table-column>
          <el-table-column prop="hospital_id" label="医院" width="110" />
          <el-table-column prop="items" label="药品清单" min-width="180">
            <template #default="{ row }">{{ parseItems(row.items).join('、') }}</template>
          </el-table-column>
          <el-table-column label="金额" width="80">
            <template #default="{ row }">¥{{ row.total_cost?.toFixed(2) || '0.00' }}</template>
          </el-table-column>
          <el-table-column prop="pharmacist_id" label="药师/护士" width="100" />
          <el-table-column label="签到状态" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="row.signed_in ? 'badge-success' : 'badge-warning'">
                <span class="status-dot" :class="row.signed_in ? 'dot-success' : 'dot-warning'"></span>
                {{ row.signed_in ? '已签到' : '未签到' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Tab 5: Minzheng Import -->
      <el-tab-pane label="民政数据" name="minzheng">
        <el-row :gutter="16" style="margin-bottom: 20px;">
          <el-col :span="12">
            <el-card shadow="never">
              <template #header><span class="section-title">上传 CSV / XLSX 文件</span></template>
              <div class="upload-zone" @click="triggerFileUpload">
                <div style="font-size:36px;margin-bottom:8px;">📁</div>
                <p>点击或拖拽文件到此处上传</p>
                <p style="font-size:11px;margin-top:4px;color:var(--el-text-color-placeholder);">支持民政局标准模板或自定义模板</p>
              </div>
              <input ref="fileInput" type="file" accept=".csv,.xlsx" style="display:none" @change="handleFileUpload" />
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card shadow="never">
              <template #header><span class="section-title">CSV 字段说明</span></template>
              <table class="template-table">
                <thead><tr><th>列名</th><th>必填</th><th>说明</th><th>示例</th></tr></thead>
                <tbody>
                  <tr><td>姓名</td><td><el-tag type="danger" size="small">是</el-tag></td><td>老人姓名</td><td>张秀兰</td></tr>
                  <tr><td>身份证号</td><td><el-tag type="danger" size="small">是</el-tag></td><td>18 位身份证号码</td><td>510101195001011234</td></tr>
                  <tr><td>福利类型</td><td><el-tag type="danger" size="small">是</el-tag></td><td>orphan / poverty_level_1 / ...</td><td>特困</td></tr>
                  <tr><td>认定等级</td><td><el-tag type="danger" size="small">是</el-tag></td><td>1 / 2 / 3</td><td>一级</td></tr>
                  <tr><td>有效期开始</td><td><el-tag type="danger" size="small">是</el-tag></td><td>YYYY-MM-DD</td><td>2025-01-01</td></tr>
                  <tr><td>有效期结束</td><td><el-tag type="danger" size="small">是</el-tag></td><td>YYYY-MM-DD</td><td>2028-12-31</td></tr>
                  <tr><td>备注</td><td><el-tag type="info" size="small">否</el-tag></td><td>额外信息</td><td>肢体残疾</td></tr>
                </tbody>
              </table>
            </el-card>
          </el-col>
        </el-row>

        <el-table :data="minzhengSyncs" v-loading="loading.minzheng" stripe>
          <el-table-column prop="source" label="数据来源" width="140" />
          <el-table-column prop="filename" label="文件名" min-width="140">
            <template #default="{ row }">{{ row.filename || '—' }}</template>
          </el-table-column>
          <el-table-column prop="imported_count" label="导入数" width="80" align="center" />
          <el-table-column prop="matched_count" label="匹配数" width="80" align="center" />
          <el-table-column prop="pending_review_count" label="待审核" width="80" align="center" />
          <el-table-column prop="status" label="状态" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="minzhengStatusClass(row.status)">
                <span class="status-dot" :class="minzhengStatusDot(row.status)"></span>
                {{ statusLabel(row.status) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180" />
        </el-table>
      </el-tab-pane>

      <!-- Tab 6: Batch Payments -->
      <el-tab-pane label="批量发放" name="payments">
        <el-row :gutter="16" style="margin-bottom: 16px;">
          <el-col :span="5">
            <el-select v-model="paymentPeriod" placeholder="选择月份" value-format="YYYY-MM" style="width: 100%;">
              <el-option v-for="m in paymentPeriods" :key="m" :label="m" :value="m" />
            </el-select>
          </el-col>
          <el-col :span="3">
            <el-button type="primary" @click="executeBatchPayment">执行发放</el-button>
          </el-col>
        </el-row>

        <el-table :data="batchPayments" v-loading="loading.payments" stripe>
          <el-table-column prop="batch_id" label="批次号" width="170">
            <template #default="{ row }"><span class="mono">{{ row.batch_id }}</span></template>
          </el-table-column>
          <el-table-column prop="period" label="月份" width="100" />
          <el-table-column prop="pay_type" label="发放类型" width="100" />
          <el-table-column prop="amount" label="金额" width="100">
            <template #default="{ row }">{{ row.amount > 0 ? '¥' + row.amount : '—' }}</template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="paymentStatusClass(row.status)">
                <span class="status-dot" :class="paymentStatusDot(row.status)"></span>
                {{ statusLabel(row.status) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="failure_reason" label="失败原因" min-width="150">
            <template #default="{ row }">{{ row.failure_reason || '—' }}</template>
          </el-table-column>
          <el-table-column prop="executed_at" label="执行时间" width="180" />
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- ===== Add/Edit Elder Dialog ===== -->
    <el-dialog v-model="showAddElder" :title="editingElder ? '编辑老人' : '新增老人'" width="640px" destroy-on-close>
      <el-form :model="elderForm" label-width="100px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="姓名"><el-input v-model="elderForm.name" placeholder="请输入姓名" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="性别">
              <el-radio-group v-model="elderForm.gender">
                <el-radio :value="1">男</el-radio>
                <el-radio :value="2">女</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="身份证号"><el-input v-model="elderForm.id_card" placeholder="18 位身份证号码" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="年龄"><el-input-number v-model="elderForm.age" :min="0" :max="150" style="width:100%;" /></el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="地址"><el-input v-model="elderForm.address" type="textarea" :rows="2" placeholder="居住地址" /></el-form-item>
        <el-form-item label="紧急联系人"><el-input v-model="elderForm.emergency_contact" placeholder="姓名 + 电话" /></el-form-item>
        <el-form-item label="所属医院">
          <el-select v-model="elderForm.hospital_id" placeholder="请选择" style="width:100%;">
            <el-option label="社区医院 A" value="hospital-a" />
            <el-option label="社区医院 B" value="hospital-b" />
            <el-option label="社区医院 C" value="hospital-c" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="elderForm.status">
            <el-radio value="active">正常</el-radio>
            <el-radio value="deactivated">停用</el-radio>
            <el-radio value="deceased">deceased</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddElder = false">取消</el-button>
        <el-button type="primary" @click="saveElder">保存</el-button>
      </template>
    </el-dialog>

    <!-- ===== Elder Detail Side Panel ===== -->
    <div class="side-panel-overlay" :class="{ show: showDetailDialog }" @click="showDetailDialog = false" />
    <div class="side-panel" :class="{ open: showDetailDialog }">
      <div class="panel-header">
        <span class="panel-title">老人档案详情</span>
        <button class="panel-close" @click="showDetailDialog = false">&#10005;</button>
      </div>
      <div class="panel-body" v-if="detailElder">
        <div class="patient-detail-header">
          <div class="patient-avatar-large" :class="detailElder.gender === 1 ? 'avatar-blue' : 'avatar-pink'">{{ detailElder.name?.[0] || '?' }}</div>
          <div>
            <div class="patient-detail-name">{{ detailElder.name }}</div>
            <div class="patient-detail-id">
              <span class="mono">{{ detailElder.id_card }}</span>
              <span class="status-badge" :class="detailElder.status === 'active' ? 'badge-success' : 'badge-gray'" style="margin-left: 8px;">
                <span class="status-dot" :class="detailElder.status === 'active' ? 'dot-success' : 'dot-gray'"></span>
                {{ detailElder.status === 'active' ? '正常' : '停用' }}
              </span>
            </div>
          </div>
        </div>

        <div class="info-section">
          <div class="section-title">基本信息</div>
          <div class="panel-row">
            <span class="panel-label">性别</span>
            <span class="panel-value">{{ detailElder.gender === 1 ? '男' : detailElder.gender === 2 ? '女' : '-' }}</span>
          </div>
          <div class="panel-row">
            <span class="panel-label">年龄</span>
            <span class="panel-value">{{ detailElder.age || '—' }} 岁</span>
          </div>
          <div class="panel-row">
            <span class="panel-label">地址</span>
            <span class="panel-value">{{ detailElder.address || '—' }}</span>
          </div>
          <div class="panel-row">
            <span class="panel-label">紧急联系人</span>
            <span class="panel-value">{{ detailElder.emergency_contact || '—' }}</span>
          </div>
        </div>

        <div class="info-section">
          <div class="section-title">福利标签</div>
          <div v-loading="loading.detail">
            <div v-for="tag in detailWelfareTags" :key="tag.tag_code" class="welfare-tag-row">
              <el-tag :type="welfareTagType(tag.tag_code)" size="small" effect="light">{{ tag.tag_name }}</el-tag>
              <span class="welfare-tag-issuer">{{ tag.issuer }}</span>
              <span class="welfare-tag-dates">{{ tag.valid_from }} ~ {{ tag.valid_to }}</span>
              <span class="welfare-tag-status" :class="{ expired: !isTagValid(tag) }">
                {{ isTagValid(tag) ? '有效' : '过期' }}
              </span>
            </div>
            <div v-if="!detailWelfareTags.length" style="color:var(--el-text-color-placeholder);font-size:13px;">暂无福利标签</div>
          </div>
        </div>

        <div class="info-section">
          <div class="section-title">最近签到记录</div>
          <div v-for="rec in detailSigninHistory" :key="rec.signin_time" class="signin-record">
            <div class="signin-time">{{ rec.signin_time }}</div>
            <div class="signin-meta">
              <span class="signin-hospital">{{ rec.hospital_id }}</span>
              <span class="status-badge" :class="rec.is_welfare_signin ? 'badge-success' : 'badge-primary'" style="font-size:11px;padding:1px 6px;">
                {{ rec.is_welfare_signin ? '福利' : '医保' }}
              </span>
            </div>
            <div class="signin-tags" v-if="rec.activated_tags">
              <el-tag v-for="t in parseTags(rec.activated_tags)" :key="t" size="small" style="margin: 2px 4px 2px 0;">{{ t }}</el-tag>
            </div>
          </div>
          <div v-if="!detailSigninHistory.length" style="color:var(--el-text-color-placeholder);font-size:13px;">暂无签到记录</div>
        </div>
      </div>
    </div>

    <!-- ===== Tag Elders Dialog ===== -->
    <el-dialog v-model="showTagEldersDialog" :title="'绑定老人 — ' + selectedTagName" width="600px" destroy-on-close>
      <el-table :data="tagEldersList" v-loading="loading.tag_elders" stripe size="small">
        <el-table-column prop="name" label="姓名" width="100" />
        <el-table-column prop="id_card" label="身份证号" width="190">
          <template #default="{ row }"><span class="mono">{{ row.id_card }}</span></template>
        </el-table-column>
        <el-table-column prop="valid_to" label="到期日" width="110" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import { communityApi, type CommunityElder, type CommunityDevice, type WelfareTagConfig } from '@/api/community'

const activeTab = ref('elders')

// Stats
const stats = ref({ total_elders: 0, active_devices: 0, welfare_tags_count: 0 })
const todaySignin = ref(0)
const pendingPayments = ref(0)
const minzhengPending = ref(0)
const pendingAlerts = ref(0)

// Elders
const elders = ref<CommunityElder[]>([])
const elderSearch = ref('')
const showAddElder = ref(false)
const editingElder = ref<CommunityElder | null>(null)
const elderForm = ref<Partial<CommunityElder>>({ gender: 1, status: 'active' })
const loading = ref({ elders: false, devices: false, welfare: false, signin: false, payments: false, minzheng: false, pharmacy: false, detail: false, tag_elders: false })

// Detail panel
const showDetailDialog = ref(false)
const detailElder = ref<CommunityElder | null>(null)
const detailWelfareTags = ref<any[]>([])
const detailSigninHistory = ref<any[]>([])

// Tag elders dialog
const showTagEldersDialog = ref(false)
const selectedTagName = ref('')
const selectedTagCode = ref('')
const tagEldersList = ref<CommunityElder[]>([])

// Devices
const devices = ref<CommunityDevice[]>([])

// Welfare
const welfareTags = ref<WelfareTagConfig[]>([])

// Sign-in
const signinPeriod = ref(dayjs().format('YYYY-MM'))
const signinHospital = ref('')
const signinRecords = ref<any[]>([])
const weekSigninData = ref<{ label: string; count: number }[]>([])
const maxWeekCount = ref(1)

// Pharmacy
const pharmacyMonth = ref(dayjs().format('YYYY-MM'))
const pharmacySearch = ref('')
const pharmacyLogs = ref<any[]>([])

// Payments
const paymentPeriod = ref(dayjs().format('YYYY-MM'))
const batchPayments = ref<any[]>([])
const paymentPeriods = ref<string[]>([dayjs().format('YYYY-MM'), dayjs().subtract(1, 'month').format('YYYY-MM')])

// Minzheng
const minzhengSyncs = ref<any[]>([])

// File upload
const fileInput = ref<HTMLInputElement | null>(null)

const filteredElders = computed(() => {
  if (!elderSearch.value) return elders.value
  const q = elderSearch.value.toLowerCase()
  return elders.value.filter(e =>
    e.name?.toLowerCase().includes(q) ||
    e.id_card?.toLowerCase().includes(q) ||
    e.emergency_contact?.toLowerCase().includes(q)
  )
})

onMounted(() => {
  loadElders()
  loadDevices()
  loadWelfareTags()
  loadSigninRecords()
  loadBatchPayments()
  loadMinzhengSync()
  loadPharmacyLogs()
  generateWeekData()
})

// --- Helpers ---
function welfareTagType(code: string): string {
  const map: Record<string, string> = {
    orphan: 'danger',
    poverty_level_1: 'warning',
    poverty_level_2: 'warning',
    disability_1: 'primary',
    disability_2: 'primary',
    disability_3: 'primary',
    special_disease: 'danger',
    bus_discount: 'success',
    medical_assistance: '',
  }
  return map[code] || ''
}

function welfareTagClass(code: string): string {
  const map: Record<string, string> = {
    orphan: 'tag-orphan',
    poverty_level_1: 'tag-poverty',
    poverty_level_2: 'tag-poverty',
    disability_1: 'tag-disability',
    disability_2: 'tag-disability',
    disability_3: 'tag-disability',
    special_disease: 'tag-special',
    bus_discount: 'tag-bus',
    medical_assistance: 'tag-medical',
  }
  return map[code] || ''
}

function countByTag(code: string): number {
  return elders.value.filter(e =>
    (e as any)._welfareTags?.some((t: any) => t.tag_code === code)
  ).length
}

function parseTags(json: string | undefined): string[] {
  if (!json) return []
  try { return JSON.parse(json) } catch { return [] }
}

function parseItems(json: string | undefined): string[] {
  if (!json) return []
  try { return JSON.parse(json) } catch { return [json] }
}

function isTagValid(tag: any): boolean {
  return tag.valid_to && dayjs(tag.valid_to).isAfter(dayjs())
}

function formatDate(ts: string | undefined): string {
  if (!ts) return '—'
  return ts.slice(5) // MM-DD
}

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    active: '正常', inactive: '离线', retired: '已退役',
    success: '成功', failed: '失败', pending: '待处理', retrying: '重试中',
    completed: '完成', processing: '处理中',
  }
  return map[status] || status
}

function minzhengStatusClass(status: string): string {
  if (status === 'completed') return 'badge-success'
  if (status === 'failed') return 'badge-danger'
  return 'badge-warning'
}

function minzhengStatusDot(status: string): string {
  if (status === 'completed') return 'dot-success'
  if (status === 'failed') return 'dot-danger'
  return 'dot-warning'
}

function paymentStatusClass(status: string): string {
  if (status === 'success') return 'badge-success'
  if (status === 'failed') return 'badge-danger'
  if (status === 'retrying') return 'badge-warning'
  return 'badge-info'
}

function paymentStatusDot(status: string): string {
  if (status === 'success') return 'dot-success'
  if (status === 'failed') return 'dot-danger'
  if (status === 'retrying') return 'dot-warning'
  return 'dot-info'
}

// --- Load functions ---
async function loadElders() {
  loading.value.elders = true
  try {
    const res = await communityApi.listElders({ page: 1, page_size: 50 })
    const list = res.data?.data || []
    elders.value = list
    stats.value.total_elders = list.length
  } finally {
    loading.value.elders = false
  }
}

async function saveElder() {
  try {
    if (editingElder.value?.id) {
      await communityApi.updateElder(editingElder.value.id, elderForm.value)
      ElMessage.success('更新成功')
    } else {
      await communityApi.createElder(elderForm.value as any)
      ElMessage.success('创建成功')
    }
    showAddElder.value = false
    editingElder.value = null
    resetElderForm()
    await loadElders()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

function resetElderForm() {
  elderForm.value = { gender: 1, status: 'active' }
}

function editElder(row: CommunityElder) {
  editingElder.value = row
  elderForm.value = { ...row }
  showAddElder.value = true
}

async function viewElderDetail(row: CommunityElder) {
  detailElder.value = row
  showDetailDialog.value = true
  detailWelfareTags.value = [
    { tag_code: 'orphan', tag_name: '孤寡老人', issuer: '民政局', valid_from: '2025-01-01', valid_to: '2028-12-31' },
    { tag_code: 'poverty_level_1', tag_name: '特困一级', issuer: '民政局', valid_from: '2025-01-01', valid_to: '2028-12-31' },
    { tag_code: 'disability_2', tag_name: '残疾二级', issuer: '残联', valid_from: '2024-06-01', valid_to: '2027-05-31' },
  ]
  detailSigninHistory.value = [
    { signin_time: '2026-07-23 10:30', hospital_id: '社区医院 A', is_welfare_signin: true, activated_tags: '["孤寡","特困一级","残疾二级"]' },
    { signin_time: '2026-07-16 09:15', hospital_id: '社区医院 A', is_welfare_signin: true, activated_tags: '["孤寡","特困一级","残疾二级"]' },
  ]
}

async function loadDevices() {
  loading.value.devices = true
  try {
    const res = await communityApi.listDevices({ page: 1, page_size: 50 })
    devices.value = res.data?.data || []
    stats.value.active_devices = devices.value.filter(d => d.status === 'active').length
  } finally {
    loading.value.devices = false
  }
}

async function loadWelfareTags() {
  loading.value.welfare = true
  try {
    const res = await communityApi.listWelfareTags()
    welfareTags.value = res.data?.data || []
    stats.value.welfare_tags_count = welfareTags.value.length
  } finally {
    loading.value.welfare = false
  }
}

async function viewTagElders(tagCode: string, tagName: string) {
  selectedTagCode.value = tagCode
  selectedTagName.value = tagName
  showTagEldersDialog.value = true
  tagEldersList.value = elders.value.slice(0, 5)
}

async function loadSigninRecords() {
  loading.value.signin = true
  try {
    const period = signinPeriod.value
    const res = await communityApi.listSigninRecords(period ? { period } : undefined)
    signinRecords.value = res.data?.data || []
    todaySignin.value = signinRecords.value.filter((r: any) => r.signin_time?.startsWith(dayjs().format('YYYY-MM-DD'))).length
  } finally {
    loading.value.signin = false
  }
}

function generateWeekData() {
  const days = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  const counts = [42, 56, 48, 41, 62, 28, 19]
  weekSigninData.value = days.map((label, i) => ({ label, count: counts[i] }))
  maxWeekCount.value = Math.max(...counts)
}

async function executeBatchPayment() {
  try {
    const elderIds = elders.value.filter(e => e.status === 'active').map(e => e.id)
    if (!elderIds.length) {
      ElMessage.warning('没有可发放的老人')
      return
    }
    await communityApi.executeBatchPayment({
      batch_id: 'BATCH-' + Date.now(),
      period: paymentPeriod.value,
      pay_type: 'welfare',
      elder_ids: elderIds,
    })
    ElMessage.success('已提交发放')
    await loadBatchPayments()
  } catch {
    ElMessage.error('发放失败')
  }
}

async function loadBatchPayments() {
  loading.value.payments = true
  try {
    const res = await communityApi.listBatchPayments()
    batchPayments.value = res.data?.data || []
  } finally {
    loading.value.payments = false
  }
}

async function loadMinzhengSync() {
  loading.value.minzheng = true
  try {
    const res = await communityApi.listMinzhengSync()
    minzhengSyncs.value = res.data?.data || []
    minzhengPending.value = minzhengSyncs.value.reduce((sum: number, s: any) => sum + (s.pending_review_count || 0), 0)
  } finally {
    loading.value.minzheng = false
  }
}

async function loadPharmacyLogs() {
  loading.value.pharmacy = true
  try {
    pharmacyLogs.value = [
      { elder_id: 'elder-1', elder_name: '张秀兰', hospital_id: '社区医院 A', items: '["氨氯地平","二甲双胍"]', total_cost: 45.50, pharmacist_id: '张护士', signed_in: true, created_at: '2026-07-23T10:30:00Z' },
      { elder_id: 'elder-2', elder_name: '李建国', hospital_id: '社区医院 B', items: '["阿司匹林肠溶片"]', total_cost: 12.00, pharmacist_id: '李药师', signed_in: true, created_at: '2026-07-23T09:15:00Z' },
      { elder_id: 'elder-3', elder_name: '王秀英', hospital_id: '社区医院 A', items: '["硝苯地平缓释片"]', total_cost: 28.00, pharmacist_id: '张护士', signed_in: false, created_at: '2026-07-22T14:20:00Z' },
      { elder_id: 'elder-1', elder_name: '张秀兰', hospital_id: '社区医院 A', items: '["维生素D胶囊"]', total_cost: 35.00, pharmacist_id: '张护士', signed_in: true, created_at: '2026-07-21T16:00:00Z' },
    ]
  } finally {
    loading.value.pharmacy = false
  }
}

function triggerFileUpload() {
  fileInput.value?.click()
}

function handleFileUpload(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    ElMessage.success(`文件 ${file.name} 已选择，正在上传...`)
    target.value = ''
  }
}
</script>

<style scoped>
.community-page {
  padding: 0;
}

/* KPI Cards */
.kpi-card :deep(.el-card__body) {
  padding: 18px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  border-radius: 14px;
}
.kpi-value {
  font-size: 28px;
  font-weight: 800;
  line-height: 1.2;
}
.kpi-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 6px;
  font-weight: 600;
}

.kpi-blue .kpi-value { color: #2563EB; }
.kpi-green .kpi-value { color: #16A34A; }
.kpi-purple .kpi-value { color: #7C3AED; }
.kpi-warning .kpi-value { color: #F59E0B; }
.kpi-danger .kpi-value { color: #EF4444; }
.kpi-orange .kpi-value { color: #EA580C; }

/* Section title */
.section-title {
  font-size: 15px;
  font-weight: 700;
}

/* Bar chart */
.bar-chart {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  padding: 16px 0 8px;
  height: 120px;
}
.bar-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}
.bar {
  width: 36px;
  background: linear-gradient(180deg, #2563EB, #7C3AED);
  border-radius: 3px 3px 0 0;
  transition: height 0.3s;
}
.bar-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-top: 6px;
}
.bar-value {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 4px;
}

/* Upload zone */
.upload-zone {
  border: 2px dashed var(--el-border-color-light);
  border-radius: 4px;
  padding: 32px;
  text-align: center;
  color: var(--el-text-color-placeholder);
  cursor: pointer;
  transition: border-color 0.2s;
}
.upload-zone:hover {
  border-color: #2563EB;
  color: #2563EB;
}

/* Template table */
.template-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}
.template-table th,
.template-table td {
  padding: 8px 12px;
  border: 1px solid var(--el-border-color-light);
  text-align: left;
}
.template-table th {
  background: #fafafa;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

/* Patient cell */
.patient-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}
.patient-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}
.avatar-blue { background: #DBEAFE; color: #2563EB; }
.avatar-pink { background: #FCE7F3; color: #EC4899; }

/* Welfare tags with colored backgrounds */
.welfare-tag-orphan { background: #FEF2F2; color: #DC2626; }
.welfare-tag-poverty { background: #FFFBEB; color: #D97706; }
.welfare-tag-disability { background: #EFF6FF; color: #2563EB; }
.welfare-tag-special { background: #FEF2F2; color: #DC2626; }
.welfare-tag-bus { background: #F0FDF4; color: #16A34A; }
.welfare-tag-medical { background: #EDE9FE; color: #7C3AED; }

/* Enabled dot */
.enabled-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #16A34A;
  display: inline-block;
}

/* Status badges with dots */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
}
.badge-success { background: #F0FDF4; color: #16A34A; }
.badge-danger { background: #FEF2F2; color: #DC2626; }
.badge-warning { background: #FFFBEB; color: #D97706; }
.badge-primary { background: #EFF6FF; color: #2563EB; }
.badge-gray { background: #F3F4F6; color: #6B7280; }
.badge-info { background: #F8FAFC; color: #94A3B8; }
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}
.dot-success { background: #16A34A; }
.dot-danger { background: #DC2626; }
.dot-warning { background: #D97706; }
.dot-primary { background: #2563EB; }
.dot-gray { background: #6B7280; }
.dot-info { background: #94A3B8; }

/* Mono font */
.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
}

/* ========== Elder Detail Side Panel ========== */
.side-panel-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  z-index: 200;
  display: none;
}
.side-panel-overlay.show {
  display: block;
}
.side-panel {
  position: fixed;
  top: 0;
  right: -520px;
  bottom: 0;
  width: 520px;
  background: white;
  z-index: 201;
  transition: right 0.3s ease;
  overflow-y: auto;
  box-shadow: -10px 0 40px rgba(0,0,0,0.1);
}
.side-panel.open {
  right: 0;
}
.panel-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--el-border-color-light);
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  background: white;
  z-index: 1;
}
.panel-title {
  font-size: 15px;
  font-weight: 700;
}
.panel-close {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: var(--el-fill-color-light);
  cursor: pointer;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
}
.panel-close:hover {
  background: var(--el-border-color-light);
}
.panel-body {
  padding: 20px 24px;
}

.patient-detail-header {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--el-border-color-light);
}
.patient-avatar-large {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  flex-shrink: 0;
}
.patient-detail-name {
  font-size: 17px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.patient-detail-id {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.info-section {
  margin-bottom: 20px;
}
.section-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-text-color-regular);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--el-border-color-light);
}
.panel-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
}
.panel-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  font-weight: 500;
}
.panel-value {
  font-size: 13px;
  color: var(--el-text-color-primary);
  font-weight: 600;
}

/* Welfare tag row */
.welfare-tag-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  font-size: 13px;
}
.welfare-tag-issuer {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.welfare-tag-dates {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  margin-left: auto;
}
.welfare-tag-status {
  font-size: 11px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  background: #F0FDF4;
  color: #16A34A;
}
.welfare-tag-status.expired {
  background: #FEF2F2;
  color: #DC2626;
}

/* Signin record */
.signin-record {
  padding: 8px 0;
  border-bottom: 1px solid var(--el-border-color-light);
}
.signin-record:last-child {
  border-bottom: none;
}
.signin-time {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
.signin-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.signin-hospital {
  font-weight: 600;
}
.signin-tags {
  margin-top: 4px;
}
</style>
