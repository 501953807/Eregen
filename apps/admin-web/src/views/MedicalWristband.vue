<template>
  <div class="medical-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="page-header__left">
        <h1 class="page-title">医疗腕带管理</h1>
        <p class="page-subtitle">护士终端交互 · NFC身份核验 · 入院登记与监管闭环</p>
      </div>
      <div class="page-header__actions">
        <HopeBtn variant="plain" size="md" @click="loadOverview">
          <template #icon>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg>
          </template>
          刷新数据
        </HopeBtn>
      </div>
    </div>

    <!-- KPI Cards — HopeStatCard -->
    <div class="kpi-grid">
      <HopeStatCard
        :value="stats.active_patients"
        label="在院患者"
        icon-color="primary"
        gradient="linear-gradient(135deg, #3a57e8 0%, #6f42c1 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.today_admitted"
        label="今日入院"
        icon-color="success"
        gradient="linear-gradient(135deg, #22c55e 0%, #1aa053 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><line x1="19" y1="8" x2="19" y2="14"/><line x1="22" y1="11" x2="16" y2="11"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="stats.bound_devices"
        label="已绑定腕带"
        icon-color="accent"
        gradient="linear-gradient(135deg, #8C57FF 0%, #6f42c1 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 6v2M8 10l4-4 4 4"/><path d="M8 14h8"/></svg></el-icon>
        </template>
      </HopeStatCard>
      <HopeStatCard
        :value="`${todayStats.matched}/${todayStats.total}`"
        label="今日核验匹配"
        icon-color="warning"
        gradient="linear-gradient(135deg, #FAA938 0%, #f59e0b 100%)"
      >
        <template #icon>
          <el-icon :size="24"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"/></svg></el-icon>
        </template>
      </HopeStatCard>
    </div>

    <!-- Tabs -->
    <HopeTabs
      :model-value="activeTab"
      :tabs="tabItems"
      :animated="true"
      @update:model-value="(v: string | number) => { activeTab = typeof v === 'string' ? v : String(v); }"
    />

    <!-- TAB: Patients -->
    <div v-show="activeTab === 'patients'">
      <!-- Filter Bar — HopeCard -->
      <HopeCard style="margin-bottom: 16px;">
        <template #header>
          <span class="filter-title">患者查询</span>
        </template>
        <div class="filter-row">
          <div class="filter-item">
            <label class="filter-label">住院号</label>
            <el-input v-model="patientForm.admission_no" placeholder="输入住院号" clearable class="hope-input" />
          </div>
          <div class="filter-item">
            <label class="filter-label">姓名</label>
            <el-input v-model="patientForm.name" placeholder="输入姓名" clearable class="hope-input" />
          </div>
          <div class="filter-item--actions">
            <HopeBtn variant="filled" size="sm" @click="searchByAdmission">
              <template #icon>
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              </template>
              查询
            </HopeBtn>
          </div>
        </div>
      </HopeCard>

      <!-- Patients Table — HopeCard wrapping el-table -->
      <HopeCard subtitle="患者列表 · 点击行查看详情">
        <template #header>
          <span class="filter-title">入院登记</span>
        </template>
        <el-table :data="patients" v-loading="loading.patients" stripe style="width: 100%;">
          <el-table-column prop="admission_no" label="住院号" width="140">
            <template #default="{ row }"><span class="mono">{{ row.admission_no }}</span></template>
          </el-table-column>
          <el-table-column prop="name" label="姓名" width="100">
            <template #default="{ row }">
              <div class="patient-cell">
                <HopeAvatar size="sm" :text="row.name?.[0] || '?'" :color="row.gender === '男' ? 'primary' : 'accent'" />
                <strong>{{ row.name }}</strong>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="gender" label="性别" width="60" />
          <el-table-column prop="age" label="年龄" width="60" />
          <el-table-column prop="department" label="科室" width="120" />
          <el-table-column prop="bed_number" label="床号" width="80" />
          <el-table-column prop="blood_type" label="血型" width="60" />
          <el-table-column prop="allergies" label="过敏史" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="90">
            <template #default="{ row }">
              <HopeBadge :color="row.status === 'admitted' ? 'success' : 'info'">
                <span class="status-dot" :class="row.status === 'admitted' ? 'dot-success' : 'dot-gray'"></span>
                {{ row.status === 'admitted' ? '在院' : '已出院' }}
              </HopeBadge>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <HopeBtn variant="text" size="sm" @click="editPatient(row)">编辑</HopeBtn>
              <HopeBtn variant="warning" size="sm" @click="bindDialogVisible = true; bindTarget = row">绑定腕带</HopeBtn>
            </template>
          </el-table-column>
        </el-table>
      </HopeCard>

      <!-- Edit/Patient Dialog -->
      <el-dialog v-model="showPatientForm" title="编辑/新增患者" width="600px">
        <el-form :model="patientForm" label-width="80px">
          <el-form-item label="住院号"><el-input v-model="patientForm.admission_no" /></el-form-item>
          <el-form-item label="姓名"><el-input v-model="patientForm.name" /></el-form-item>
          <el-form-item label="性别">
            <el-radio-group v-model="patientForm.gender">
              <el-radio value="男">男</el-radio>
              <el-radio value="女">女</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="年龄"><el-input-number v-model="patientForm.age" :min="0" :max="150" /></el-form-item>
          <el-form-item label="科室"><el-input v-model="patientForm.department" /></el-form-item>
          <el-form-item label="床号"><el-input v-model="patientForm.bed_number" /></el-form-item>
          <el-form-item label="血型"><el-input v-model="patientForm.blood_type" /></el-form-item>
          <el-form-item label="过敏史"><el-input v-model="patientForm.allergies" type="textarea" /></el-form-item>
          <el-form-item label="特殊状况"><el-input v-model="patientForm.special_conditions" type="textarea" /></el-form-item>
        </el-form>
        <template #footer>
          <HopeBtn variant="plain" @click="showPatientForm = false">取消</HopeBtn>
          <HopeBtn variant="filled" @click="savePatient">保存</HopeBtn>
        </template>
      </el-dialog>
    </div>

    <!-- TAB: Wristbands -->
    <div v-show="activeTab === 'wristbands'">
      <!-- Filter Bar — HopeCard -->
      <HopeCard style="margin-bottom: 16px;">
        <template #header>
          <span class="filter-title">腕带筛选</span>
        </template>
        <div class="filter-row">
          <div class="filter-item">
            <label class="filter-label">状态</label>
            <el-select v-model="wristbandFilter.status" placeholder="全部状态" clearable @change="loadWristbands" style="width: 100%;">
              <el-option label="空闲" value="idle" />
              <el-option label="已绑定" value="bound" />
              <el-option label="已清空" value="cleared" />
            </el-select>
          </div>
          <div class="filter-item--actions">
            <HopeBtn variant="filled" size="sm" @click="loadWristbands">
              <template #icon>
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg>
              </template>
              刷新
            </HopeBtn>
          </div>
        </div>
      </HopeCard>

      <!-- Wristbands Table -->
      <HopeCard subtitle="腕带设备列表">
        <el-table :data="wristbands" v-loading="loading.wristbands" stripe style="width: 100%;">
          <el-table-column prop="device_id" label="设备ID" width="160">
            <template #default="{ row }"><span class="mono">{{ row.device_id }}</span></template>
          </el-table-column>
          <el-table-column prop="firmware_version" label="固件版本" width="120" />
          <el-table-column prop="status" label="状态" width="90">
            <template #default="{ row }">
              <HopeBadge :color="wristbandHopeColor(row.status)">
                <span class="status-dot" :class="wristbandDotClass(row.status)"></span>
                {{ row.status }}
              </HopeBadge>
            </template>
          </el-table-column>
          <el-table-column prop="bound_patient_id" label="绑定患者" show-overflow-tooltip />
          <el-table-column label="操作" width="240">
            <template #default="{ row }">
              <HopeBtn variant="text" size="sm" @click="clearWristband(row.device_id)">清空数据</HopeBtn>
              <HopeBtn variant="info" size="sm" @click="writeToFirmware(row.device_id)">写入配置</HopeBtn>
            </template>
          </el-table-column>
        </el-table>
      </HopeCard>
    </div>

    <!-- TAB: Verifications -->
    <div v-show="activeTab === 'verifications'">
      <HopeCard subtitle="NFC扫描核验历史记录">
        <el-table :data="verifications" v-loading="loading.verifications" stripe style="width: 100%;">
          <el-table-column prop="timestamp" label="时间" width="180" />
          <el-table-column prop="patient_id" label="患者ID" width="140">
            <template #default="{ row }"><span class="mono">{{ row.patient_id }}</span></template>
          </el-table-column>
          <el-table-column prop="device_id" label="腕带设备" width="140">
            <template #default="{ row }"><span class="mono">{{ row.device_id }}</span></template>
          </el-table-column>
          <el-table-column prop="scan_type" label="类型" width="100">
            <template #default="{ row }">
              <HopeBadge color="primary">
                <span class="status-dot dot-primary"></span>
                {{ scanTypeLabel(row.scan_type) }}
              </HopeBadge>
            </template>
          </el-table-column>
          <el-table-column prop="result" label="结果" width="100">
            <template #default="{ row }">
              <HopeBadge :color="resultHopeColor(row.result)">
                <span class="status-dot" :class="resultDotClass(row.result)"></span>
                {{ resultLabel(row.result) }}
              </HopeBadge>
            </template>
          </el-table-column>
          <el-table-column prop="verified_by" label="操作人" width="100" />
          <el-table-column prop="notes" label="备注" show-overflow-tooltip />
        </el-table>
      </HopeCard>
    </div>

    <!-- TAB: Daily -->
    <div v-show="activeTab === 'daily'">
      <HopeCard subtitle="每日录入记录">
        <template #header>
          <div class="filter-title-row">
            <span class="filter-title">每日录入</span>
            <el-date-picker v-model="dailyDate" type="date" placeholder="选择日期" style="width: 180px;" />
          </div>
        </template>
        <el-table :data="dailyEntries" v-loading="loading.daily" stripe style="width: 100%;">
          <el-table-column prop="timestamp" label="时间" width="180" />
          <el-table-column prop="patient_id" label="患者ID" width="140">
            <template #default="{ row }"><span class="mono">{{ row.patient_id }}</span></template>
          </el-table-column>
          <el-table-column prop="entry_type" label="类型" width="100" />
          <el-table-column prop="content" label="内容" show-overflow-tooltip />
          <el-table-column prop="created_by" label="录入人" width="100" />
        </el-table>
      </HopeCard>
    </div>

    <!-- TAB: Admissions -->
    <div v-show="activeTab === 'admissions'">
      <HopeCard subtitle="入院与出院管理">
        <template #header>
          <div class="filter-title-row">
          <span class="filter-title">出入院管理</span>
            <div class="filter-title-actions">
              <HopeBtn variant="filled" size="sm" @click="showAdmitDialog = true; admitForm = { bed_no: '', department: '', patient_id: '', expected_stay_days: 7 }">
                <template #icon>
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                </template>
                办理入院
              </HopeBtn>
              <HopeBtn variant="plain" size="sm" @click="loadAdmissions">刷新</HopeBtn>
            </div>
          </div>
        </template>
        <el-table :data="admissions" v-loading="loading.admissions" stripe style="width: 100%;">
          <el-table-column prop="admission_no" label="住院号" width="140">
            <template #default="{ row }"><span class="mono">{{ row.admission_no }}</span></template>
          </el-table-column>
          <el-table-column prop="patient_id" label="患者ID" width="140">
            <template #default="{ row }"><span class="mono">{{ row.patient_id }}</span></template>
          </el-table-column>
          <el-table-column prop="bed_no" label="床号" width="80" />
          <el-table-column prop="department" label="科室" width="120" />
          <el-table-column prop="admitted_at" label="入院时间" width="180" />
          <el-table-column prop="expected_discharge_at" label="预计出院" width="180" />
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <HopeBadge :color="row.discharged_at ? 'info' : 'success'">
                <span class="status-dot" :class="row.discharged_at ? 'dot-gray' : 'dot-success'"></span>
                {{ row.discharged_at ? '已出院' : '在院' }}
              </HopeBadge>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <HopeBtn variant="warning" size="sm" @click="showDischargeDialog = true; dischargeTarget = row">出院结算</HopeBtn>
            </template>
          </el-table-column>
        </el-table>
      </HopeCard>
    </div>

    <!-- TAB: Ward Rounds -->
    <div v-show="activeTab === 'ward-rounds'">
      <!-- Filter Bar — HopeCard -->
      <HopeCard style="margin-bottom: 16px;">
        <template #header>
          <span class="filter-title">巡房查询</span>
        </template>
        <div class="filter-row">
          <div class="filter-item">
            <label class="filter-label">选择患者</label>
            <el-select v-model="wardRoundPatientId" placeholder="选择患者" clearable filterable style="width: 100%;">
              <el-option v-for="p in patients" :key="p.id" :label="`${p.name} (${p.admission_no})`" :value="p.id" />
            </el-select>
          </div>
          <div class="filter-item--actions">
            <HopeBtn variant="filled" size="sm" @click="loadWardRounds" :disabled="!wardRoundPatientId">查询</HopeBtn>
            <HopeBtn variant="success" size="sm" @click="showWardRoundForm = true" :disabled="!wardRoundPatientId">开始巡房</HopeBtn>
          </div>
        </div>
      </HopeCard>

      <!-- Ward Rounds Table -->
      <HopeCard subtitle="巡房记录">
        <el-table :data="wardRounds" v-loading="loading.wardRounds" stripe style="width: 100%;">
          <el-table-column prop="nurse_id" label="护士ID" width="140">
            <template #default="{ row }"><span class="mono">{{ row.nurse_id }}</span></template>
          </el-table-column>
          <el-table-column label="血压" width="100">
            <template #default="{ row }">{{ row.blood_pressure || '—' }}</template>
          </el-table-column>
          <el-table-column label="心率" width="80">
            <template #default="{ row }">{{ row.heart_rate ? row.heart_rate + ' bpm' : '—' }}</template>
          </el-table-column>
          <el-table-column label="SpO2" width="80">
            <template #default="{ row }">{{ row.spo2 ? row.spo2 + '%' : '—' }}</template>
          </el-table-column>
          <el-table-column label="体温" width="80">
            <template #default="{ row }">{{ row.temperature ? row.temperature + '°C' : '—' }}</template>
          </el-table-column>
          <el-table-column prop="notes" label="备注" show-overflow-tooltip />
          <el-table-column prop="completed_at" label="完成时间" width="180" />
        </el-table>
      </HopeCard>
    </div>

    <!-- TAB: Regulatory Alerts -->
    <div v-show="activeTab === 'regulatory-alerts'">
      <HopeCard subtitle="监管规则引擎告警">
        <template #header>
          <div class="filter-title-row">
            <span class="filter-title">规则告警</span>
            <HopeBtn variant="plain" size="sm" @click="loadRegulatoryAlerts">刷新</HopeBtn>
          </div>
        </template>
        <el-table :data="regulatoryAlerts" v-loading="loading.regulatoryAlerts" stripe style="width: 100%;">
          <el-table-column prop="rule_code" label="规则" width="80">
            <template #default="{ row }">
              <span class="mono el-tag el-tag--small el-tag--info">{{ row.rule_code }}</span>
            </template>
          </el-table-column>
          <el-table-column label="严重程度" width="100">
            <template #default="{ row }">
              <HopeBadge :color="alertSeverityColor(row.severity)">
                <span class="status-dot" :class="alertSeverityDotClass(row.severity)"></span>
                {{ row.severity }}
              </HopeBadge>
            </template>
          </el-table-column>
          <el-table-column prop="message" label="告警信息" show-overflow-tooltip />
          <el-table-column prop="triggered_at" label="触发时间" width="180" />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <HopeBadge :color="row.resolved ? 'success' : 'warning'">
                <span class="status-dot" :class="row.resolved ? 'dot-success' : 'dot-warning'"></span>
                {{ row.resolved ? '已解决' : '未解决' }}
              </HopeBadge>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <HopeBtn v-if="!row.resolved" variant="filled" size="sm" @click="handleResolveAlert(row)">处理</HopeBtn>
              <span v-else class="mono" style="font-size: 16px; color: var(--hope-success);">✓</span>
            </template>
          </el-table-column>
        </el-table>
      </HopeCard>
    </div>
  </div>

  <!-- Admission Dialog -->
  <el-dialog v-model="showAdmitDialog" title="办理入院" width="500px">
    <el-form :model="admitForm" label-width="100px">
      <el-form-item label="患者ID"><el-input v-model="admitForm.patient_id" /></el-form-item>
      <el-form-item label="床号"><el-input v-model="admitForm.bed_no" /></el-form-item>
      <el-form-item label="科室"><el-input v-model="admitForm.department" /></el-form-item>
      <el-form-item label="诊断"><el-input v-model="admitForm.diagnosis" type="textarea" /></el-form-item>
      <el-form-item label="紧急联系人"><el-input v-model="admitForm.emergency_contact" /></el-form-item>
      <el-form-item label="过敏史"><el-input v-model="admitForm.allergies" type="textarea" /></el-form-item>
      <el-form-item label="预计住院天数"><el-input-number v-model="admitForm.expected_stay_days" :min="1" :max="365" /></el-form-item>
    </el-form>
    <template #footer>
      <HopeBtn variant="plain" @click="showAdmitDialog = false">取消</HopeBtn>
      <HopeBtn variant="filled" @click="handleAdmit">确认入院</HopeBtn>
    </template>
  </el-dialog>

  <!-- Discharge Dialog -->
  <el-dialog v-model="showDischargeDialog" title="出院结算" width="500px">
    <el-form :model="dischargeForm" label-width="100px">
      <el-form-item label="出院类型">
        <el-select v-model="dischargeForm.discharge_type" style="width: 100%;">
          <el-option label="正常出院" value="discharged" />
          <el-option label="转院" value="transferred" />
          <el-option label="死亡" value="deceased" />
        </el-select>
      </el-form-item>
      <el-form-item label="备注"><el-input v-model="dischargeForm.notes" type="textarea" /></el-form-item>
      <el-form-item v-if="dischargeForm.discharge_type === 'transferred'" label="转入科室">
        <el-input v-model="dischargeForm.transferred_to" />
      </el-form-item>
    </el-form>
    <template #footer>
      <HopeBtn variant="plain" @click="showDischargeDialog = false">取消</HopeBtn>
      <HopeBtn variant="filled" @click="handleDischarge">确认出院</HopeBtn>
    </template>
  </el-dialog>

  <!-- Ward Round Form Dialog -->
  <el-dialog v-model="showWardRoundForm" title="填写巡房记录" width="600px">
    <el-form :model="wardRoundForm" label-width="100px">
      <el-form-item label="血压"><el-input v-model="wardRoundForm.blood_pressure" placeholder="如 120/80" /></el-form-item>
      <el-form-item label="心率"><el-input-number v-model="wardRoundForm.heart_rate" :min="0" :max="300" /></el-form-item>
      <el-form-item label="SpO2"><el-input-number v-model="wardRoundForm.spo2" :min="0" :max="100" suffix-icon="%" /></el-form-item>
      <el-form-item label="体温(℃)"><el-input-number v-model="wardRoundForm.temperature" :min="20" :max="45" :precision="1" /></el-form-item>
      <el-form-item label="体重(kg)"><el-input-number v-model="wardRoundForm.weight" :min="0" :max="300" :precision="1" /></el-form-item>
      <el-form-item label="备注"><el-input v-model="wardRoundForm.notes" type="textarea" /></el-form-item>
      <el-form-item label="观察项">
        <el-checkbox-group v-model="wardRoundForm.observationList">
          <el-checkbox label="falls" name="obs">跌倒风险</el-checkbox>
          <el-checkbox label="confusion" name="obs">意识混乱</el-checkbox>
          <el-checkbox label="pain" name="obs">疼痛</el-checkbox>
          <el-checkbox label="appetite" name="obs">食欲不佳</el-checkbox>
        </el-checkbox-group>
      </el-form-item>
    </el-form>
    <template #footer>
      <HopeBtn variant="plain" @click="showWardRoundForm = false">取消</HopeBtn>
      <HopeBtn variant="filled" @click="handleWardRound">提交巡房</HopeBtn>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { medicalApi, type Patient, type WristbandDevice, type VerificationRecord, type HospitalAdmission, type WardRoundEntry, type RegulatoryAlert } from '@/api/medical'
import { HopeStatCard, HopeCard, HopeBadge, HopeBtn, HopeTabs, HopeAvatar } from '@/components/hope'

const activeTab = ref('patients')

const tabItems = [
  { label: '入院登记', value: 'patients' },
  { label: '腕带管理', value: 'wristbands' },
  { label: '核验记录', value: 'verifications' },
  { label: '每日录入', value: 'daily' },
  { label: '出入院管理', value: 'admissions' },
  { label: '巡房记录', value: 'ward-rounds' },
  { label: '规则告警', value: 'regulatory-alerts' },
]

// Stats
const stats = ref({ active_patients: 0, today_admitted: 0, bound_devices: 0, total_devices: 0 })
const todayStats = ref({ matched: 0, total: 0, unmatched: 0 })

// Patients
const patients = ref<Patient[]>([])
const patientForm = ref<Partial<Patient>>({ status: 'admitted' })
const showPatientForm = ref(false)
const bindTarget = ref<Patient | null>(null)
const bindDialogVisible = ref(false)

// Wristbands
const wristbands = ref<WristbandDevice[]>([])
const wristbandFilter = ref({ status: '' })

// Verifications
const verifications = ref<VerificationRecord[]>([])

// Daily
const dailyDate = ref(new Date())
const dailyEntries = ref<any[]>([])

// Admissions
const admissions = ref<HospitalAdmission[]>([])
const showAdmitDialog = ref(false)
const admitForm = ref<Partial<HospitalAdmission>>({ expected_stay_days: 7 })
const showDischargeDialog = ref(false)
const dischargeTarget = ref<HospitalAdmission | null>(null)
const dischargeForm = ref({ discharge_type: 'discharged', notes: '', transferred_to: '' })

// Ward Rounds
const wardRounds = ref<WardRoundEntry[]>([])
const wardRoundPatientId = ref('')
const showWardRoundForm = ref(false)
const wardRoundForm = ref({ nurse_id: 'nurse-1', blood_pressure: '', heart_rate: undefined, spo2: undefined, temperature: undefined, weight: undefined, notes: '', observationList: [] as string[] })

// Regulatory Alerts
const regulatoryAlerts = ref<RegulatoryAlert[]>([])

// Loading states
const loading = ref({
  patients: false,
  wristbands: false,
  verifications: false,
  daily: false,
  admissions: false,
  wardRounds: false,
  regulatoryAlerts: false,
})

onMounted(async () => {
  await Promise.all([loadOverview(), loadPatients(), loadWristbands(), loadVerifications()])
})

async function loadOverview() {
  try {
    const res = await medicalApi.getOverview()
    stats.value = res.data?.data || {}
  } catch { /* ignore */ }
}

async function loadPatients() {
  loading.value.patients = true
  try {
    const res = await medicalApi.listPatients({ page: 1, page_size: 50 })
    patients.value = res.data?.data || []
  } finally {
    loading.value.patients = false
  }
}

async function searchByAdmission() {
  try {
    const res = await medicalApi.getByAdmissionNo(patientForm.value.admission_no!)
    patients.value = [res.data?.data]
  } catch {
    ElMessage.error('未找到该住院号')
  }
}

function editPatient(row: Patient) {
  patientForm.value = { ...row }
  showPatientForm.value = true
}

async function savePatient() {
  try {
    if (patientForm.value.id) {
      await medicalApi.updatePatient(patientForm.value.id!, patientForm.value)
      ElMessage.success('更新成功')
    } else {
      await medicalApi.createPatient(patientForm.value)
      ElMessage.success('创建成功')
    }
    showPatientForm.value = false
    await loadPatients()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

async function loadWristbands() {
  loading.value.wristbands = true
  try {
    const res = await medicalApi.listWristbands({
      page: 1,
      page_size: 50,
      status: wristbandFilter.value.status || undefined,
    })
    wristbands.value = res.data?.data || []
  } finally {
    loading.value.wristbands = false
  }
}

async function clearWristband(deviceId: string) {
  try {
    await medicalApi.clearWristband(deviceId)
    ElMessage.success('腕带已清空')
    await loadWristbands()
  } catch {
    ElMessage.error('清空失败')
  }
}

async function writeToFirmware(deviceId: string) {
  try {
    await medicalApi.writeToFirmware(deviceId, JSON.stringify({ config: 'default' }))
    ElMessage.success('写入成功')
  } catch {
    ElMessage.error('写入失败')
  }
}

async function loadVerifications() {
  loading.value.verifications = true
  try {
    const res = await medicalApi.listVerifications({ page: 1, page_size: 50 })
    verifications.value = res.data?.data || []
    const statsRes = await medicalApi.getTodayStats()
    todayStats.value = statsRes.data?.data || { matched: 0, total: 0, unmatched: 0 }
  } finally {
    loading.value.verifications = false
  }
}

function scanTypeLabel(type: string) {
  const map: Record<string, string> = {
    round: '巡房', medication: '用药', treatment: '治疗', discharge: '出院'
  }
  return map[type] || type
}

function resultLabel(result: string) {
  const map: Record<string, string> = {
    matched: '匹配', unmatched: '不匹配', not_found: '未找到'
  }
  return map[result] || result
}

function wristbandStatusClass(status: string): string {
  if (status === 'bound') return 'badge-success'
  if (status === 'idle') return 'badge-info'
  return 'badge-gray'
}
function wristbandDotClass(status: string): string {
  if (status === 'bound') return 'dot-success'
  if (status === 'idle') return 'dot-info'
  return 'dot-gray'
}

function wristbandHopeColor(status: string): 'success' | 'info' | 'info' {
  if (status === 'bound') return 'success'
  if (status === 'idle') return 'info'
  return 'info'
}

function resultBadgeClass(result: string): string {
  return result === 'matched' ? 'badge-success' : 'badge-danger'
}
function resultDotClass(result: string): string {
  return result === 'matched' ? 'dot-success' : 'dot-danger'
}

function resultHopeColor(result: string): 'success' | 'error' {
  return result === 'matched' ? 'success' : 'error'
}

function alertSeverityColor(severity: string): 'error' | 'warning' | 'primary' {
  if (severity === 'p0') return 'error'
  if (severity === 'p1') return 'warning'
  return 'primary'
}

function alertSeverityDotClass(severity: string): string {
  if (severity === 'p0') return 'dot-danger'
  if (severity === 'p1') return 'dot-warning'
  return 'dot-primary'
}

// --- Admissions ---

async function loadAdmissions() {
  loading.value.admissions = true
  try {
    const res = await medicalApi.listAdmissions({ page: 1, page_size: 50 })
    admissions.value = res.data?.data || []
  } finally {
    loading.value.admissions = false
  }
}

async function handleAdmit() {
  try {
    const days = admitForm.value.expected_stay_days || 7
    await medicalApi.admitPatient({
      patient_id: admitForm.value.patient_id!,
      bed_no: admitForm.value.bed_no!,
      department: admitForm.value.department!,
      diagnosis: admitForm.value.diagnosis,
      emergency_contact: admitForm.value.emergency_contact,
      allergies: admitForm.value.allergies,
      expected_stay_days: days,
    })
    ElMessage.success('入院办理成功')
    showAdmitDialog.value = false
    await loadAdmissions()
  } catch (e: any) {
    ElMessage.error(e.message || '入院办理失败')
  }
}

async function handleDischarge() {
  if (!dischargeTarget.value) return
  try {
    await medicalApi.dischargePatient(dischargeTarget.value.id!, {
      discharge_type: dischargeForm.value.discharge_type,
      notes: dischargeForm.value.notes,
      transferred_to: dischargeForm.value.transferred_to,
    })
    ElMessage.success('出院结算完成')
    showDischargeDialog.value = false
    await loadAdmissions()
  } catch (e: any) {
    ElMessage.error(e.message || '出院结算失败')
  }
}

// --- Ward Rounds ---

async function loadWardRounds() {
  if (!wardRoundPatientId.value) return
  loading.value.wardRounds = true
  try {
    const res = await medicalApi.getWardRounds(wardRoundPatientId.value)
    wardRounds.value = res.data?.data || []
  } finally {
    loading.value.wardRounds = false
  }
}

async function handleWardRound() {
  if (!wardRoundPatientId.value) return
  try {
    const observations = wardRoundForm.value.observationList.join(',') || undefined
    await medicalApi.completeWardRound(wardRoundPatientId.value, {
      nurse_id: wardRoundForm.value.nurse_id || 'nurse-1',
      blood_pressure: wardRoundForm.value.blood_pressure,
      heart_rate: wardRoundForm.value.heart_rate,
      spo2: wardRoundForm.value.spo2,
      temperature: wardRoundForm.value.temperature,
      weight: wardRoundForm.value.weight,
      notes: wardRoundForm.value.notes,
      observations,
    })
    ElMessage.success('巡房记录已提交')
    showWardRoundForm.value = false
    await loadWardRounds()
  } catch (e: any) {
    ElMessage.error(e.message || '巡房提交失败')
  }
}

// --- Regulatory Alerts ---

async function loadRegulatoryAlerts() {
  loading.value.regulatoryAlerts = true
  try {
    const res = await medicalApi.getRegulatoryAlerts({ page: 1, page_size: 50 })
    regulatoryAlerts.value = res.data?.data || []
  } finally {
    loading.value.regulatoryAlerts = false
  }
}

async function handleResolveAlert(row: RegulatoryAlert) {
  try {
    await medicalApi.resolveRegulatoryAlert(row.id, { user_id: 'admin', notes: '已处理' })
    ElMessage.success('告警已解决')
    await loadRegulatoryAlerts()
  } catch (e: any) {
    ElMessage.error(e.message || '处理失败')
  }
}
</script>

<style scoped>
/* ── Page Header ── */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 24px;
}
.page-header__left {}
.page-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--hope-text);
  letter-spacing: -0.02em;
  margin: 0;
  line-height: 1.2;
}
.page-subtitle {
  font-size: 14px;
  color: var(--hope-text-muted);
  margin-top: 4px;
}
.page-header__actions {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

/* ── KPI Grid ── */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}
@media (max-width: 1200px) { .kpi-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 640px)  { .kpi-grid { grid-template-columns: 1fr; } }

/* ── Filter Row (inside HopeCard) ── */
.filter-row {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  flex-wrap: wrap;
}
.filter-item {
  flex: 1;
  min-width: 160px;
}
.filter-item--actions {
  display: flex;
  gap: 8px;
  align-items: flex-end;
  flex-shrink: 0;
}
.filter-label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: var(--hope-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 6px;
}
.filter-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--hope-text);
}
.filter-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}
.filter-title-actions {
  display: flex;
  gap: 8px;
}

/* ── Status dots ── */
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
  flex-shrink: 0;
}
.dot-success { background: var(--hope-success); }
.dot-danger  { background: var(--hope-error); }
.dot-warning { background: var(--hope-warning); }
.dot-primary { background: var(--hope-primary); }
.dot-gray    { background: var(--hope-text-muted); }
.dot-info    { background: #079aa2; }

/* ── Patient cell ── */
.patient-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* ── Mono IDs ── */
.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  color: var(--hope-text-secondary);
}

/* ── Responsive tables ── */
@media (max-width: 768px) {
  .medical-page :deep(.el-table) { font-size: 12px; }
  .medical-page :deep(.el-table th),
  .medical-page :deep(.el-table td) { padding: 8px 6px; }
}
</style>
