# ⑤ 管理后台 — 详细设计文档

> 生成日期：2026-07-17  
> 对应子系统：⑤ 管理后台 (Vue 3 + TypeScript + Element Plus)  
> 语言：TypeScript | 框架：Vue 3 Composition API | UI：Element Plus

---

## 1. 概述

### 1.1 职责

管理后台是运营人员管理整个 Eregen 平台的 Web 应用，提供设备管理、用户管理、订阅管理、仪表盘总览、系统配置等核心功能。采用 SPA 单页应用架构，通过 REST API 与云平台后端通信。

**医护工作站（管理后台内嵌）：**
- 入院登记：护士在管理后台录入患者信息、分配腕带
- 信息录入：每日诊疗记录、用药医嘱、检验检查申请
- 腕带绑定/解绑：管理腕带设备与患者的关联关系
- 核验记录查询：查看近场核验历史日志

**注意：** 护士近场核验功能由独立硬件终端（Android手持PDA + 手机APP）实现，不走浏览器Web NFC API。

### 1.2 输入输出

| 类型 | 来源/目标 | 说明 |
|------|-----------|------|
| **输入** | 运营人员操作 | Vue Router 路由切换 + 组件事件 |
| **输入** | WebSocket 实时数据 | 告警推送、设备状态变更 |
| **输出** | REST API 调用 | 数据查询/管理操作 → api-server/admin-api |
| **输出** | 数据导出 | CSV/PDF 格式运营报表 |

---

## 2. 功能模块

### 2.1 核心页面

| 页面 | 路由 | 核心元素 | 原型文件 |
|------|------|---------|---------|
| 仪表盘总览 | `/dashboard` | 设备在线率、告警统计、用户活跃度、订阅转化率 | `admin-web-dashboard.html` |
| 设备管理 | `/devices` | 设备列表、状态筛选、固件版本分布、OTA 批量升级 | `admin-web-device.html` |
| 用户管理 | `/users` | 老人/家属/机构用户列表、角色标签页、权限管理 | `admin-web-user.html` |
| 订阅管理 | `/subscriptions` | 订阅状态分布、续费记录、降级原因分析 | `admin-web-subscription.html` |
| **监管专区** | **`/regulatory`** | **在院总览看板、异常告警列表、穿透审计详情页、规则配置、合规审查报表** | **`admin-web-regulatory.html`（新增）** |
| **社区老人专区** | **`/community-wb`** | **老人档案管理、福利标签管理、签到总览、药房发药记录、民政数据导入、统计看板** | **`admin-web-community-wb.html`（新增）** |
| **医护工作站** | **`/medical/workstation`** | **入院登记表单、腕带绑定管理、每日诊疗录入、核验记录查询** | **`admin-web-medical-workstation.html`** |
| **医疗腕带管理** | **`/medical-wristband`** | **患者列表、腕带绑定/解绑、核验记录、今日统计** | — |
| **审计详情** | **`/audit-detail/:id`** | **患者全链路时间线（入院→核验→用药→围栏→出院）** | — |
| **系统设置** | **`/settings`** | **通知设置(SOS推送/跌倒告警/用药提醒)、API Key管理、安全设置(密码修改)** | **新建** |
| **OTA 升级** | **`/ota`** | **固件版本列表、适用设备筛选、批量升级推送** | — |
| **老人管理** | **`/elderly`** | **老人档案列表、搜索、详情查看** | **新建** |

### 2.1.1 Dashboard API 对接

```typescript
// views/Dashboard.vue 改动
import { useDashboardStore } from '@/stores/dashboard'
import { useDeviceStore } from '@/stores/device'

const dashboard = useDashboardStore()
const devices = useDeviceStore()

onMounted(async () => {
  await dashboard.fetchOverview()
  await devices.fetchList({ page: 1, page_size: 50 })
})
// ECharts 图表使用 dashboard.stats.alert_trend 等真实数据
// WebSocket 实时告警连接: ws://host/api/v1/admin/stream/alerts
```

### 2.1.2 OTA 升级页面

```vue
<!-- views/OTA.vue -->
<template>
  <el-card>
    <h3>固件版本管理</h3>
    <el-table :data="firmwares">
      <el-table-column prop="version" label="版本号" />
      <el-table-column prop="device_type" label="适用设备" />
      <el-table-column prop="size" label="大小" />
      <el-table-column prop="released_at" label="发布时间" />
      <el-table-column label="操作">
        <el-button @click="triggerOTA($row)">推送升级</el-button>
      </el-table-column>
    </el-table>
  </el-card>
</template>
```

### 2.1.3 系统设置页面

```vue
<!-- views/Settings.vue -->
<template>
  <el-card>
    <el-tabs>
      <!-- 通知设置 -->
      <el-tab-pane label="通知设置">
        <el-form :model="settings">
          <el-switch v-model="settings.sosPush" label="SOS 推送通知" />
          <el-switch v-model="settings.fallAlert" label="跌倒检测告警" />
          <el-switch v-model="settings.medReminder" label="用药提醒通知" />
          <el-switch v-model="settings.emailReport" label="周报邮件" />
        </el-form>
      </el-tab-pane>
      <!-- API Key 管理 -->
      <el-tab-pane label="API Key">
        <el-button @click="showCreateKeyDialog">创建新密钥</el-button>
        <el-table :data="keys">
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="created_at" label="创建时间" />
          <el-table-column label="操作">
            <el-button type="danger" @click="revokeKey">撤销</el-button>
          </el-table-column>
        </el-table>
      </el-tab-pane>
      <!-- 安全设置 -->
      <el-tab-pane label="安全设置">
        <el-form :model="security">
          <el-input v-model="security.oldPassword" type="password" placeholder="当前密码" />
          <el-input v-model="security.newPassword" type="password" placeholder="新密码" />
          <el-button @click="changePassword">修改密码</el-button>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>
```

### 2.2 共享组件

| 组件 | 说明 |
|------|------|
| 侧边栏导航 | 可折叠菜单，图标+文字 |
| 顶部导航栏 | 搜索框、通知铃铛、管理员头像 |
| 数据表格 | 基于 Element Plus Table，支持分页/排序/筛选 |
| 统计卡片 | 数字+趋势箭头+时间范围选择 |
| 图表组件 | ECharts 折线图/饼图/柱状图 |
| 模态对话框 | 表单编辑/详情查看/确认操作 |

---

## 3. 技术架构

### 3.1 项目结构

```
apps/admin-web/
├── src/
│   ├── main.ts                      # 入口，注册 Element Plus
│   ├── App.vue                      # 根组件 (侧边栏 + 内容区布局)
│   ├── router/index.ts              # Vue Router 路由定义
│   ├── views/
│   │   ├── Dashboard.vue            # 仪表盘
│   │   ├── Devices.vue              # 设备管理
│   │   ├── Users.vue                # 用户管理
│   │   ├── Subscriptions.vue        # 订阅管理
│   │   ├── RegulatoryDashboard.vue  # 监管专区（在院总览+告警+审计+规则+合规）
│   │   └── CommunityWristband.vue   # 社区老人专区（老人档案+福利标签+签到+药房+民政）
│   ├── components/                  # 可复用组件
│   │   ├── Sidebar.vue              # 侧边栏导航
│   │   ├── TopNav.vue               # 顶部导航栏
│   │   ├── StatCard.vue             # 统计卡片
│   │   ├── DataTable.vue            # 通用数据表格
│   │   └── ChartWidget.vue          # ECharts 图表封装
│   ├── api/                         # API 请求层
│   │   ├── client.ts                # Axios 实例 (拦截器+Token)
│   │   ├── devices.ts               # 设备相关接口
│   │   ├── users.ts                 # 用户相关接口
│   │   ├── alerts.ts                # 告警相关接口
│   │   └── subscriptions.ts         # 订阅相关接口
│   │   ├── regulatory.ts            # 监管闭环 API（dashboard/alerts/audit/rules/fence/compliance）
│   │   └── community.ts             # 社区老人腕带 API（elders/devices/welfare/signin/pharmacy/minzheng/batch-pay）
│   ├── stores/                      # Pinia 状态管理
│   │   ├── auth.ts                  # 认证状态
│   │   ├── dashboard.ts             # 仪表盘数据
│   │   └── device.ts                # 设备列表状态
│   ├── types/                       # TypeScript 类型定义
│   │   ├── user.ts
│   │   ├── device.ts
│   │   ├── alert.ts
│   │   └── health.ts
│   └── utils/                       # 工具函数
│       ├── format.ts                # 日期/数字格式化
│       └── constants.ts             # 常量定义
├── public/                          # 静态资源
├── index.html                       # HTML 模板
└── vite.config.ts                   # Vite 构建配置
```

### 3.2 技术栈

| 库 | 版本 | 用途 |
|----|------|------|
| Vue | 3.4+ | 响应式 UI 框架 |
| TypeScript | 5.4+ | 类型安全 |
| Element Plus | 2.7+ | UI 组件库 |
| Vue Router | 4.x | 路由管理 |
| Pinia | 2.x | 状态管理 |
| Axios | 1.x | HTTP 客户端 |
| ECharts | 5.x | 数据可视化图表 |
| Vite | 5.x | 构建工具 |

### 3.3 状态管理 (Pinia)

```typescript
// stores/auth.ts — 认证状态
export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: '',
    user: null as User | null,
    role: '' as Role,  // 'admin' | 'operator'
  }),
  actions: {
    async login(identifier: string, password: string) { ... },
    async logout() { ... },
    hasPermission(resource: string): boolean { ... },
  },
})

// stores/dashboard.ts — 仪表盘数据
export const useDashboardStore = defineStore('dashboard', {
  state: () => ({
    onlineDevices: 0,
    totalDevices: 0,
    activeAlerts: 0,
    totalUsers: 0,
    activeSubscriptions: 0,
    alertTrend: [] as AlertTrendPoint[],
  }),
  actions: {
    async fetchOverview() { ... },
  },
})
```

---

## 4. 各页面详细设计

### 4.1 仪表盘 (`/dashboard`)

```
┌─────────────────────────────────────────────────────┐
│  [Logo] Eregen 管理后台                    🔔 [管理员] │
├──────────┬──────────────────────────────────────────┤
│ 📊 概览   │  仪表盘                                │
│ 📱 设备   │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  │
│ 👥 用户   │  │在线率│ │告警  │ │用户  │ │订阅  │  │
│ 💊 订阅   │  │ 92%↑│ │ 12  │ │ 1,234│ │ 856  │  │
│ 🏥 医护工作站│  └──────┘ └──────┘ └──────┘ └──────┘  │
│ 🔍 监管专区│                                      │
│ 👴 社区老人│  ┌─────────────────┐ ┌─────────────────┐ │
│          │                                          │
│          │  ┌─────────────────┐ ┌─────────────────┐ │
│          │  │ 设备在线趋势     │ │ 告警类型分布     │ │
│          │  │ [折线图 7天]     │ │ [饼图 P0/P1/P2] │ │
│          │  └─────────────────┘ └─────────────────┘ │
│          │                                          │
│          │  ┌─────────────────┐ ┌─────────────────┐ │
│          │  │ 用户增长趋势     │ │ 最近告警列表     │ │
│          │  │ [柱状图 30天]    │ │ [表格 最新10条] │ │
│          │  └─────────────────┘ └─────────────────┘ │
└──────────┴──────────────────────────────────────────┘
```

### 4.2 设备管理 (`/devices`)

| 功能 | 说明 |
|------|------|
| 设备列表 | 表格展示所有设备，显示 ID、类型、档位、状态、最后在线时间 |
| 状态筛选 | 全部/在线/离线 |
| 类型筛选 | 手环/药盒 |
| 档位筛选 | Starter/Plus/Pro |
| 搜索 | 按设备 ID 或所有者名称搜索 |
| OTA 升级 | 选中设备 → 上传固件包 → 批量下发升级指令 |
| 配置更新 | 修改设备参数 (采样间隔、音量、定位频率) |
| 详情弹窗 | 查看设备完整信息、历史告警、健康数据摘要 |

### 4.3 用户管理 (`/users`)

| 功能 | 说明 |
|------|------|
| 角色标签页 | 老人 / 家属 / 机构 |
| 用户列表 | 姓名、联系方式、注册时间、关联设备数 |
| 权限管理 | 分配/撤销角色权限 |
| 操作日志 | 查看用户关键操作记录 |
| 搜索/筛选 | 按名称、注册时间范围筛选 |

### 4.4 订阅管理 (`/subscriptions`)

| 功能 | 说明 |
|------|------|
| 订阅分布 | Free / Premium / Enterprise 数量占比饼图 |
| 续费记录 | 历史续费列表，支持导出 CSV |
| 降级分析 | 按降级原因分类统计 (价格/需求/体验/流失) |
| 到期提醒 | 即将到期的订阅列表 |

### 4.5 医护工作站 (`/medical/workstation`)

**职责：** 护士在管理后台进行入院登记、腕带绑定、每日诊疗录入、核验记录查询。

#### 4.5.1 患者入院登记

| 功能 | 说明 |
|------|------|
| 手动录入 | 填写患者姓名、性别、年龄、住院号、科室、床号、血型、过敏史、特殊疾病、警示标签 |
| 扫码导入 | 扫描住院单据二维码，自动解析并填充患者信息 |
| 批量导入 | Excel/CSV 模板批量导入，支持错误行提示和跳过 |
| 住院号校验 | 全院唯一性校验，防止重复入院登记 |

#### 4.5.2 患者信息管理

| 功能 | 说明 |
|------|------|
| 信息编辑 | 修改患者基础信息、医疗信息、警示标签 |
| 信息更新推送 | 修改后一键推送到已绑定腕带，实时更新本地存储 |
| 医疗清单管理 | 录入/编辑/删除住院费用、用药清单、检测报告 |
| 每日诊疗录入 | 护士每日录入查房记录、护理记录、医嘱执行 |
| 患者状态管理 | 入院/转科/出院/死亡状态流转 |

#### 4.5.3 腕带绑定管理

| 功能 | 说明 |
|------|------|
| 单独绑定 | 选择患者 → 选择腕带 → 绑定，腕带写入患者信息 |
| 批量绑定 | 批量选择患者和腕带，按顺序绑定 |
| 批量写入 | 选中多个患者 → 批量下发到对应腕带 |
| 信息更新 | 修改患者信息后重新写入腕带 |
| 出院解绑 | 患者出院 → 解除绑定 → 清空腕带数据（远程清除 + 长按按键 10 秒本地清空） |
| 腕带库存管理 | 查看腕带设备列表、固件版本、在线状态、绑定状态 |

#### 4.5.4 数据统计看板

| 统计项 | 说明 |
|--------|------|
| 在院患者总数 | 当前绑定腕带的在院患者数量 |
| 今日入院/出院 | 当日入院和出院患者数量 |
| 今日核验次数 | 护士端读取腕带的总次数 |
| 腕带使用率 | 已绑定腕带数 / 腕带总数 |
| 警示标签分布 | 过敏/跌倒高危/隔离等标签的患者数量分布 |
| 腕带固件版本分布 | 各版本固件的设备数量统计 |

**与护士核验终端的关系：**
- 管理后台医护工作站负责**信息录入和管理**（入院登记、腕带绑定、诊疗记录）
- 护士核验终端（Android PDA/手机APP）负责**现场近场核验**（NFC读取腕带数据、比对医嘱一致性）
- 两者通过云端 API 同步数据：`/api/v1/medical/`

---

## 5. 接口定义

```
GET    /api/v1/devices                    # 设备列表
GET    /api/v1/devices/:id                # 设备详情
PUT    /api/v1/devices/:id/settings       # 更新配置
POST   /api/v1/devices/:id/ota            # 触发 OTA 升级

GET    /api/v1/users?role=&page=&page_size=  # 用户列表
GET    /api/v1/users/:id                  # 用户详情
PUT    /api/v1/users/:id                  # 更新用户
PUT    /api/v1/users/:id/role             # 修改角色

GET    /api/v1/alerts?severity=&status=   # 告警列表
PUT    /api/v1/alerts/:id/status          # 标记处理

GET    /api/v1/subscriptions              # 订阅列表
GET    /api/v1/subscriptions/stats        # 统计数据
GET    /api/v1/admin/stats/overview       # 仪表盘总览数据

# 医护工作站 API
POST   /api/v1/medical/patients           # 入院登记
GET    /api/v1/medical/patients           # 患者列表
GET    /api/v1/medical/patients/:id       # 患者详情
PUT    /api/v1/medical/patients/:id       # 更新患者信息
DELETE /api/v1/medical/patients/:id       # 出院注销
POST   /api/v1/medical/patients/batch-import  # 批量导入（Excel/CSV）
GET    /api/v1/medical/patients/by-admission-no  # 按住院号查询
POST   /api/v1/medical/patients/:id/bind      # 腕带绑定
POST   /api/v1/medical/patients/:id/unbind    # 腕带解绑
POST   /api/v1/medical/patients/batch-bind    # 批量绑定
POST   /api/v1/medical/wristbands/:device_id/write  # 写入腕带固件
POST   /api/v1/medical/wristbands/:device_id/clear  # 出院清空腕带
GET    /api/v1/medical/wristbands             # 腕带设备列表
GET    /api/v1/medical/wristbands/:device_id/firmware  # 腕带固件版本
POST   /api/v1/medical/lists/expenses         # 录入费用
GET    /api/v1/medical/lists/expenses         # 查询费用清单
POST   /api/v1/medical/lists/medications      # 录入用药
GET    /api/v1/medical/lists/medications      # 查询用药清单
POST   /api/v1/medical/lists/tests            # 录入检测报告
GET    /api/v1/medical/lists/tests            # 查询检测报告
POST   /api/v1/medical/daily/entries          # 每日诊疗录入
GET    /api/v1/medical/daily/entries          # 诊疗记录列表
GET    /api/v1/medical/history?elderly_id=    # 治疗经过（家属端）
GET    /api/v1/medical/verifications          # 核验记录列表
PUT    /api/v1/medical/verifications/:id/status  # 标记核验完成
GET    /api/v1/medical/verifications/stats/today  # 今日核验统计
GET    /api/v1/medical/stats/overview         # 数据统计看板

GET    /api/v1/medical/stats/overview         # 数据统计看板

# 监管专区 API
GET    /api/v1/admin/regulatory/dashboard/patient-overview  # 在院总览摘要
GET    /api/v1/admin/regulatory/dashboard/patient-list      # 在院患者列表
GET    /api/v1/admin/regulatory/alerts                      # 告警列表
POST   /api/v1/admin/regulatory/alerts                      # 创建告警
POST   /api/v1/admin/regulatory/alerts/:id/acknowledge      # 确认告警
POST   /api/v1/admin/regulatory/alerts/:id/resolve          # 标记解决
GET    /api/v1/admin/regulatory/audit/patient/:id           # 穿透审计全链路
GET    /api/v1/admin/regulatory/rules                       # 规则配置
PUT    /api/v1/admin/regulatory/rules                       # 更新规则配置
GET    /api/v1/admin/regulatory/fence/config                # 围栏配置
POST   /api/v1/admin/regulatory/fence/config                # 创建围栏
GET    /api/v1/admin/regulatory/compliance/report           # 合规报表

# 社区老人专区 API
GET    /api/v1/admin/community-wb/elders                    # 老人档案列表
POST   /api/v1/admin/community-wb/elders                    # 创建老人档案
PUT    /api/v1/admin/community-wb/elders/:id                # 更新老人档案
DELETE /api/v1/admin/community-wb/elders/:id                # 删除老人档案
GET    /api/v1/admin/community-wb/devices                   # 腕带设备列表
POST   /api/v1/admin/community-wb/devices                   # 注册腕带设备
GET    /api/v1/admin/community-wb/welfare-tags              # 福利标签配置
POST   /api/v1/admin/community-wb/welfare-tags              # 新增福利标签
PUT    /api/v1/admin/community-wb/welfare-tags/:id          # 更新福利标签
DELETE /api/v1/admin/community-wb/welfare-tags/:id          # 删除福利标签
POST   /api/v1/admin/community-wb/signin/trigger            # 签到激活
GET    /api/v1/admin/community-wb/signin/records            # 签到记录
POST   /api/v1/admin/community-wb/pharmacy/dispense         # 药房发药
POST   /api/v1/admin/community-wb/minzheng/import           # 民政数据导入
POST   /api/v1/admin/community-wb/batch-pay/execute         # 批量发放执行
GET    /api/v1/admin/community-wb/batch-payments            # 发放记录

WS     /api/v1/admin/stream/alerts        # WebSocket 实时告警
```

---

## 6. 编译与运行

```bash
cd apps/admin-web

# 安装依赖
npm install

# 开发模式 (热更新)
npm run dev
# 访问 http://localhost:5173

# 类型检查
npm run type-check

# 构建生产包
npm run build
# 输出: dist/

# 预览生产构建
npm run preview
```

---

## 7. 测试策略

| 测试类型 | 工具 | 覆盖范围 |
|---------|------|---------|
| 单元测试 | Vitest | API 客户端、工具函数、Store |
| 组件测试 | Vue Test Utils | 各页面组件渲染和交互 |
| E2E 测试 | Playwright | 完整业务流程：登录→设备管理→告警处理 |

---

## 8. 业务链权限与页面隔离

### 8.1 角色-页面映射矩阵

| 页面路由 | super_admin | operator | hospital_doc | nurse | community_staff | regulator |
|---------|-------------|----------|--------------|-------|-----------------|-----------|
| `/dashboard` | ✅ | ✅ 自营数据 | ❌ | ❌ | ❌ | ❌ |
| `/elderly` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| `/devices` | ✅ | ✅ 自营设备 | ❌ | ❌ | ❌ | ❌ |
| `/medical/workstation` | ✅ | ❌ | ✅ | ✅ 执行 | ❌ | ❌ |
| `/medical-wristband` | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| `/community-wb` | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ |
| `/regulatory` | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| `/subscriptions` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| `/settings` | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

### 8.2 关键变更说明

1. **operator（运营平台）**：仅能看到自营链数据（老人档案、自营设备、订阅），看不到住院链和社区链
2. **hospital_doc（医院医生）**：仅能看到住院链数据，看到社区链只读
3. **community_staff（社区工作人员）**：仅能看到社区链数据
4. **regulator（监管角色）**：能看到住院链和社区链（只读），不能看自营链
5. **super_admin（超级用户）**：能看到所有链的全部数据

### 8.3 路由守卫实现

```typescript
// router/guards.ts
const roleChainMap: Record<Role, BusinessChain[]> = {
  super_admin: ['self', 'hospital', 'community', 'regulatory'],
  operator: ['self'],
  hospital_doc: ['hospital'],
  nurse: ['hospital'],
  community_staff: ['community'],
  regulator: ['hospital', 'community', 'regulatory'],
}

function checkChainAccess(role: Role, requiredChain: BusinessChain): boolean {
  return roleChainMap[role]?.includes(requiredChain) ?? false
}
```

---

© 2026 Eregen (颐贞). All rights reserved.
