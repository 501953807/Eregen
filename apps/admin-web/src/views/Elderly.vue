<template>
  <div class="community-wb-page">
    <!-- Page Header -->
    <div class="page-header">
      <el-breadcrumb separator="/">
        <el-breadcrumb-item>首页</el-breadcrumb-item>
        <el-breadcrumb-item>社区老人专区</el-breadcrumb-item>
        <el-breadcrumb-item>{{ pageTitles[activePage] }}</el-breadcrumb-item>
      </el-breadcrumb>
      <h2 class="page-title">{{ pageTitles[activePage] }}</h2>
    </div>

    <!-- ==================== 1. 老人档案管理 ==================== -->
    <template v-if="activePage === 'elderly'">
      <!-- KPI Row (6 columns) -->
      <el-row :gutter="12" style="margin-bottom: 20px;">
        <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-blue"><div class="kpi-value">{{ kpis.total }}</div><div class="kpi-label">登记老人</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-green"><div class="kpi-value">{{ kpis.wearable }}</div><div class="kpi-label">在线腕带</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="hover" class="kpi-card"><div class="kpi-value">{{ kpis.welfareTags }}</div><div class="kpi-label">福利标签</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-warning"><div class="kpi-value">{{ kpis.todaySignin }}</div><div class="kpi-label">今日签到</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-danger"><div class="kpi-value">{{ kpis.pendingReview }}</div><div class="kpi-label">待审核民政</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-warning"><div class="kpi-value">{{ kpis.alerts }}</div><div class="kpi-label">异常告警</div></el-card></el-col>
      </el-row>

      <!-- Filter Bar -->
      <div class="filter-bar">
        <el-button type="primary" @click="showAddElder = true">＋ 新增老人</el-button>
        <el-input v-model="elderlySearch" placeholder="搜索姓名 / 身份证号 / 手机号" clearable style="width: 300px;" :prefix-icon="Search" />
      </div>

      <!-- Elderly Table -->
      <el-card shadow="never" class="table-card">
        <el-table :data="filteredElderly" stripe v-loading="loading.elderly">
          <el-table-column prop="name" label="姓名" width="90">
            <template #default="{ row }"><strong>{{ row.name }}</strong></template>
          </el-table-column>
          <el-table-column label="身份证号" width="180">
            <template #default="{ row }"><span class="mono">{{ row.id_card || '—' }}</span></template>
          </el-table-column>
          <el-table-column label="年龄" width="60">
            <template #default="{ row }">{{ calculateAge(row.birth_date) }}</template>
          </el-table-column>
          <el-table-column label="性别" width="60">
            <template #default="{ row }">{{ row.gender || '—' }}</template>
          </el-table-column>
          <el-table-column label="紧急联系人" width="110">
            <template #default="{ row }">{{ row.emergency_contact || '—' }}</template>
          </el-table-column>
          <el-table-column label="福利标签" min-width="200">
            <template #default="{ row }">
              <el-tag v-for="tag in (row.welfare_tags || [])" :key="tag.code" :type="welfareTagType(tag.code)" size="small" style="margin-right: 4px;">{{ tag.name }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <span class="status-badge" :class="row.status === '正常' ? 'badge-success' : 'badge-gray'">
                <span class="status-dot" :class="row.status === '正常' ? 'dot-success' : 'dot-gray'"></span>{{ row.status }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="140" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="openDetail(row)">详情</el-button>
              <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-wrapper">
          <el-pagination background layout="total, prev, pager, next" :total="elderlyStore.total" :current-page="page" :page-size="pageSize" @current-change="handlePageChange" />
        </div>
      </el-card>
    </template>

    <!-- ==================== 2. 福利标签管理 ==================== -->
    <template v-if="activePage === 'welfare'">
      <el-row :gutter="12" style="margin-bottom: 20px;">
        <el-col :span="6"><el-card shadow="hover" class="kpi-card"><div class="kpi-value">{{ welfareKpis.valid }}</div><div class="kpi-label">有效标签</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="hover" class="kpi-card kpi-warning"><div class="kpi-value">{{ welfareKpis.expiring }}</div><div class="kpi-label">本月到期</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="hover" class="kpi-card kpi-green"><div class="kpi-value">{{ welfareKpis.newIssued }}</div><div class="kpi-label">本月新发</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="hover" class="kpi-card"><div class="kpi-value">{{ welfareKpis.revoked }}</div><div class="kpi-label">本月撤销</div></el-card></el-col>
      </el-row>
      <div class="filter-bar">
        <el-button type="primary">＋ 新增标签</el-button>
        <el-button>批量分配</el-button>
      </div>
      <el-card shadow="never" class="table-card">
        <el-table :data="welfareList" stripe>
          <el-table-column prop="code" label="标签代码" width="150">
            <template #default="{ row }"><span class="mono">{{ row.code }}</span></template>
          </el-table-column>
          <el-table-column prop="name" label="标签名称" width="120" />
          <el-table-column prop="issuer" label="发放机构" width="100" />
          <el-table-column label="Renewal 周期" width="100">
            <template #default="{ row }">{{ row.renewal_days }} 天</template>
          </el-table-column>
          <el-table-column label="补助金额" width="100">
            <template #default="{ row }">¥{{ row.subsidy_amount }}</template>
          </el-table-column>
          <el-table-column label="绑定老人" width="90" align="center">
            <template #default="{ row }"><strong>{{ row.bound_count }}</strong></template>
          </el-table-column>
          <el-table-column label="启用" width="80" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.enabled" @change="toggleWelfare(row)" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="viewBoundElders(row)">查看绑定</el-button>
              <el-button link type="primary" size="small">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>

    <!-- ==================== 3. 签到总览 ==================== -->
    <template v-if="activePage === 'signin'">
      <el-row :gutter="12" style="margin-bottom: 20px;">
        <el-col :span="4"><el-card shadow="hover" class="kpi-card"><div class="kpi-value">856</div><div class="kpi-label">本月签到</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-green"><div class="kpi-value">234</div><div class="kpi-label">本月首次</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-danger"><div class="kpi-value">3</div><div class="kpi-label">跨院重复</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="hover" class="kpi-card"><div class="kpi-value">189</div><div class="kpi-label">医保签到</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-green"><div class="kpi-value">667</div><div class="kpi-label">福利签到</div></el-card></el-col>
        <el-col :span="4"><el-card shadow="hover" class="kpi-card kpi-warning"><div class="kpi-value">2</div><div class="kpi-label">异常</div></el-card></el-col>
      </el-row>
      <div class="filter-bar">
        <el-date-picker v-model="signinMonth" type="month" placeholder="月份选择" value-format="YYYY-MM" />
        <el-select v-model="signinHospital" placeholder="全部医院" style="width: 140px;">
          <el-option label="社区医院 A" value="A" />
          <el-option label="社区医院 B" value="B" />
          <el-option label="社区医院 C" value="C" />
        </el-select>
        <el-button type="primary" :icon="Search">查询</el-button>
      </div>
      <!-- CSS Bar Chart -->
      <el-card shadow="never" class="chart-card">
        <template #header><span class="panel-title">近 7 天签到趋势</span></template>
        <div class="bar-chart">
          <div v-for="(d, i) in signinTrend" :key="i" class="bar-col">
            <div class="bar-value">{{ d.count }}</div>
            <div class="bar" :style="{ height: (d.count / 62 * 100) + 'px' }"></div>
            <div class="bar-label">{{ d.day }}</div>
          </div>
        </div>
      </el-card>
      <el-card shadow="never" class="table-card" style="margin-top: 20px;">
        <el-table :data="signinRecords" stripe>
          <el-table-column prop="name" label="老人姓名" width="100" />
          <el-table-column label="身份证号" width="140">
            <template #default="{ row }"><span class="mono">{{ row.id_card }}</span></template>
          </el-table-column>
          <el-table-column prop="hospital" label="医院" width="120" />
          <el-table-column prop="signin_time" label="签到时间" width="170" />
          <el-table-column label="激活标签" min-width="180">
            <template #default="{ row }">
              <el-tag v-for="t in (row.tags || [])" :key="t" size="small" style="margin-right: 4px;">{{ t }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="类型" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="row.type === '福利签到' ? 'badge-success' : 'badge-primary'">
                <span class="status-dot" :class="row.type === '福利签到' ? 'dot-success' : 'dot-primary'"></span>{{ row.type }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>

    <!-- ==================== 4. 药房发药记录 ==================== -->
    <template v-if="activePage === 'pharmacy'">
      <el-row :gutter="12" style="margin-bottom: 20px;">
        <el-col :span="6"><el-card shadow="hover" class="kpi-card"><div class="kpi-value">34</div><div class="kpi-label">今日发药</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="hover" class="kpi-card kpi-green"><div class="kpi-value">512</div><div class="kpi-label">本月发药</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="hover" class="kpi-card"><div class="kpi-value">28</div><div class="kpi-label">药品种类</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="hover" class="kpi-card kpi-warning"><div class="kpi-value">¥12,450</div><div class="kpi-label">总金额</div></el-card></el-col>
      </el-row>
      <div class="filter-bar">
        <el-button type="primary">＋ 手动发药</el-button>
        <el-date-picker v-model="pharmacyMonth" type="month" placeholder="月份" value-format="YYYY-MM" />
        <el-input v-model="pharmacySearch" placeholder="搜索老人姓名 / 药品名" clearable style="width: 220px;" :prefix-icon="Search" />
      </div>
      <el-card shadow="never" class="table-card">
        <el-table :data="pharmacyRecords" stripe>
          <el-table-column label="日期" width="80">
            <template #default="{ row }">{{ row.date }}</template>
          </el-table-column>
          <el-table-column prop="name" label="老人姓名" width="100" />
          <el-table-column prop="hospital" label="医院" width="110" />
          <el-table-column prop="medications" label="药品清单" min-width="180" show-overflow-tooltip />
          <el-table-column label="金额" width="80">
            <template #default="{ row }">¥{{ row.amount }}</template>
          </el-table-column>
          <el-table-column prop="staff" label="药师/护士" width="90" />
          <el-table-column label="签到状态" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="row.signed_in ? 'badge-success' : 'badge-warning'">
                <span class="status-dot" :class="row.signed_in ? 'dot-success' : 'dot-warning'"></span>{{ row.signed_in ? '已签到' : '未签到' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination-wrapper">
          <el-pagination background layout="total, prev, pager, next" :total="512" :current-page="1" :page-size="20" />
        </div>
      </el-card>
    </template>

    <!-- ==================== 5. 民政数据导入 ==================== -->
    <template v-if="activePage === 'minzheng'">
      <el-row :gutter="12" style="margin-bottom: 20px;">
        <el-col :span="6"><el-card shadow="hover" class="kpi-card"><div class="kpi-value">12</div><div class="kpi-label">导入批次</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="hover" class="kpi-card kpi-green"><div class="kpi-value">1,234</div><div class="kpi-label">总导入</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="hover" class="kpi-card"><div class="kpi-value">1,198</div><div class="kpi-label">已匹配</div></el-card></el-col>
        <el-col :span="6"><el-card shadow="hover" class="kpi-card kpi-warning"><div class="kpi-value">36</div><div class="kpi-label">待审核</div></el-card></el-col>
      </el-row>
      <div class="filter-bar">
        <el-upload action="#" :auto-upload="false" :show-file-list="false" class="upload-zone">
          <div class="upload-inner">
            <div class="upload-icon">📁</div>
            <p>点击或拖拽 CSV/XLSX 文件到此处上传</p>
            <p style="font-size:11px;margin-top:4px;color:#c0c4cc;">支持民政局标准模板或自定义模板</p>
          </div>
        </el-upload>
        <el-button>📥 下载 CSV 模板</el-button>
      </div>
      <!-- Template Info -->
      <el-card shadow="never" class="table-card" style="margin-bottom: 20px;">
        <template #header><span class="panel-title">CSV 字段说明</span></template>
        <el-table :data="csvTemplateFields" stripe size="small">
          <el-table-column prop="field" label="列名" width="120" />
          <el-table-column label="必填" width="80">
            <template #default="{ row }">
              <span class="status-badge" :class="row.required ? 'badge-danger' : 'badge-gray'">
                <span class="status-dot" :class="row.required ? 'dot-danger' : 'dot-gray'"></span>{{ row.required ? '是' : '否' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="desc" label="说明" />
          <el-table-column prop="example" label="示例" width="180">
            <template #default="{ row }"><span class="mono">{{ row.example }}</span></template>
          </el-table-column>
        </el-table>
      </el-card>
      <!-- Import Records -->
      <el-card shadow="never" class="table-card">
        <el-table :data="importRecords" stripe>
          <el-table-column prop="source" label="数据来源" width="120" />
          <el-table-column prop="filename" label="文件名" width="140" />
          <el-table-column label="导入数" width="80">
            <template #default="{ row }"><strong>{{ row.imported }}</strong></template>
          </el-table-column>
          <el-table-column label="匹配数" width="80">
            <template #default="{ row }">{{ row.matched }}</template>
          </el-table-column>
          <el-table-column label="待审核" width="80">
            <template #default="{ row }"><strong :style="{ color: row.pending > 0 ? '#EF4444' : '' }">{{ row.pending }}</strong></template>
          </el-table-column>
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <span class="status-badge" :class="row.status === '完成' ? 'badge-success' : 'badge-warning'">
                <span class="status-dot" :class="row.status === '完成' ? 'dot-success' : 'dot-warning'"></span>{{ row.status }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="170" />
        </el-table>
      </el-card>
    </template>

    <!-- ==================== 6. 统计看板 ==================== -->
    <template v-if="activePage === 'stats'">
      <div class="filter-bar" style="margin-bottom: 20px;">
        <el-date-picker v-model="statsMonth" type="month" placeholder="月份" value-format="YYYY-MM" />
        <el-select v-model="statsHospital" placeholder="全部医院" style="width: 140px;">
          <el-option label="社区医院 A" value="A" />
          <el-option label="社区医院 B" value="B" />
          <el-option label="社区医院 C" value="C" />
        </el-select>
        <el-button type="primary" :icon="Refresh">刷新</el-button>
      </div>

      <!-- Row 1: 3 stat boxes -->
      <el-row :gutter="16" style="margin-bottom: 16px;">
        <el-col :span="8">
          <el-card shadow="hover" class="stat-box">
            <template #header><span class="panel-title">登记老人总数</span></template>
            <div class="stat-center">
              <div class="stat-big-num">482</div>
              <div style="font-size:12px;color:#909399;">在线腕带 312 · 离线 170</div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="hover" class="stat-box">
            <template #header><span class="panel-title">福利标签分布</span></template>
            <div class="h-bars">
              <div v-for="w in welfareDist" :key="w.code" class="h-bar-row">
                <span class="h-bar-label">{{ w.label }}</span>
                <div class="h-bar-track"><div class="h-bar-fill" :style="{ width: Math.min(w.pct, 100) + '%', background: w.color }"></div></div>
                <span class="h-bar-val">{{ w.count }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="hover" class="stat-box">
            <template #header><span class="panel-title">签到活跃度</span></template>
            <div class="activity-stats">
              <div class="act-row">
                <span>本月签到率</span>
                <el-progress :percentage="87" :color="'#2563EB'" :stroke-width="8" />
                <strong>87%</strong>
              </div>
              <div class="act-row">
                <span>连续签到≥3月</span>
                <el-progress :percentage="68" :color="'#16A34A'" :stroke-width="8" />
                <strong>68%</strong>
              </div>
              <div class="act-row">
                <span>本月首次签到</span>
                <span style="font-weight:600;color:#F59E0B;">234 人</span>
              </div>
              <div class="act-row">
                <span>跨院重复</span>
                <span style="font-weight:600;color:#EF4444;">3 人次</span>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- Row 2: Hospital, Payment, Alerts -->
      <el-row :gutter="16">
        <el-col :span="8">
          <el-card shadow="hover" class="stat-box">
            <template #header><span class="panel-title">医院分布</span></template>
            <div class="h-bars">
              <div v-for="h in hospitalDist" :key="h.name" class="h-bar-row">
                <span class="h-bar-label">{{ h.name }}</span>
                <div class="h-bar-track"><div class="h-bar-fill" :style="{ width: h.pct + '%', background: h.color }"></div></div>
                <span class="h-bar-val">{{ h.count }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="hover" class="stat-box">
            <template #header><span class="panel-title">补助发放统计</span></template>
            <div class="payment-stats">
              <div class="pay-total">¥45,200</div>
              <div style="font-size:12px;color:#909399;margin-bottom:12px;">本月发放总额</div>
              <div class="pay-metrics">
                <div class="pay-metric"><div style="font-size:18px;font-weight:600;color:#16A34A;">92%</div><div style="font-size:11px;color:#909399;">成功率</div></div>
                <div class="pay-metric"><div style="font-size:18px;font-weight:600;color:#EF4444;">8</div><div style="font-size:11px;color:#909399;">失败笔数</div></div>
                <div class="pay-metric"><div style="font-size:18px;font-weight:600;color:#F59E0B;">12</div><div style="font-size:11px;color:#909399;">待发笔数</div></div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card shadow="hover" class="stat-box">
            <template #header><span class="panel-title">规则引擎告警</span></template>
            <div class="alert-list">
              <div v-for="a in ruleAlerts" :key="a.code" class="alert-item">
                <span class="status-badge" :class="a.tagType === 'danger' ? 'badge-danger' : a.tagType === 'warning' ? 'badge-warning' : 'badge-gray'">
                <span class="status-dot" :class="a.tagType === 'danger' ? 'dot-danger' : a.tagType === 'warning' ? 'dot-warning' : 'dot-gray'"></span>{{ a.code }}</span>
                <span>{{ a.desc }}</span>
                <strong>{{ a.count }}</strong>
              </div>
            </div>
            <div style="text-align:center;margin-top:12px;">
              <el-button size="small">查看全部告警 →</el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </template>

    <!-- ==================== Detail Dialog ==================== -->
    <el-dialog v-model="showDetailDialog" :title="'老人档案详情 — ' + (detailElder?.name || '')" width="640px" destroy-on-close>
      <div v-if="detailElder">
        <h4 style="margin-bottom:12px;color:#303133;">基本信息</h4>
        <div class="detail-grid">
          <div class="detail-item"><span class="label">姓名：</span><span class="value">{{ detailElder.name }}</span></div>
          <div class="detail-item"><span class="label">性别：</span><span class="value">{{ detailElder.gender || '—' }}</span></div>
          <div class="detail-item"><span class="label">年龄：</span><span class="value">{{ detailElder.birth_date ? calculateAge(detailElder.birth_date) + ' 岁' : '—' }}</span></div>
          <div class="detail-item"><span class="label">身份证号：</span><span class="value mono">{{ detailElder.id_card || '—' }}</span></div>
          <div class="detail-item"><span class="label">地址：</span><span class="value">{{ detailElder.address || '—' }}</span></div>
          <div class="detail-item"><span class="label">紧急联系人：</span><span class="value">{{ detailElder.emergency_contact || '—' }}</span></div>
          <div class="detail-item"><span class="label">腕带设备：</span><span class="value">{{ detailElder.wearable_id || '—' }} <span v-if="detailElder.wearable_online" class="status-badge badge-success"><span class="status-dot dot-success"></span>在线</span></span></div>
          <div class="detail-item"><span class="label">状态：</span><span class="value"><span class="status-badge" :class="detailElder.status === '正常' ? 'badge-success' : 'badge-gray'"><span class="status-dot" :class="detailElder.status === '正常' ? 'dot-success' : 'dot-gray'"></span>{{ detailElder.status }}</span></span></div>
        </div>
        <h4 style="margin:20px 0 12px;color:#303133;">福利标签</h4>
        <el-table :data="(detailElder.welfare_tags || []).map(t => ({ ...t, status: '有效' }))" stripe size="small">
          <el-table-column label="标签" width="100">
            <template #default="{ row }"><el-tag :type="welfareTagType(row.code)" size="small">{{ row.name }}</el-tag></template>
          </el-table-column>
          <el-table-column prop="issuer" label="发放机构" width="100" />
          <el-table-column label="生效日期" width="110">
            <template #default="{ row }">{{ row.start_date }}</template>
          </el-table-column>
          <el-table-column label="到期日期" width="110">
            <template #default="{ row }">{{ row.end_date }}</template>
          </el-table-column>
          <el-table-column label="状态" width="80">
            <template #default="{ row }"><span class="status-badge badge-success"><span class="status-dot dot-success"></span>{{ row.status }}</span></template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="showDetailDialog = false">关闭</el-button>
        <el-button type="primary">编辑档案</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Search, Refresh } from '@element-plus/icons-vue'

const activePage = ref('elderly')
const pageTitles: Record<string, string> = {
  elderly: '老人档案管理',
  welfare: '福利标签管理',
  signin: '签到总览',
  pharmacy: '药房发药记录',
  minzheng: '民政数据导入',
  stats: '统计看板',
}

// Switch page
function switchPage(page: string) {
  activePage.value = page
}

// ==================== Elderly ====================
const loading = ref({ elderly: false })
const elderlySearch = ref('')
const page = ref(1)
const pageSize = ref(20)

const kpis = ref({ total: 482, wearable: 312, welfareTags: 9, todaySignin: 28, pendingReview: 5, alerts: 3 })

interface WelfareTag { code: string; name: string }
interface ElderlyRow {
  id: string; name: string; id_card?: string; birth_date?: string; gender?: string
  emergency_contact?: string; welfare_tags: WelfareTag[]; status: string
  address?: string; wearable_id?: string; wearable_online?: boolean
}

const elderlyList = ref<ElderlyRow[]>([
  { id: '1', name: '张秀兰', id_card: '510101195001011234', gender: '女', birth_date: '1950-01-01', emergency_contact: '张明（子）', welfare_tags: [{ code: 'orphan', name: '孤寡' }, { code: 'poverty_1', name: '特困一级' }, { code: 'disability_2', name: '残疾二级' }], status: '正常' },
  { id: '2', name: '李建国', id_card: '510101194805055678', gender: '男', birth_date: '1948-05-05', emergency_contact: '李华（女）', welfare_tags: [{ code: 'special_disease', name: '特病门诊' }, { code: 'bus_discount', name: '公交优惠' }], status: '正常' },
  { id: '3', name: '王秀英', id_card: '510101195503127890', gender: '女', birth_date: '1955-03-12', emergency_contact: '王芳（女）', welfare_tags: [{ code: 'medical_assist', name: '医疗救助' }], status: '停用' },
  { id: '4', name: '赵德柱', id_card: '510101194208153456', gender: '男', birth_date: '1942-08-15', emergency_contact: '赵强（子）', welfare_tags: [{ code: 'orphan', name: '孤寡' }, { code: 'poverty_1', name: '特困一级' }, { code: 'special_disease', name: '特病门诊' }], status: '正常' },
  { id: '5', name: '刘美华', id_card: '510101195812256789', gender: '女', birth_date: '1958-12-25', emergency_contact: '刘晓（女）', welfare_tags: [{ code: 'disability_3', name: '残疾三级' }, { code: 'bus_discount', name: '公交优惠' }], status: '正常' },
])

const filteredElderly = computed(() => {
  if (!elderlySearch.value) return elderlyList.value
  const q = elderlySearch.value.toLowerCase()
  return elderlyList.value.filter(e =>
    e.name.toLowerCase().includes(q) ||
    (e.id_card && e.id_card.includes(q))
  )
})

function calculateAge(birthDate?: string): number {
  if (!birthDate) return 0
  const today = new Date()
  const birth = new Date(birthDate)
  let age = today.getFullYear() - birth.getFullYear()
  if (today.getMonth() < birth.getMonth() || (today.getMonth() === birth.getMonth() && today.getDate() < birth.getDate())) age--
  return age
}

function welfareTagType(code: string): 'danger' | 'warning' | 'primary' | 'success' | 'info' | 'danger' {
  const map: Record<string, 'danger' | 'warning' | 'primary' | 'success' | 'info'> = {
    orphan: 'danger', poverty_1: 'warning', poverty_2: 'warning',
    disability_1: 'primary', disability_2: 'primary', disability_3: 'primary',
    special_disease: 'info', bus_discount: 'success', medical_assist: 'primary',
  }
  return map[code] || 'info'
}

function handlePageChange(p: number) { page.value = p }

// Detail dialog
const showDetailDialog = ref(false)
const detailElder = ref<ElderlyRow | null>(null)
function openDetail(row: ElderlyRow) { detailElder.value = row; showDetailDialog.value = true }
function openEdit(row: ElderlyRow) { /* TODO */ }

// ==================== Welfare ====================
const welfareKpis = ref({ valid: 9, expiring: 3, newIssued: 12, revoked: 2 })
const welfareList = ref([
  { code: 'orphan', name: '孤寡老人', issuer: '民政局', renewal_days: 365, subsidy_amount: 0, bound_count: 12, enabled: true },
  { code: 'poverty_level_1', name: '特困一级', issuer: '民政局', renewal_days: 365, subsidy_amount: 800, bound_count: 28, enabled: true },
  { code: 'poverty_level_2', name: '特困二级', issuer: '民政局', renewal_days: 365, subsidy_amount: 500, bound_count: 15, enabled: true },
  { code: 'disability_1', name: '残疾一级', issuer: '残联', renewal_days: 365, subsidy_amount: 600, bound_count: 22, enabled: true },
  { code: 'disability_2', name: '残疾二级', issuer: '残联', renewal_days: 365, subsidy_amount: 400, bound_count: 35, enabled: true },
  { code: 'disability_3', name: '残疾三级', issuer: '残联', renewal_days: 365, subsidy_amount: 200, bound_count: 42, enabled: true },
  { code: 'special_disease', name: '特病门诊', issuer: '医保局', renewal_days: 180, subsidy_amount: 0, bound_count: 89, enabled: true },
  { code: 'bus_discount', name: '公交优惠', issuer: '交通局', renewal_days: 30, subsidy_amount: 0, bound_count: 156, enabled: true },
  { code: 'medical_assist', name: '医疗救助', issuer: '民政局', renewal_days: 365, subsidy_amount: 1000, bound_count: 67, enabled: true },
])

function toggleWelfare(row: any) { /* TODO */ }
function viewBoundElders(row: any) { /* TODO */ }

// ==================== Signin ====================
const signinMonth = ref('2026-07')
const signinHospital = ref('')
const signinTrend = [
  { day: '周一', count: 42 }, { day: '周二', count: 56 }, { day: '周三', count: 48 },
  { day: '周四', count: 41 }, { day: '周五', count: 62 }, { day: '周六', count: 28 }, { day: '周日', count: 19 },
]
const signinRecords = ref([
  { name: '张秀兰', id_card: '...1234', hospital: '社区医院 A', signin_time: '2026-07-23 10:30', tags: ['孤寡', '特困一级', '残疾二级'], type: '福利签到' },
  { name: '李建国', id_card: '...5678', hospital: '社区医院 B', signin_time: '2026-07-23 09:15', tags: ['特病门诊', '公交优惠'], type: '医保签到' },
  { name: '王秀英', id_card: '...7890', hospital: '社区医院 A', signin_time: '2026-07-22 14:20', tags: ['医疗救助'], type: '福利签到' },
  { name: '赵德柱', id_card: '...3456', hospital: '社区医院 C', signin_time: '2026-07-22 11:05', tags: ['孤寡', '特困一级', '特病门诊'], type: '福利签到' },
  { name: '刘美华', id_card: '...6789', hospital: '社区医院 A', signin_time: '2026-07-21 16:40', tags: ['残疾三级', '公交优惠'], type: '福利签到' },
])

// ==================== Pharmacy ====================
const pharmacyMonth = ref('2026-07')
const pharmacySearch = ref('')
const pharmacyRecords = ref([
  { date: '07-23', name: '张秀兰', hospital: '社区医院 A', medications: '氨氯地平、二甲双胍', amount: '45.50', staff: '张护士', signed_in: true },
  { date: '07-23', name: '李建国', hospital: '社区医院 B', medications: '阿司匹林肠溶片', amount: '12.00', staff: '李药师', signed_in: true },
  { date: '07-22', name: '王秀英', hospital: '社区医院 A', medications: '硝苯地平缓释片', amount: '28.00', staff: '张护士', signed_in: false },
  { date: '07-22', name: '赵德柱', hospital: '社区医院 C', medications: '格列本脲、二甲双胍', amount: '67.30', staff: '王药师', signed_in: true },
  { date: '07-21', name: '刘美华', hospital: '社区医院 A', medications: '氨氯地平', amount: '18.50', staff: '张护士', signed_in: true },
])

// ==================== Minzheng ====================
const csvTemplateFields = [
  { field: '姓名', required: true, desc: '老人姓名', example: '张秀兰' },
  { field: '身份证号', required: true, desc: '18位身份证号码', example: '510101195001011234' },
  { field: '福利类型', required: true, desc: 'orphan/poverty_level_1/disability_2/...', example: '特困' },
  { field: '认定等级', required: true, desc: '1/2/3', example: '一级' },
  { field: '有效期开始', required: true, desc: 'YYYY-MM-DD', example: '2025-01-01' },
  { field: '有效期结束', required: true, desc: 'YYYY-MM-DD', example: '2028-12-31' },
  { field: '备注', required: false, desc: '额外信息', example: '肢体残疾' },
]

const importRecords = ref([
  { source: 'XX区民政局', filename: '202607.csv', imported: 234, matched: 230, pending: 4, status: '完成', created_at: '2026-07-23 10:00' },
  { source: 'XX街道办', filename: '7月数据.xlsx', imported: 156, matched: 155, pending: 1, status: '完成', created_at: '2026-07-20 14:30' },
  { source: 'XX区残联', filename: '残疾人补贴.csv', imported: 89, matched: 85, pending: 4, status: '完成', created_at: '2026-07-18 09:15' },
  { source: 'XX市民政局', filename: '特困人员汇总.csv', imported: 312, matched: 298, pending: 14, status: '处理中', created_at: '2026-07-24 08:00' },
])

// ==================== Stats ====================
const statsMonth = ref('2026-07')
const statsHospital = ref('')

const welfareDist = [
  { code: 'orphan', label: '孤寡', count: 12, pct: 12, color: '#c62828' },
  { code: 'poverty_1', label: '特困一', count: 28, pct: 28, color: '#e65100' },
  { code: 'poverty_2', label: '特困二', count: 15, pct: 15, color: '#f57c00' },
  { code: 'disability_1', label: '残疾一', count: 22, pct: 22, color: '#1565c0' },
  { code: 'disability_2', label: '残疾二', count: 35, pct: 35, color: '#1976d2' },
  { code: 'disability_3', label: '残疾三', count: 42, pct: 42, color: '#42a5f5' },
  { code: 'special_disease', label: '特病', count: 89, pct: 89, color: '#9c27b0' },
  { code: 'bus_discount', label: '公交', count: 156, pct: 100, color: '#4caf50' },
  { code: 'medical_assist', label: '医疗', count: 67, pct: 67, color: '#00897b' },
]

const hospitalDist = [
  { name: '社区医院A', count: 234, pct: 100, color: '#2563EB' },
  { name: '社区医院B', count: 156, pct: 67, color: '#7C3AED' },
  { name: '社区医院C', count: 92, pct: 39, color: '#EC4899' },
]

const ruleAlerts = [
  { code: 'R_C01', desc: '重复领取', count: 3, tagType: 'danger' },
  { code: 'R_C02', desc: '冒领嫌疑', count: 1, tagType: 'danger' },
  { code: 'R_C03', desc: '异常高频', count: 2, tagType: 'warning' },
  { code: 'R_C04', desc: '僵尸账户', count: 5, tagType: 'info' },
  { code: 'R_C05', desc: '补助未到账', count: 1, tagType: 'warning' },
]

onMounted(() => { /* load data */ })
</script>

<style scoped>
.community-wb-page {
  padding: 0;
}

.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 22px;
  font-weight: 800;
  color: var(--el-text-color-primary);
  margin: 8px 0 0;
}

/* KPI Cards — v2 palette */
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
.kpi-warning .kpi-value { color: #F59E0B; }
.kpi-danger .kpi-value { color: #EF4444; }

/* Filter bar */
.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

/* Upload zone */
.upload-zone {
  border: 2px dashed var(--el-border-color);
  border-radius: 12px;
  padding: 32px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.upload-zone:hover {
  border-color: #2563EB;
  background: #EFF6FF;
}

.upload-inner p {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.upload-icon {
  font-size: 36px;
  margin-bottom: 8px;
}

/* Table card */
.table-card {
  margin-bottom: 20px;
}

.mono {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

/* Panel title */
.panel-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  border-left: 3px solid #2563EB;
  padding-left: 8px;
}

/* Bar chart */
.chart-card :deep(.el-card__header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.bar-chart {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  padding: 16px 0;
  height: 130px;
}

.bar-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.bar {
  width: 32px;
  background: linear-gradient(180deg, #2563EB, #7C3AED);
  border-radius: 4px 4px 0 0;
  min-height: 4px;
  transition: height 0.3s;
}

.bar-label {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  margin-top: 6px;
}

.bar-value {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 4px;
}

/* Horizontal bars */
.h-bars {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.h-bar-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.h-bar-label {
  width: 60px;
  text-align: right;
  color: var(--el-text-color-regular);
  flex-shrink: 0;
  font-size: 12px;
}

.h-bar-track {
  flex: 1;
  height: 14px;
  background: var(--el-fill-color-lighter);
  border-radius: 3px;
  overflow: hidden;
}

.h-bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s;
}

.h-bar-val {
  width: 40px;
  color: var(--el-text-color-placeholder);
  font-size: 12px;
  flex-shrink: 0;
  text-align: right;
}

/* Stat boxes */
.stat-box :deep(.el-card__body) {
  padding: 16px;
}

.stat-center {
  text-align: center;
  padding: 12px 0;
}

.stat-big-num {
  font-size: 36px;
  font-weight: 800;
  background: linear-gradient(135deg, #2563EB, #7C3AED);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* Activity stats */
.activity-stats {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.act-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  gap: 8px;
}

.act-row span:first-child {
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}

/* Payment stats */
.payment-stats {
  text-align: center;
  padding: 8px 0;
}

.pay-total {
  font-size: 24px;
  font-weight: 800;
  color: #2563EB;
}

.pay-metrics {
  display: flex;
  justify-content: space-around;
}

.pay-metric {
  text-align: center;
}

/* Alert list */
.alert-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.alert-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  font-size: 13px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.alert-item:last-child {
  border-bottom: none;
}

/* Detail grid */
.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.detail-item {
  display: flex;
  font-size: 13px;
}

.detail-item .label {
  width: 80px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}

.detail-item .value {
  color: var(--el-text-color-primary);
  font-weight: 500;
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

/* Responsive */
@media (max-width: 1200px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>
