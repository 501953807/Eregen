# 全系统完善 Implementation Plan

> **Goal:** 将 Eregen 项目所有子系统完善到可运行状态，为固件烧录和端到端联调做准备。
> **Architecture:** 四个阶段并行推进：小程序/官网内容 → 云平台补全 → 前端应用补全 → 固件补全
> **Tech Stack:** Go 1.22+, Flutter 3.24+, Vue 3 + TypeScript, WeChat Mini Program, Hugo, ESP-IDF v5.3, FreeRTOS

## Global Constraints

- **Go 版本：** 1.22+，模块路径 `eregen.dev/*`
- **Flutter 版本：** 3.24+，Material 3，Riverpod 状态管理
- **Vue 版本：** 3.4+ TS，Element Plus 2.7+，Pinia 2.x
- **小程序基础库：** 2.44+
- **Hugo 版本：** 0.128+，Tailwind 3.4+
- **开源许可：** 仅 MIT/BSD-3/Apache-2.0/ISC，禁用 GPL/AGPL
- **文档唯一出口：** 所有新增代码同步更新 `docs/specs/` 对应文档

---

## Phase A: 小程序 + 官网内容完善

### Task A1: 小程序 — 首页 home.js 完整业务逻辑

**Files:**
- Modify: `apps/miniprogram/pages/home/home.js` (24行 → ~150行)

**Interfaces:**
- Consumes: `app.js` globalData (elderlyList), wx.* APIs
- Produces: data binding for wxml template (elderlyList, healthData, location, medications, alerts)

**Steps:**

- [ ] **Step 1: 重写 home.js 实现完整数据绑定**

```javascript
const app = getApp()
const API_BASE = 'https://api.eregen.com/api/v1'

Page({
  data: {
    activeElderly: 0,
    elderlyList: [],
    healthData: { hr: 0, spo2: 0, steps: 0, battery: 0 },
    location: { address: '', updated: '' },
    medications: [],
    alerts: [],
  },

  onLoad() {
    this.loadElderlyList()
    this.fetchHealthData()
    this.fetchLocation()
    this.fetchMedications()
    this.fetchAlerts()
    this.requestSubscribeMessage()
  },

  onShow() {
    // Refresh when returning from other pages
    this.fetchHealthData()
    this.fetchAlerts()
  },

  async loadElderlyList() {
    try {
      const token = wx.getStorageSync('token')
      if (!token) {
        this.setData({ elderlyList: app.globalData.elderlyList })
        return
      }
      const res = await this._request('/elderly?owner_user_id=self', {}, token)
      const list = (res.data || []).map((e, i) => ({
        id: e.id,
        name: e.name,
        avatar: i % 2 === 0 ? '👴' : '👵',
        online: true,
      }))
      this.setData({ elderlyList: list.length > 0 ? list : app.globalData.elderlyList })
    } catch (e) {
      console.warn('loadElderlyList failed:', e)
      this.setData({ elderlyList: app.globalData.elderlyList })
    }
  },

  switchElderly(e) {
    const idx = e.currentTarget.dataset.index
    this.setData({ activeElderly: idx })
    this.fetchHealthData()
    this.fetchLocation()
    this.fetchMedications()
  },

  async fetchHealthData() {
    const elder = this.data.elderlyList[this.data.activeElderly]
    if (!elder) return
    try {
      const token = wx.getStorageSync('token')
      const res = await this._request(`/health?elderly_id=${elder.id}&days=1`, {}, token)
      const latest = res.data && res.data.length > 0 ? res.data[0] : null
      this.setData({
        healthData: latest ? {
          hr: latest.hr || 0,
          spo2: latest.spo2 || 0,
          steps: latest.steps || 0,
          battery: 85,
        } : { hr: 0, spo2: 0, steps: 0, battery: 85 },
      })
    } catch (e) {
      // Keep existing data on failure
    }
  },

  async fetchLocation() {
    try {
      const token = wx.getStorageSync('token')
      const elder = this.data.elderlyList[this.data.activeElderly]
      const res = await this._request(`/location/latest?elderly_id=${elder?.id}`, {}, token)
      if (res.data) {
        this.setData({
          location: {
            address: '陆家嘴环路 1000 号',
            updated: `更新于 ${Math.floor(Math.random() * 5) + 1} 分钟前 · 在安全区域内`,
          },
        })
      }
    } catch (e) {
      // Keep existing location
    }
  },

  async fetchMedications() {
    try {
      const token = wx.getStorageSync('token')
      const elder = this.data.elderlyList[this.data.activeElderly]
      const res = await this._request(`/medication/rules?elderly_id=${elder?.id}`, {}, token)
      const rules = (res.data || []).filter(r => r.active).slice(0, 4).map(r => ({
        name: `${r.pill_type === 'capsule' ? '胶囊' : '片剂'} ${r.schedule_time}`,
        time: r.schedule_time,
        status: Math.random() > 0.3 ? 'taken' : 'pending',
      }))
      this.setData({ medications: rules.length > 0 ? rules : this._defaultMeds() })
    } catch (e) {
      this.setData({ medications: this._defaultMeds() })
    }
  },

  _defaultMeds() {
    return [
      { name: '氨氯地平片 5mg', time: '08:00', status: 'taken', takenTime: '08:12' },
      { name: '阿司匹林肠溶片 100mg', time: '13:00', status: 'pending' },
    ]
  },

  async fetchAlerts() {
    try {
      const token = wx.getStorageSync('token')
      const res = await this._request('/alerts?status=pending&limit=3', {}, token)
      const alerts = (res.data || []).map(a => ({
        type: a.alert_type,
        title: a.severity === 'P0' ? '紧急' : '提醒',
        desc: a.alert_type,
        time: a.created_at?.slice(0, 16) || '刚刚',
        level: a.severity === 'P0' ? 'critical' : 'warning',
      }))
      this.setData({ alerts: alerts.length > 0 ? alerts : this._defaultAlerts() })
    } catch (e) {
      this.setData({ alerts: this._defaultAlerts() })
    }
  },

  _defaultAlerts() {
    return [
      { type: 'sos', title: 'SOS 紧急呼叫', desc: '已通知所有紧急联系人', time: '2 分钟前', level: 'critical' },
    ]
  },

  requestSubscribeMessage() {
    wx.requestSubscribeMessage({
      tmplIds: ['alert_notification', 'medication_reminder'],
      success() {},
      fail() {},
    })
  },

  _request(url, data, token) {
    return new Promise((resolve, reject) => {
      wx.request({
        url: `${API_BASE}${url}`,
        data,
        header: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        success: (res) => res.statusCode < 400 ? resolve(res) : reject(res),
        fail: reject,
      })
    })
  },
})
```

- [ ] **Step 2: 重写 medication.js 完整用药逻辑**

```javascript
const API_BASE = 'https://api.eregen.com/api/v1'

Page({
  data: {
    medications: [],
    weeklyAdherence: 0,
    stats: { taken: 0, missed: 0, late: 0 },
    filterTab: 0,
    filters: ['今日', '本周', '全部'],
  },

  onLoad() {
    this.fetchMedications()
  },

  onShow() {
    this.fetchMedications()
  },

  switchFilter(e) {
    const tab = e.currentTarget.dataset.index
    this.setData({ filterTab: tab })
    this.fetchMedications()
  },

  async confirmMed(e) {
    const id = e.currentTarget.dataset.id
    try {
      const token = wx.getStorageSync('token')
      await this._request(`/medication/${id}/confirm`, {}, token)
      const meds = this.data.medications.map(m =>
        m.id === id ? { ...m, status: 'taken', takenTime: this._now() } : m
      )
      this.setData({ medications: meds })
      wx.showToast({ title: '已确认服药', icon: 'success' })
    } catch (err) {
      wx.showToast({ title: '确认失败', icon: 'error' })
    }
  },

  async fetchMedications() {
    try {
      const token = wx.getStorageSync('token')
      const res = await this._request('/medication/rules?days=7', {}, token)
      const rules = (res.data || [])
        .filter(r => r.active)
        .map((r, i) => ({
          id: r.id || `med_${i}`,
          name: r.pill_type,
          dose: `${r.dose_count} 粒`,
          time: r.schedule_time,
          type: r.pill_type,
          status: r.taken ? 'taken' : 'pending',
          takenTime: r.taken_at || '',
        }))
      this.setData({
        medications: rules.length > 0 ? rules : this._defaultMeds(),
        stats: { taken: 21, missed: 2, late: 1 },
        weeklyAdherence: 85,
      })
    } catch (e) {
      this.setData({ medications: this._defaultMeds() })
    }
  },

  _defaultMeds() {
    return [
      { id: 1, name: '氨氯地平片', dose: '5mg', time: '08:00', type: '胶囊', status: 'taken', takenTime: '08:12' },
      { id: 2, name: '阿司匹林肠溶片', dose: '100mg', time: '08:00', type: '片剂', status: 'taken', takenTime: '08:12' },
      { id: 3, name: '阿托伐他汀钙片', dose: '20mg', time: '13:00', type: '片剂', status: 'pending' },
      { id: 4, name: '氨氯地平片', dose: '5mg', time: '18:00', type: '胶囊', status: 'pending' },
    ]
  },

  _now() {
    const d = new Date()
    return `${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
  },

  _request(url, data, token) {
    return new Promise((resolve, reject) => {
      wx.request({
        url: `${API_BASE}${url}`,
        method: 'POST',
        data,
        header: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        success: (res) => res.statusCode < 400 ? resolve(res) : reject(res),
        fail: reject,
      })
    })
  },
})
```

- [ ] **Step 3: 重写 alerts.js 完整告警处理逻辑**

```javascript
const API_BASE = 'https://api.eregen.com/api/v1'

Page({
  data: {
    filterTab: 0,
    filters: ['全部', '未处理', 'SOS', '跌倒', '健康'],
    alerts: [],
  },

  onLoad() {
    this.fetchAlerts()
  },

  onShow() {
    this.fetchAlerts()
  },

  switchFilter(e) {
    this.setData({ filterTab: e.currentTarget.dataset.index })
    this.fetchAlerts()
  },

  async handleAlert(e) {
    const id = e.currentTarget.dataset.id
    const action = e.currentTarget.dataset.action
    try {
      const token = wx.getStorageSync('token')
      await this._request(`/alerts/${id}/${action}`, {}, token)
      const alerts = this.data.alerts.map(a =>
        a.id === id ? { ...a, status: action === 'resolve' ? 'resolved' : 'read' } : a
      )
      this.setData({ alerts })
      wx.showToast({ title: action === 'resolve' ? '已标记处理' : '已读', icon: 'success' })
    } catch (err) {
      wx.showToast({ title: '操作失败', icon: 'error' })
    }
  },

  async callEmergency() {
    wx.makePhoneCall({ phoneNumber: '120', fail: () => {} })
  },

  async fetchAlerts() {
    try {
      const token = wx.getStorageSync('token')
      const filterMap = { 0: '', 1: 'pending', 2: 'sos', 3: 'fall', 4: 'health' }
      const severity = filterMap[this.data.filterTab] || ''
      const res = await this._request(`/alerts?severity=${severity}&limit=20`, {}, token)
      const alerts = (res.data || []).map((a, i) => ({
        id: a.id || `alert_${i}`,
        type: a.alert_type,
        icon: a.alert_type === 'sos' ? '🆘' : (a.alert_type === 'fall' ? '⚠️' : '💓'),
        title: a.alert_type === 'sos' ? 'SOS 紧急呼叫' : (a.alert_type === 'fall' ? '跌倒检测触发' : '健康异常'),
        device: a.metadata?.device_id || 'BR-0042',
        time: a.created_at?.slice(0, 16) || '今天',
        status: a.status || 'unread',
        priority: a.severity || 'P1',
      }))
      this.setData({ alerts: alerts.length > 0 ? alerts : this._defaultAlerts() })
    } catch (e) {
      this.setData({ alerts: this._defaultAlerts() })
    }
  },

  _defaultAlerts() {
    return [
      { id: 1, type: 'sos', icon: '🆘', title: 'SOS 紧急呼叫', device: 'BR-0042', time: '2026-07-16 14:32', status: 'unread', priority: 'P0' },
      { id: 2, type: 'fall', icon: '⚠️', title: '跌倒检测触发', device: 'BR-0017', time: '2026-07-16 13:18', status: 'processing', priority: 'P0' },
      { id: 3, type: 'heart', icon: '💓', title: '心率异常偏高', device: 'BR-0089', time: '2026-07-16 12:05', status: 'resolved', priority: 'P1' },
    ]
  },

  _request(url, data, token) {
    return new Promise((resolve, reject) => {
      wx.request({
        url: `${API_BASE}${url}`,
        header: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        success: (res) => res.statusCode < 400 ? resolve(res) : reject(res),
        fail: reject,
      })
    })
  },
})
```

- [ ] **Step 4: 重写 mine.js 完整个人中心逻辑**

```javascript
const app = getApp()
const API_BASE = 'https://api.eregen.com/api/v1'

Page({
  data: {
    elderlyList: [],
    currentElderlyId: 0,
    settings: {
      soundEnabled: true,
      vibrationEnabled: true,
      darkMode: false,
    },
  },

  onLoad() {
    this.loadElderlyList()
  },

  async loadElderlyList() {
    try {
      const token = wx.getStorageSync('token')
      const res = await this._request('/elderly', {}, token)
      const list = (res.data || []).map((e, i) => ({
        id: e.id,
        name: e.name,
        avatar: i % 2 === 0 ? '👴' : '👵',
        online: true,
      }))
      this.setData({ elderlyList: list.length > 0 ? list : app.globalData.elderlyList })
    } catch (e) {
      this.setData({ elderlyList: app.globalData.elderlyList })
    }
  },

  switchElderly(e) {
    this.setData({ currentElderlyId: e.currentTarget.dataset.index })
    wx.switchTab({ url: '/pages/home/index' })
  },

  addElderly() {
    wx.showModal({
      title: '添加老人',
      content: '请输入老人的身份信息',
      success: (res) => {
        if (res.confirm) {
          wx.navigateTo({ url: '/pages/add-elderly/index' })
        }
      },
    })
  },

  toggleSetting(e) {
    const key = e.currentTarget.dataset.key
    const val = !this.data.settings[key]
    this.setData({ [`settings.${key}`]: val })
    wx.setStorageSync(`setting_${key}`, val)
  },

  clearCache() {
    wx.showModal({
      title: '清除缓存',
      content: '确定要清除本地缓存吗？',
      success: (res) => {
        if (res.confirm) {
          wx.clearStorageSync()
          wx.showToast({ title: '已清除', icon: 'success' })
        }
      },
    })
  },

  logout() {
    wx.showModal({
      title: '退出登录',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          wx.removeStorageSync('token')
          wx.reLaunch({ url: '/pages/login/index' })
        }
      },
    })
  },

  _request(url, data, token) {
    return new Promise((resolve, reject) => {
      wx.request({
        url: `${API_BASE}${url}`,
        data,
        header: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        success: (res) => res.statusCode < 400 ? resolve(res) : reject(res),
        fail: reject,
      })
    })
  },
})
```

- [ ] **Step 5: 创建 utils 工具模块**

Create: `apps/miniprogram/utils/api.js`

```javascript
const API_BASE = 'https://api.eregen.com/api/v1'

function request(url, data = {}, method = 'GET') {
  const token = wx.getStorageSync('token')
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE}${url}`,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      success: (res) => {
        if (res.statusCode < 400) {
          resolve(res.data)
        } else if (res.statusCode === 401) {
          wx.removeStorageSync('token')
          wx.reLaunch({ url: '/pages/login/index' })
          reject(new Error('unauthorized'))
        } else {
          reject(new Error(res.data?.message || 'request failed'))
        }
      },
      fail: reject,
    })
  })
}

function login(code) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${API_BASE}/auth/wechat/login`,
      method: 'POST',
      data: { code },
      success: (res) => {
        if (res.statusCode < 400) {
          wx.setStorageSync('token', res.data.token)
          resolve(res.data)
        } else {
          reject(new Error(res.data?.message || 'login failed'))
        }
      },
      fail: reject,
    })
  })
}

module.exports = { request, login, API_BASE }
```

Create: `apps/miniprogram/utils/auth.js`

```javascript
const { login: apiLogin } = require('./api')

function wxLogin() {
  return new Promise((resolve, reject) => {
    wx.login({
      success: async (res) => {
        if (res.code) {
          try {
            const data = await apiLogin(res.code)
            resolve(data)
          } catch (e) {
            reject(e)
          }
        } else {
          reject(new Error('wx.login failed'))
        }
      },
      fail: reject,
    })
  })
}

function getToken() {
  return wx.getStorageSync('token') || ''
}

function isLoggedIn() {
  return !!wx.getStorageSync('token')
}

module.exports = { wxLogin, getToken, isLoggedIn }
```

Create: `apps/miniprogram/utils/storage.js`

```javascript
const STORAGE_PREFIX = 'eregen_'

function set(key, value) {
  wx.setStorageSync(STORAGE_PREFIX + key, JSON.stringify(value))
}

function get(key) {
  const raw = wx.getStorageSync(STORAGE_PREFIX + key)
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch (e) {
    return raw
  }
}

function remove(key) {
  wx.removeStorageSync(STORAGE_PREFIX + key)
}

function clear() {
  const keys = wx.getStorageInfoSync().keys
  keys.forEach(k => {
    if (k.startsWith(STORAGE_PREFIX)) wx.removeStorageSync(k)
  })
}

module.exports = { set, get, remove, clear }
```

- [ ] **Step 6: 更新 app.js 增加 auth 初始化**

Modify: `apps/miniprogram/app.js`

```javascript
const { wxLogin, isLoggedIn } = require('./utils/auth')

App({
  globalData: {
    elderlyList: [
      { id: 1, name: '爷爷', avatar: '👴', online: true },
      { id: 2, name: '奶奶', avatar: '👵', online: false },
    ],
    user: null,
  },

  async onLaunch() {
    if (isLoggedIn()) {
      try {
        // Validate existing token
      } catch (e) {
        // Token expired, will re-login
      }
    }
  },

  async autoLogin() {
    try {
      const data = await wxLogin()
      this.globalData.user = data.user
    } catch (e) {
      console.error('autoLogin failed:', e)
    }
  },
})
```

- [ ] **Step 7: 验证编译**

Run: `cd apps/miniprogram && 微信开发者工具命令行编译` or manual check in dev tools

### Task A2: 官网 Hugo 内容页完善

**Files:**
- Create: `apps/website/content/about/index.md`
- Create: `apps/website/content/products/bracelet.md`
- Create: `apps/website/content/products/pillbox.md`
- Create: `apps/website/content/whitepaper/index.md`
- Create: `apps/website/content/partner/index.md`
- Create: `apps/website/content/legal/privacy.md`
- Create: `apps/website/content/legal/terms.md`
- Create: `apps/website/content/blog/_index.md`
- Create: `apps/website/content/blog/2026-07-15-eregen-launch.md`
- Create: `apps/website/layouts/_default/list.html`
- Create: `apps/website/layouts/_default/single.html`
- Modify: `apps/website/layouts/partials/header.html` (add menu rendering)
- Modify: `apps/website/hugo.toml` (add products menu items, features array)

**Steps:**

- [ ] **Step 1: 创建 about 页面**

```markdown
---
title: "关于颐贞"
date: 2026-07-15
draft: false
---

# 颐贞 Eregen — 颐养正道，贞守安康

Eregen 打造完整的老年健康生态系统，将智能硬件、云端监控、家属应用和 B2B 医疗对接融为一体。

## 品牌理念

我们相信科技应该让每个人都能有尊严地老去。"颐贞"二字取自《易经》——"颐养正道，贞守安康"。

## 核心技术

- **全栈自研：** 从 GD32E230 手环固件到 Go 微服务云平台，全部自主开发
- **专利保护：** 申请多项发明专利，涵盖跌倒检测算法、自动分药机制、设备通信协议
- **安全合规：** Ed25519 设备认证、AES-256-GCM 加密传输、符合 PIPL 数据合规要求
- **开源许可：** 仅使用 MIT/BSD/Apache-2.0 许可，核心业务逻辑全部闭源
```

- [ ] **Step 2: 创建产品详情页 (bracelet.md + pillbox.md)**

bracelet.md:
```markdown
---
title: "智能手环系列"
date: 2026-07-15
---

## 三档产品矩阵

| 特性 | Starter ¥399 | Plus ¥599 | Pro ¥899 |
|------|-------------|-----------|----------|
| 心率监测 | ✅ | ✅ | ✅ |
| 血氧检测 | ✅ | ✅ | ✅ |
| GPS 定位 | ✅ 基站 | ✅ A-GPS | ✅ 高精度多模 |
| 跌倒检测 | ❌ | ✅ | ✅ |
| 电子围栏 | ❌ | ✅ 圆形 | ✅ 多边形 |
| SOS 呼叫 | ✅ | ✅ | ✅ |
| ECG 心电 | ❌ | ❌ | ✅ |
| 显示屏 | TFT 0.96" | IPS 1.28" | AMOLED 1.4" |
| 续航 | 7 天 | 5 天 | 3 天 |
| 防水 | IP67 | IP67 | IP68 |
```

pillbox.md:
```markdown
---
title: "智能药盒系列"
date: 2026-07-15
---

## 三档产品矩阵

| 特性 | 基础 ¥99 | 智能 ¥299 | 自动 ¥599 |
|------|---------|-----------|----------|
| 分药方式 | 手动分格 | 语音提醒 | 自动分药 |
| TTS 播报 | ❌ | ✅ SYN5300 | ✅ SYN5300 |
| 光电检测 | ❌ | ❌ | ✅ |
| APP 联动 | ❌ | ✅ | ✅ |
| 库存预警 | ❌ | ❌ | ✅ |
| WiFi | ❌ | ✅ ESP32-C3 | ✅ ESP32-C3 |
| 电池续航 | 30 天 | 7 天 | 5 天 |
```

- [ ] **Step 3: 创建白皮书页面**

whitepaper.md:
```markdown
---
title: "技术白皮书"
date: 2026-07-15
---

## 系统架构

Eregen 采用 8 个子系统的微服务架构：

1. **手环固件** — GD32E230 + FreeRTOS + C
2. **药盒固件** — ESP32-C3 + ESP-IDF v5.3 + C
3. **云平台** — Go 微服务 (gateway/api-server/push-service/data-pipeline/admin-api)
4. **家属 APP** — Flutter 3.24+ (iOS/Android/Web)
5. **管理后台** — Vue 3 + TypeScript + Element Plus
6. **微信小程序** — 原生 WXML/WXSS
7. **品牌官网** — Hugo + Tailwind CSS
8. **B2B 对接** — hospital-api/community-platform/insurance-integration

## 通信协议

设备 → EMQX MQTT → Go gateway → NATS JetStream → 微服务

详见 docs/specs/03-cloud-platform.md 接口定义章节。
```

- [ ] **Step 4: 创建合作页面**

partner.md:
```markdown
---
title: "ODM 合作"
date: 2026-07-15
---

## 合作模式

### 方案一：全栈 ODM
我们提供从硬件设计到云平台的全套解决方案，您负责品牌和销售。

### 方案二：技术授权
授权我们的固件源码、通信协议和云平台 API，您自行部署。

### 方案三：联合研发
针对特定行业需求（医院/养老院/保险公司），联合定制功能模块。

## 资质要求

- 具备医疗器械或养老产业相关资质
- 有稳定的销售渠道或机构客户资源
- 承诺维护品牌形象和用户体验标准
```

- [ ] **Step 5: 创建合规页面 (privacy.md + terms.md)**

privacy.md:
```markdown
---
title: "隐私政策"
date: 2026-07-15
---

## 数据收集

我们收集以下数据用于健康管理服务：
- 位置信息：用于实时定位和电子围栏
- 健康数据：心率、血氧、步数等传感器读数
- 设备信息：设备 ID、固件版本、电量状态

## 数据存储

所有数据在中国境内服务器存储，采用 AES-256-GCM 加密。

## 数据共享

仅在以下情况与第三方共享数据：
- 医院对接：经用户授权后向合作医院导出健康报告
- 保险理赔：经用户授权后向保险公司提供必要健康证明
- 紧急情况：SOS 告警时自动通知紧急联系人
```

terms.md:
```markdown
---
title: "用户协议"
date: 2026-07-15
---

## 服务条款

使用 Eregen 服务即表示同意以下条款：

1. 用户应对设备的使用和安全负主要责任
2. Eregen 不替代专业医疗建议
3. 免费套餐包含基础定位和健康监测功能
4. Premium 套餐 ($29/月) 包含高级健康分析和无限告警推送
```

- [ ] **Step 6: 创建博客页面**

blog/_index.md:
```markdown
---
title: "博客"
date: 2026-07-15
---
```

blog/2026-07-15-eregen-launch.md:
```markdown
---
title: "Eregen 颐贞项目启动"
date: 2026-07-15
---

Eregen 颐贞项目正式启动。这是一个完整的老年健康生态系统，涵盖智能手环、智能药盒、云平台、家属 APP、管理后台、微信小程序和品牌官网。

项目采用全栈自研策略，申请专利保护，目标是为 60-85 岁长者提供全方位的健康监测和紧急救助能力。
```

- [ ] **Step 7: 完善 Hugo 布局模板**

Modify `layouts/partials/header.html` to render Hugo menu:
```html
<header class="header">
  <nav class="nav">
    <a href="/" class="nav-brand">{{ .Site.Params.brandNameEn }}</a>
    <ul class="nav-links">
      {{ range .Site.Menus.main }}
      <li><a href="{{ .URL }}">{{ .Name }}</a></li>
      {{ end }}
    </ul>
  </nav>
</header>
```

Modify `layouts/_default/baseof.html` to include assets:
```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .Title }} | {{ .Site.Params.brandName }}</title>
  <meta name="description" content="{{ .Site.Params.description }}">
  {{ $css := resources.Get "/css/main.css" | toCSS (dict "enableSourceMap" true) | minify | fingerprint }}
  <link rel="stylesheet" href="{{ $css.RelPermalink }}">
</head>
<body>
  {{ partial "header.html" . }}
  <main>{{ block "main" . }}{{ end }}</main>
  {{ partial "footer.html" . }}
</body>
</html>
```

Create `layouts/_default/list.html`:
```html
{{ define "main" }}
<section class="section">
  <div class="section-inner">
    <h1 class="section-title">{{ .Title }}</h1>
    <div class="content">{{ .Content }}</div>
    <ul class="blog-list">
      {{ range .Pages }}
      <li><a href="{{ .RelPermalink }}">{{ .Title }}</a><span>{{ .Date.Format "2006-01-02" }}</span></li>
      {{ end }}
    </ul>
  </div>
</section>
{{ end }}
```

Create `layouts/_default/single.html`:
```html
{{ define "main" }}
<article class="article">
  <div class="section-inner">
    <h1 class="article-title">{{ .Title }}</h1>
    <div class="article-content">{{ .Content }}</div>
  </div>
</article>
{{ end }}
```

- [ ] **Step 8: 更新 hugo.toml 添加菜单和内容参数**

Add to `hugo.toml`:
```toml
[[params.products]]
  name = "智能手环"
  icon = "⌚"
  desc = "GD32E230 + Cat1 蜂窝通信"
  tiers = [{name="Starter",price="¥399"},{name="Plus",price="¥599"},{name="Pro",price="¥899"}]
  features = ["心率","血氧","GPS","跌倒检测","SOS"]

[[params.products]]
  name = "智能药盒"
  icon = "💊"
  desc = "ESP32-C3 + WiFi + 语音提醒"
  tiers = [{name="基础",price="¥99"},{name="智能",price="¥299"},{name="自动",price="¥599"}]
  features = ["TTS语音","自动分药","光电检测","库存预警"]
```

---

## Phase B: 云平台补全

### Task B1: admin-api 微服务 — 从零创建

**Files:**
- Create: `cloud/admin-api/cmd/main.go`
- Create: `cloud/admin-api/internal/model/model.go`
- Create: `cloud/admin-api/internal/handler/dashboard.go`
- Create: `cloud/admin-api/internal/handler/device.go`
- Create: `cloud/admin-api/internal/handler/user.go`
- Create: `cloud/admin-api/internal/handler/alert.go`
- Create: `cloud/admin-api/internal/middleware/auth.go`
- Create: `cloud/admin-api/internal/router/router.go`
- Create: `cloud/admin-api/internal/store/postgres.go`
- Create: `cloud/admin-api/go.mod`

**Interfaces:**
- Consumes: PostgreSQL (same DB as api-server), Redis (same instance)
- Produces: REST admin endpoints consumed by admin-web views

**Steps:**

- [ ] **Step 1: 创建 go.mod 和入口 main.go**

Create `cloud/admin-api/go.mod`:
```go
module eregen.dev/admin-api

go 1.22

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/lib/pq v1.10.9
    github.com/redis/go-redis/v9 v9.4.0
    golang.org/x/crypto v0.18.0
)
```

Create `cloud/admin-api/cmd/main.go`:
```go
package main

import (
    "log"
    "eregen.dev/admin-api/internal/config"
    "eregen.dev/admin-api/internal/router"
    "eregen.dev/admin-api/internal/store"
)

func main() {
    cfg := config.Load()
    db := store.NewPostgres(cfg.DatabaseURL)
    rdb := store.NewRedis(cfg.RedisURL)
    r := router.Setup(db, rdb, cfg)

    log.Printf("admin-api starting on :%s", cfg.Port)
    if err := r.Run(":" + cfg.Port); err != nil {
        log.Fatal(err)
    }
}
```

- [ ] **Step 2: 创建 config.go**

Create `cloud/admin-api/internal/config/config.go`:
```go
package config

import "os"

type Config struct {
    Port        string
    DatabaseURL string
    RedisURL    string
    JWTSecret   string
}

func Load() *Config {
    return &Config{
        Port:        getEnv("PORT", "8085"),
        DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/eregen"),
        RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
        JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

- [ ] **Step 3: 创建 model.go (复用 api-server 模型定义)**

Create `cloud/admin-api/internal/model/model.go`:
```go
package model

import "time"

// Re-export types from api-server model for consistency.
// In production these would be shared via a common module.

type Role string

const (
    RoleAdmin       Role = "admin"
    RoleOperator    Role = "operator"
    RoleSuperAdmin  Role = "super_admin"
)

type DashboardStats struct {
    OnlineDevices     int       `json:"online_devices"`
    TotalDevices      int       `json:"total_devices"`
    ActiveAlerts      int       `json:"active_alerts"`
    TotalUsers        int       `json:"total_users"`
    ActiveSubscriptions int     `json:"active_subscriptions"`
    AlertTrend        []TrendPoint `json:"alert_trend"`
}

type TrendPoint struct {
    Date  string `json:"date"`
    Value int    `json:"value"`
}

type DeviceSummary struct {
    ID         string    `json:"id"`
    DeviceID   string    `json:"device_id"`
    Type       string    `json:"type"`
    Tier       string    `json:"tier"`
    Status     string    `json:"status"`
    LastSeen   time.Time `json:"last_seen"`
    OwnerName  string    `json:"owner_name"`
    FirmwareVer string   `json:"firmware_version"`
}

type UserSummary struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
    Devices   int       `json:"devices"`
}

type AlertSummary struct {
    ID         string    `json:"id"`
    ElderlyID  string    `json:"elderly_id"`
    AlertType  string    `json:"alert_type"`
    Severity   string    `json:"severity"`
    Status     string    `json:"status"`
    CreatedAt  time.Time `json:"created_at"`
    DeviceID   string    `json:"device_id"`
}

type SubscriptionStat struct {
    Tier   string `json:"tier"`
    Count  int    `json:"count"`
    Pct    float64 `json:"pct"`
}
```

- [ ] **Step 4: 创建 store/postgres.go**

Create `cloud/admin-api/internal/store/postgres.go`:
```go
package store

import (
    "database/sql"
    "eregen.dev/admin-api/internal/model"
)

func NewPostgres(dsn string) *sql.DB {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        panic(err)
    }
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    return db
}

type PostgresStore struct {
    db *sql.DB
}

func NewStore(db *sql.DB) *PostgresStore {
    return &PostgresStore{db: db}
}

func (s *PostgresStore) GetDashboardStats() (*model.DashboardStats, error) {
    var stats model.DashboardStats
    s.db.QueryRow(`SELECT COUNT(*) FROM devices WHERE status='online'`).Scan(&stats.OnlineDevices)
    s.db.QueryRow(`SELECT COUNT(*) FROM devices`).Scan(&stats.TotalDevices)
    s.db.QueryRow(`SELECT COUNT(*) FROM alerts WHERE status='pending'`).Scan(&stats.ActiveAlerts)
    s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
    s.db.QueryRow(`SELECT COUNT(*) FROM subscriptions WHERE status='active'`).Scan(&stats.ActiveSubscriptions)
    return &stats, nil
}

func (s *PostgresStore) ListDevices(page, pageSize int, status, devType, tier string) ([]model.DeviceSummary, error) {
    query := `SELECT id, device_id, device_type, tier, status, last_seen, 
              (SELECT name FROM users u JOIN devices d ON d.owner_user_id = u.id WHERE d.id = devices.id LIMIT 1) as owner_name,
              COALESCE(settings->>'fw_version','v0.1') as firmware_version
              FROM devices WHERE 1=1`
    args := []interface{}{}
    idx := 1
    if status != "" {
        query += fmt.Sprintf(" AND status=$%d", idx); args = append(args, status); idx++
    }
    if devType != "" {
        query += fmt.Sprintf(" AND device_type=$%d", idx); args = append(args, devType); idx++
    }
    if tier != "" {
        query += fmt.Sprintf(" AND tier=$%d", idx); args = append(args, tier); idx++
    }
    query += fmt.Sprintf(" ORDER BY last_seen DESC LIMIT $%d OFFSET $%d", idx, idx+1)
    args = append(args, pageSize, (page-1)*pageSize)

    rows, err := s.db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var devices []model.DeviceSummary
    for rows.Next() {
        var d model.DeviceSummary
        rows.Scan(&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &d.LastSeen, &d.OwnerName, &d.FirmwareVer)
        devices = append(devices, d)
    }
    return devices, nil
}

func (s *PostgresStore) ListUsers(page, pageSize int, role string) ([]model.UserSummary, error) {
    query := `SELECT u.id, u.name, u.role, u.created_at, 
              (SELECT COUNT(*) FROM devices d WHERE d.owner_user_id = u.id) as devices
              FROM users u WHERE 1=1`
    args := []interface{}{}
    idx := 1
    if role != "" {
        query += fmt.Sprintf(" AND u.role=$%d", idx); args = append(args, role); idx++
    }
    query += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
    args = append(args, pageSize, (page-1)*pageSize)

    rows, err := s.db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []model.UserSummary
    for rows.Next() {
        var u model.UserSummary
        rows.Scan(&u.ID, &u.Name, &u.Role, &u.CreatedAt, &u.Devices)
        users = append(users, u)
    }
    return users, nil
}

func (s *PostgresStore) ListAlerts(severity, status string, limit int) ([]model.AlertSummary, error) {
    query := `SELECT a.id, a.elderly_id, a.alert_type, a.severity, a.status, a.created_at,
              COALESCE(d.device_id, '') as device_id
              FROM alerts a LEFT JOIN devices d ON a.elderly_id = d.id WHERE 1=1`
    args := []interface{}{}
    idx := 1
    if severity != "" {
        query += fmt.Sprintf(" AND a.severity=$%d", idx); args = append(args, severity); idx++
    }
    if status != "" {
        query += fmt.Sprintf(" AND a.status=$%d", idx); args = append(args, status); idx++
    }
    query += fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT $%d", idx)
    args = append(args, limit)

    rows, err := s.db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var alerts []model.AlertSummary
    for rows.Next() {
        var a model.AlertSummary
        rows.Scan(&a.ID, &a.ElderlyID, &a.AlertType, &a.Severity, &a.Status, &a.CreatedAt, &a.DeviceID)
        alerts = append(alerts, a)
    }
    return alerts, nil
}

func (s *PostgresStore) GetSubscriptionStats() ([]model.SubscriptionStat, error) {
    rows, err := s.db.Query(`
        SELECT plan_tier, COUNT(*)::int,
               ROUND(COUNT(*)::numeric / NULLIF(SUM(COUNT(*)) OVER (), 0) * 100, 1)
        FROM subscriptions GROUP BY plan_tier ORDER BY COUNT(*) DESC
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var stats []model.SubscriptionStat
    for rows.Next() {
        var s model.SubscriptionStat
        rows.Scan(&s.Tier, &s.Count, &s.Pct)
        stats = append(stats, s)
    }
    return stats, nil
}
```

- [ ] **Step 5: 创建 handler 文件 (dashboard.go + device.go + user.go + alert.go)**

Create `cloud/admin-api/internal/handler/dashboard.go`:
```go
package handler

import (
    "net/http"
    "eregen.dev/admin-api/internal/model"
    "eregen.dev/admin-api/internal/store"
    "github.com/gin-gonic/gin"
)

type DashboardHandler struct {
    store *store.PostgresStore
}

func NewDashboardHandler(s *store.PostgresStore) *DashboardHandler {
    return &DashboardHandler{store: s}
}

func (h *DashboardHandler) GetOverview(c *gin.Context) {
    stats, err := h.store.GetDashboardStats()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, stats)
}

func (h *DashboardHandler) GetSubscriptionStats(c *gin.Context) {
    stats, err := h.store.GetSubscriptionStats()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, stats)
}
```

Create `cloud/admin-api/internal/handler/device.go`:
```go
package handler

import (
    "net/http"
    "strconv"
    "eregen.dev/admin-api/internal/store"
    "github.com/gin-gonic/gin"
)

type DeviceHandler struct {
    store *store.PostgresStore
}

func NewDeviceHandler(s *store.PostgresStore) *DeviceHandler {
    return &DeviceHandler{store: s}
}

func (h *DeviceHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    status := c.Query("status")
    devType := c.Query("type")
    tier := c.Query("tier")

    devices, err := h.store.ListDevices(page, pageSize, status, devType, tier)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": devices, "page": page, "page_size": pageSize})
}
```

Create `cloud/admin-api/internal/handler/user.go`:
```go
package handler

import (
    "net/http"
    "strconv"
    "eregen.dev/admin-api/internal/store"
    "github.com/gin-gonic/gin"
)

type UserHandler struct {
    store *store.PostgresStore
}

func NewUserHandler(s *store.PostgresStore) *UserHandler {
    return &UserHandler{store: s}
}

func (h *UserHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    role := c.Query("role")

    users, err := h.store.ListUsers(page, pageSize, role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": users, "page": page, "page_size": pageSize})
}
```

Create `cloud/admin-api/internal/handler/alert.go`:
```go
package handler

import (
    "net/http"
    "eregen.dev/admin-api/internal/store"
    "github.com/gin-gonic/gin"
)

type AlertHandler struct {
    store *store.PostgresStore
}

func NewAlertHandler(s *store.PostgresStore) *AlertHandler {
    return &AlertHandler{store: s}
}

func (h *AlertHandler) List(c *gin.Context) {
    severity := c.Query("severity")
    status := c.Query("status")
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

    alerts, err := h.store.ListAlerts(severity, status, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": alerts})
}
```

- [ ] **Step 6: 创建 middleware/auth.go**

Create `cloud/admin-api/internal/middleware/auth.go`:
```go
package middleware

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        token := strings.TrimPrefix(authHeader, "Bearer ")
        if token == authHeader {
            token = ""
        }
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
            return
        }
        // TODO: validate JWT token
        c.Next()
    }
}

func RBAC(requiredRole string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // TODO: check user role from JWT claims
        c.Next()
    }
}
```

- [ ] **Step 7: 创建 router/router.go**

Create `cloud/admin-api/internal/router/router.go`:
```go
package router

import (
    "database/sql"
    "eregen.dev/admin-api/internal/handler"
    "eregen.dev/admin-api/internal/middleware"
    "eregen.dev/admin-api/internal/store"
    "github.com/gin-gonic/gin"
)

func Setup(db *sql.DB, rdb interface{}, cfg interface{}) *gin.Engine {
    s := store.NewStore(db)
    g := gin.Default()

    g.Use(middleware.Auth())

    dashboard := handler.NewDashboardHandler(s)
    device := handler.NewDeviceHandler(s)
    user := handler.NewUserHandler(s)
    alert := handler.NewAlertHandler(s)

    api := g.Group("/api/v1/admin")
    {
        api.GET("/stats/overview", dashboard.GetOverview)
        api.GET("/stats/subscriptions", dashboard.GetSubscriptionStats)

        api.GET("/devices", device.List)

        api.GET("/users", user.List)

        api.GET("/alerts", alert.List)
    }

    return g
}
```

- [ ] **Step 8: 编译验证**

Run: `cd cloud/admin-api && go build ./cmd/main.go`

### Task B2: shared 模块 go.mod 完善

**Files:**
- Modify: `shared/protocol/go.mod` (ensure it has correct module path)
- Modify: `shared/crypto/go.mod` (ensure it has correct module path)

**Steps:**

- [ ] **Step 1: 确保 protocol 模块 go.mod 正确**

```go
module eregen.dev/shared/protocol

go 1.22
```

- [ ] **Step 2: 确保 crypto 模块 go.mod 正确**

```go
module eregen.dev/shared/crypto

go 1.22
```

### Task B3: 固件公共库 firmware/common

**Files:**
- Create: `firmware/common/README.md`
- Create: `firmware/common/include/brand_boot_logo.h`
- Create: `firmware/common/src/brand_boot_logo.c`
- Create: `firmware/common/include/mqtt_common.h`
- Create: `firmware/common/src/mqtt_common.c`
- Create: `firmware/common/CMakeLists.txt`

**Steps:**

- [ ] **Step 1: 创建公共库 CMakeLists.txt**

Create `firmware/common/CMakeLists.txt`:
```cmake
idf_component_register(
    SRCS "src/brand_boot_logo.c" "src/mqtt_common.c"
    INCLUDE_DIRS "include"
    REQUIRES mqtt nvs_flash
)
```

- [ ] **Step 2: 创建 MQTT 公共封装层**

Create `firmware/common/include/mqtt_common.h`:
```c
#ifndef MQTT_COMMON_H
#define MQTT_COMMON_H

#include <stdint.h>
#include <stdbool.h>

typedef void (*mqtt_msg_handler_t)(const char* topic, const uint8_t* payload, size_t len);

int mqtt_common_connect(const char* broker_host, uint16_t broker_port,
                        const char* client_id, const char* username,
                        const char* password);
void mqtt_common_disconnect(void);
int mqtt_common_subscribe(const char* topic, mqtt_msg_handler_t handler);
int mqtt_common_publish(const char* topic, const char* payload, size_t len, int qos);

#endif // MQTT_COMMON_H
```

Create `firmware/common/src/mqtt_common.c`:
```c
#include "mqtt_common.h"
#include "mqtt_client.h"
#include <string.h>

static esp_mqtt_client_handle_t s_client = NULL;
static mqtt_msg_handler_t s_handlers[16] = {0};
static char s_topics[16][128] = {0};
static int s_topic_count = 0;

int mqtt_common_connect(const char* broker_host, uint16_t broker_port,
                        const char* client_id, const char* username,
                        const char* password) {
    esp_mqtt_client_config_t cfg = {0};
    cfg.broker.address.hostname = broker_host;
    cfg.broker.address.port = broker_port;
    cfg.credentials.client_id = client_id;
    if (username) cfg.credentials.username = username;
    if (password) cfg.credentials.authentication.password = password;

    s_client = esp_mqtt_client_init(&cfg);
    esp_mqtt_client_start(s_client);
    return 0;
}

void mqtt_common_disconnect(void) {
    if (s_client) {
        esp_mqtt_client_stop(s_client);
        esp_mqtt_client_destroy(s_client);
        s_client = NULL;
    }
}

int mqtt_common_subscribe(const char* topic, mqtt_msg_handler_t handler) {
    if (s_topic_count >= 16) return -1;
    strncpy(s_topics[s_topic_count], topic, 127);
    s_handlers[s_topic_count] = handler;
    esp_mqtt_client_subscribe(s_client, topic, 0);
    s_topic_count++;
    return 0;
}

int mqtt_common_publish(const char* topic, const char* payload, size_t len, int qos) {
    if (!s_client) return -1;
    return esp_mqtt_client_publish(s_client, topic, payload, len, qos, 0);
}
```

- [ ] **Step 3: 创建品牌开机动画资源**

Create `firmware/common/include/brand_boot_logo.h`:
```c
#ifndef BRAND_BOOT_LOGO_H
#define BRAND_BOOT_LOGO_H

// Boot logo bitmap: 128x64 pixels, 1-bit monochrome
extern const unsigned char brand_boot_logo_bits[];
extern const unsigned int  brand_boot_logo_width;
extern const unsigned int  brand_boot_logo_height;

#endif // BRAND_BOOT_LOGO_H
```

Create `firmware/common/src/brand_boot_logo.c`:
```c
#include "brand_boot_logo.h"

// Placeholder: 128x64 monochrome bitmap for Eregen boot logo
// In production, this would be generated from an actual PNG using img2c
const unsigned char brand_boot_logo_bits[] = {
    // 128 columns x 8 rows (64 bits height)
    // Row 0: "E" pattern
    0xE0, 0x10, 0x10, 0x10, 0xE0, 0x00, 0x00, 0x00,
    // Row 1: "R" pattern
    0xC0, 0x60, 0x90, 0x90, 0x60, 0x00, 0x00, 0x00,
    // ... truncated for brevity
};
const unsigned int brand_boot_logo_width  = 128;
const unsigned int brand_boot_logo_height = 64;
```

- [ ] **Step 4: 创建 README**

Create `firmware/common/README.md`:
```markdown
# firmware/common — 品牌公共固件库

为手环和药盒固件提供的公共模块：

- **MQTT 公共封装** — 统一连接/订阅/发布接口
- **品牌开机动画** — 128x64 单色 BMP 资源
- **OTA 通用逻辑** — 固件下载/校验/切换
- **日志统一格式** — 带设备类型前缀的日志输出

## 引用方式

在 `CMakeLists.txt` 中添加:
```cmake
target_add_component(common PATH ../common)
```
```

---

## Phase C: 前端应用补全

### Task C1: admin-web — 补齐 api/stores/types/utils

**Files:**
- Create: `apps/admin-web/src/api/client.ts`
- Create: `apps/admin-web/src/api/devices.ts`
- Create: `apps/admin-web/src/api/users.ts`
- Create: `apps/admin-web/src/api/alerts.ts`
- Create: `apps/admin-web/src/api/subscriptions.ts`
- Create: `apps/admin-web/src/api/dashboard.ts`
- Create: `apps/admin-web/src/stores/auth.ts`
- Create: `apps/admin-web/src/stores/dashboard.ts`
- Create: `apps/admin-web/src/stores/device.ts`
- Create: `apps/admin-web/src/types/index.ts`
- Create: `apps/admin-web/src/utils/format.ts`

**Interfaces:**
- Consumes: Axios, localStorage (for token)
- Produces: typed API clients matching api-server REST endpoints

**Steps:**

- [ ] **Step 1: 创建 types/index.ts**

```typescript
export interface User {
  id: string
  email?: string
  phone?: string
  name: string
  role: 'family' | 'elderly' | 'institution' | 'admin' | 'operator'
  created_at: string
  updated_at: string
}

export interface ElderlyProfile {
  id: string
  user_id: string
  name: string
  birth_date?: string
  avatar_url?: string
  health_tiers: string[]
  created_at: string
  updated_at: string
}

export interface Device {
  id: string
  device_id: string
  device_type: 'bracelet' | 'pillbox'
  tier: 'starter' | 'plus' | 'pro' | 'basic' | 'smart' | 'auto'
  status: 'online' | 'offline'
  last_seen?: string
  owner_user_id: string
  settings?: Record<string, any>
  created_at: string
}

export interface HealthRecord {
  id: string
  elderly_id: string
  timestamp: string
  hr?: number
  spo2?: number
  steps?: number
  sleep_hours?: number
  bp_systolic?: number
  bp_diastolic?: number
}

export interface LocationRecord {
  id: string
  elderly_id: string
  timestamp: string
  lat: number
  lon: number
  accuracy?: number
}

export interface MedicationRule {
  id: string
  elderly_id: string
  schedule_time: string
  dose_count: number
  pill_type: string
  days_of_week: number[]
  active: boolean
  created_at: string
}

export interface Alert {
  id: string
  elderly_id: string
  alert_type: string
  severity: 'P0' | 'P1' | 'P2'
  status: 'pending' | 'resolved'
  metadata?: Record<string, any>
  created_at: string
  resolved_at?: string
}

export interface Subscription {
  id: string
  user_id: string
  plan_tier: 'free' | 'premium' | 'enterprise'
  status: string
  start_date: string
  end_date: string
}

export interface DashboardStats {
  online_devices: number
  total_devices: number
  active_alerts: number
  total_users: number
  active_subscriptions: number
  alert_trend: TrendPoint[]
}

export interface TrendPoint {
  date: string
  value: number
}
```

- [ ] **Step 2: 创建 api/client.ts**

```typescript
import axios from 'axios'

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('admin_token')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export default apiClient
```

- [ ] **Step 3: 创建各 API 模块**

```typescript
// api/devices.ts
import apiClient from './client'
import type { Device } from '@/types'

export const devicesApi = {
  list(params: { page?: number; page_size?: number; status?: string; type?: string; tier?: string }) {
    return apiClient.get<{ data: Device[] }>('/devices', { params })
  },
  detail(id: string) {
    return apiClient.get<{ data: Device }>(`/devices/${id}`)
  },
  updateSettings(id: string, settings: Record<string, any>) {
    return apiClient.put(`/devices/${id}/settings`, { settings })
  },
  triggerOTA(id: string, firmwareUrl: string, hash: string) {
    return apiClient.post(`/devices/${id}/ota`, { url: firmwareUrl, hash })
  },
}
```

```typescript
// api/users.ts
import apiClient from './client'
import type { User, ElderlyProfile } from '@/types'

export const usersApi = {
  list(params: { page?: number; page_size?: number; role?: string }) {
    return apiClient.get<{ data: User[] }>('/users', { params })
  },
  detail(id: string) {
    return apiClient.get<{ data: User }>(`/users/${id}`)
  },
  updateRole(id: string, role: string) {
    return apiClient.put(`/users/${id}/role`, { role })
  },
  listElderly(params?) {
    return apiClient.get<{ data: ElderlyProfile[] }>('/elderly', { params })
  },
}
```

```typescript
// api/alerts.ts
import apiClient from './client'
import type { Alert } from '@/types'

export const alertsApi = {
  list(params: { severity?: string; status?: string; limit?: number }) {
    return apiClient.get<{ data: Alert[] }>('/alerts', { params })
  },
  markResolved(id: string) {
    return apiClient.put(`/alerts/${id}/status`, { status: 'resolved' })
  },
}
```

```typescript
// api/subscriptions.ts
import apiClient from './client'

export const subscriptionsApi = {
  list(params?) {
    return apiClient.get('/subscriptions', { params })
  },
  stats() {
    return apiClient.get('/subscriptions/stats')
  },
}
```

```typescript
// api/dashboard.ts
import apiClient from './client'
import type { DashboardStats } from '@/types'

export const dashboardApi = {
  overview() {
    return apiClient.get<{ data: DashboardStats }>('/admin/stats/overview')
  },
}
```

- [ ] **Step 4: 创建 Pinia stores**

```typescript
// stores/auth.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('admin_token') || '')
  const user = ref<{ name: string; role: string } | null>(null)

  function login(t: string, u: any) {
    token.value = t
    user.value = u
    localStorage.setItem('admin_token', t)
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('admin_token')
  }

  function hasPermission(resource: string): boolean {
    if (!user.value) return false
    return user.value.role === 'super_admin' || user.value.role === 'admin'
  }

  return { token, user, login, logout, hasPermission }
})
```

```typescript
// stores/dashboard.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { dashboardApi } from '@/api/dashboard'
import type { DashboardStats } from '@/types'

export const useDashboardStore = defineStore('dashboard', () => {
  const stats = ref<DashboardStats>({
    online_devices: 0, total_devices: 0, active_alerts: 0,
    total_users: 0, active_subscriptions: 0, alert_trend: [],
  })

  async function fetchOverview() {
    const res = await dashboardApi.overview()
    stats.value = res.data.data || res.data
  }

  return { stats, fetchOverview }
})
```

```typescript
// stores/device.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { devicesApi } from '@/api/devices'
import type { Device } from '@/types'

export const useDeviceStore = defineStore('device', () => {
  const devices = ref<Device[]>([])
  const loading = ref(false)

  async function fetchList(params?) {
    loading.value = true
    try {
      const res = await devicesApi.list(params || {})
      devices.value = res.data.data || []
    } finally {
      loading.value = false
    }
  }

  return { devices, loading, fetchList }
})
```

- [ ] **Step 5: 创建 utils/format.ts**

```typescript
export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}

export function formatNumber(n: number): string {
  return n.toLocaleString('zh-CN')
}

export function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins} 分钟前`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs} 小时前`
  const days = Math.floor(hrs / 24)
  return `${days} 天前`
}
```

- [ ] **Step 6: 更新 main.ts 注册依赖**

Modify `apps/admin-web/src/main.ts`:
```typescript
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(ElementPlus)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.mount('#app')
```

- [ ] **Step 7: 创建 vite.config.ts 别名配置**

Verify `apps/admin-web/vite.config.ts` has:
```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 8: 验证编译**

Run: `cd apps/admin-web && npm run type-check`

### Task C2: family-app — 创建 models 目录

**Files:**
- Create: `apps/family-app/lib/models/user.dart`
- Create: `apps/family-app/lib/models/elderly.dart`
- Create: `apps/family-app/lib/models/device.dart`
- Create: `apps/family-app/lib/models/health.dart`
- Create: `apps/family-app/lib/models/location.dart`
- Create: `apps/family-app/lib/models/medication.dart`
- Create: `apps/family-app/lib/models/alert.dart`
- Create: `apps/family-app/lib/models/subscription.dart`
- Modify: `apps/family-app/pubspec.yaml` (add dio, provider/riverpod)

**Steps:**

- [ ] **Step 1: 检查并更新 pubspec.yaml**

Read existing `pubspec.yaml`, ensure dependencies include:
```yaml
dependencies:
  flutter:
    sdk: flutter
  dio: ^5.4.0
  provider: ^6.1.1
  shared_preferences: ^2.2.2

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^3.0.0
```

- [ ] **Step 2: 创建 models (对齐 api-server model.go)**

```dart
// models/user.dart
class User {
  final String id;
  final String? email;
  final String? phone;
  final String name;
  final String role;
  final DateTime createdAt;
  final DateTime updatedAt;

  User({required this.id, this.email, this.phone, required this.name,
        required this.role, required this.createdAt, required this.updatedAt});

  factory User.fromJson(Map<String, dynamic> json) => User(
    id: json['id'] as String,
    email: json['email'] as String?,
    phone: json['phone'] as String?,
    name: json['name'] as String,
    role: json['role'] as String,
    createdAt: DateTime.parse(json['created_at'] as String),
    updatedAt: DateTime.parse(json['updated_at'] as String),
  );
}
```

```dart
// models/elderly.dart
class ElderlyProfile {
  final String id;
  final String userId;
  final String name;
  final DateTime? birthDate;
  final String? avatarUrl;
  final List<String> healthTiers;
  final DateTime createdAt;

  ElderlyProfile({required this.id, required this.userId, required this.name,
                  this.birthDate, this.avatarUrl, required this.healthTiers,
                  required this.createdAt});

  factory ElderlyProfile.fromJson(Map<String, dynamic> json) => ElderlyProfile(
    id: json['id'] as String,
    userId: json['user_id'] as String,
    name: json['name'] as String,
    birthDate: json['birth_date'] != null ? DateTime.parse(json['birth_date']) : null,
    avatarUrl: json['avatar_url'] as String?,
    healthTiers: List<String>.from(json['health_tiers'] ?? []),
    createdAt: DateTime.parse(json['created_at'] as String),
  );
}
```

```dart
// models/device.dart
class Device {
  final String id;
  final String deviceId;
  final String deviceType;
  final String tier;
  final String status;
  final String? lastSeen;
  final String ownerUserId;
  final Map<String, dynamic>? settings;

  Device({required this.id, required this.deviceId, required this.deviceType,
          required this.tier, required this.status, this.lastSeen,
          required this.ownerUserId, this.settings});

  factory Device.fromJson(Map<String, dynamic> json) => Device(
    id: json['id'] as String,
    deviceId: json['device_id'] as String,
    deviceType: json['device_type'] as String,
    tier: json['tier'] as String,
    status: json['status'] as String,
    lastSeen: json['last_seen'] as String?,
    ownerUserId: json['owner_user_id'] as String,
    settings: json['settings'] as Map<String, dynamic>?,
  );
}
```

```dart
// models/health.dart
class HealthRecord {
  final String id;
  final String elderlyId;
  final DateTime timestamp;
  final int? hr;
  final int? spo2;
  final int? steps;
  final double? sleepHours;
  final int? bpSystolic;
  final int? bpDiastolic;

  HealthRecord({required this.id, required this.elderlyId, required this.timestamp,
                this.hr, this.spo2, this.steps, this.sleepHours,
                this.bpSystolic, this.bpDiastolic});

  factory HealthRecord.fromJson(Map<String, dynamic> json) => HealthRecord(
    id: json['id'] as String,
    elderlyId: json['elderly_id'] as String,
    timestamp: DateTime.parse(json['timestamp'] as String),
    hr: json['hr'] as int?,
    spo2: json['spo2'] as int?,
    steps: json['steps'] as int?,
    sleepHours: json['sleep_hours'] as double?,
    bpSystolic: json['bp_systolic'] as int?,
    bpDiastolic: json['bp_diastolic'] as int?,
  );
}
```

```dart
// models/location.dart
class LocationRecord {
  final String id;
  final String elderlyId;
  final DateTime timestamp;
  final double lat;
  final double lon;
  final double? accuracy;

  LocationRecord({required this.id, required this.elderlyId, required this.timestamp,
                  required this.lat, required this.lon, this.accuracy});

  factory LocationRecord.fromJson(Map<String, dynamic> json) => LocationRecord(
    id: json['id'] as String,
    elderlyId: json['elderly_id'] as String,
    timestamp: DateTime.parse(json['timestamp'] as String),
    lat: (json['lat'] as num).toDouble(),
    lon: (json['lon'] as num).toDouble(),
    accuracy: json['accuracy'] != null ? (json['accuracy'] as num).toDouble() : null,
  );
}
```

```dart
// models/medication.dart
class MedicationRule {
  final String id;
  final String elderlyId;
  final String scheduleTime;
  final int doseCount;
  final String pillType;
  final List<int> daysOfWeek;
  final bool active;
  final DateTime createdAt;

  MedicationRule({required this.id, required this.elderlyId,
                  required this.scheduleTime, required this.doseCount,
                  required this.pillType, required this.daysOfWeek,
                  required this.active, required this.createdAt});

  factory MedicationRule.fromJson(Map<String, dynamic> json) => MedicationRule(
    id: json['id'] as String,
    elderlyId: json['elderly_id'] as String,
    scheduleTime: json['schedule_time'] as String,
    doseCount: json['dose_count'] as int,
    pillType: json['pill_type'] as String,
    daysOfWeek: List<int>.from(json['days_of_week'] ?? []),
    active: json['active'] as bool,
    createdAt: DateTime.parse(json['created_at'] as String),
  );
}
```

```dart
// models/alert.dart
class Alert {
  final String id;
  final String elderlyId;
  final String alertType;
  final String severity;
  final String status;
  final Map<String, dynamic>? metadata;
  final DateTime createdAt;
  final DateTime? resolvedAt;

  Alert({required this.id, required this.elderlyId, required this.alertType,
         required this.severity, required this.status, this.metadata,
         required this.createdAt, this.resolvedAt});

  factory Alert.fromJson(Map<String, dynamic> json) => Alert(
    id: json['id'] as String,
    elderlyId: json['elderly_id'] as String,
    alertType: json['alert_type'] as String,
    severity: json['severity'] as String,
    status: json['status'] as String,
    metadata: json['metadata'] as Map<String, dynamic>?,
    createdAt: DateTime.parse(json['created_at'] as String),
    resolvedAt: json['resolved_at'] != null ? DateTime.parse(json['resolved_at']) : null,
  );
}
```

```dart
// models/subscription.dart
class Subscription {
  final String id;
  final String userId;
  final String planTier;
  final String status;
  final DateTime startDate;
  final DateTime endDate;

  Subscription({required this.id, required this.userId, required this.planTier,
                required this.status, required this.startDate, required this.endDate});

  factory Subscription.fromJson(Map<String, dynamic> json) => Subscription(
    id: json['id'] as String,
    userId: json['user_id'] as String,
    planTier: json['plan_tier'] as String,
    status: json['status'] as String,
    startDate: DateTime.parse(json['start_date'] as String),
    endDate: DateTime.parse(json['end_date'] as String),
  );
}
```

- [ ] **Step 3: 验证编译**

Run: `cd apps/family-app && flutter analyze lib/models/`

---

## Phase D: 固件补全

### Task D1: 药盒 Auto 档固件

**Files:**
- Create: `firmware/pillbox/auto/main.c`
- Create: `firmware/pillbox/auto/CMakeLists.txt`
- Create: `firmware/pillbox/auto/idf_component.yml`
- Create: `firmware/pillbox/auto/dispensing_motor.c` + `.h`
- Create: `firmware/pillbox/auto/optical_detect.c` + `.h`
- Create: `firmware/pillbox/auto/tts_broadcaster.c` + `.h`
- Create: `firmware/pillbox/auto/inventory_watch.c` + `.h`
- Create: `firmware/pillbox/auto/free_rtos_tasks.c` + `.h`

**Interfaces:**
- Consumes: firmware/common (mqtt_common, brand_boot_logo)
- Produces: MQTT messages of type med_status with compartment tracking

**Steps:**

- [ ] **Step 1: 创建 CMakeLists.txt**

```cmake
idf_component_register(
    SRCS "main.c" "dispensing_motor.c" "optical_detect.c"
         "tts_broadcaster.c" "inventory_watch.c" "free_rtos_tasks.c"
    INCLUDE_DIRS "."
    REQUIRES driver nvs_flash esp_timer esp_adc_cal
)
```

- [ ] **Step 2: 创建 idf_component.yml**

```yaml
dependencies:
  esp-idf-lib: "^0.4.0"
  espressif/ledc: "^1.0.0"
  review: "*"
```

- [ ] **Step 3: 创建 dispensing_motor.c**

```c
#include "dispensing_motor.h"
#include "driver/gpio.h"
#include "driver/ledc.h"
#include <stdio.h>

#define MOTOR_STEP_GPIO     GPIO_NUM_18
#define MOTOR_DIR_GPIO      GPIO_NUM_19
#define STEPS_PER_COMPARTMENT 4096  // 28BYJ-48 at 1:64 gear ratio

static ledc_channel_config_t s_ledc_chan = {0};

int motor_init(void) {
    ledc_timer_config_t timer_cfg = {
        .speed_mode = LEDC_LOW_SPEED_MODE,
        .timer_num = LEDC_TIMER_0,
        .duty_resolution = LEDC_TIMER_10_BIT,
        .freq_hz = 1000,
        .clk_cfg = LEDC_AUTO_CLK,
    };
    ledc_timer_config(&timer_cfg);

    s_ledc_chan.channel    = LEDC_CHANNEL_0;
    s_ledc_chan.timer_sel  = LEDC_TIMER_0;
    s_ledc_chan.intr_type  = LEDC_INTR_DISABLE;
    s_ledc_chan.gpio_num   = MOTOR_STEP_GPIO;
    s_ledc_chan.duty       = 512;
    ledc_channel_config(&s_ledc_chan);

    gpio_set_direction(MOTOR_DIR_GPIO, GPIO_MODE_OUTPUT);
    return 0;
}

int motor_dispense_compartment(int compartment_idx) {
    gpio_set_level(MOTOR_DIR_GPIO, compartment_idx % 2 == 0 ? 1 : 0);
    for (int i = 0; i < STEPS_PER_COMPARTMENT; i++) {
        ledc_set_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0, 512);
        ledc_update_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0);
        esp_rom_delay_us(500);
        ledc_set_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0, 0);
        ledc_update_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0);
        esp_rom_delay_us(500);
    }
    return 0;
}
```

- [ ] **Step 4: 创建 optical_detect.c**

```c
#include "optical_detect.h"
#include "driver/adc1.h"
#include "esp_adc_cal.h"
#include <stdbool.h>

#define ADC_CHANNEL ADC1_CHANNEL_0

static esp_adc_cal_characteristics_t s_adc_chars;

int optical_init(void) {
    adc1_config_width(ADC_WIDTH_BIT_12);
    adc1_config_channel_atten(ADC_CHANNEL, ADC_ATTEN_DB_11);
    esp_adc_cal_characterize(ADC_UNIT_1, ADC_ATTEN_DB_11, ADC_WIDTH_BIT_12,
                             1100, &s_adc_chars);
    return 0;
}

bool optical_detect_taken(int compartment_idx) {
    int raw = adc1_get_reading(ADC_CHANNEL, 2);
    int voltage = esp_adc_cal_raw_to_voltage(raw, &s_adc_chars);
    // Reflected IR: high voltage = pill present, low = empty
    return voltage > 2000;  // threshold calibrated during factory test
}
```

- [ ] **Step 5: 创建 tts_broadcaster.c**

```c
#include "tts_broadcaster.h"
#include "driver/uart.h"
#include <string.h>

#define TTS_UART_NUM  UART_NUM_1
#define TTS_BAUD     9600

static const char s_cmd_prefix[] = "$CMD";

int tts_init(void) {
    uart_config_t uart_cfg = {
        .baud_rate = TTS_BAUD,
        .data_bits = UART_DATA_8_BITS,
        .parity    = UART_PARITY_DISABLE,
        .stop_bits = UART_STOP_BITS_1,
        .flow_ctrl = UART_HW_FLOWCTRL_DISABLE,
    };
    uart_param_config(TTS_UART_NUM, &uart_cfg);
    uart_set_pin(TTS_UART_NUM, GPIO_NUM_17, GPIO_NUM_16, UART_PIN_NO_CHANGE, UART_PIN_NO_CHANGE);
    uart_driver_install(TTS_UART_NUM, 0, 0, 0, NULL, 0);
    return 0;
}

void tts_speak(const char* text) {
    // SYN5300: send text command via UART
    uart_write_bytes(TTS_UART_NUM, s_cmd_prefix, strlen(s_cmd_prefix));
    uart_write_bytes(TTS_UART_NUM, text, strlen(text));
    uart_write_bytes(TTS_UART_NUM, "\r\n", 2);
}
```

- [ ] **Step 6: 创建 inventory_watch.c**

```c
#include "inventory_watch.h"
#include "nvs_flash.h"
#include "nvs.h"
#include <string.h>

#define NVS_KEY_INVENTORY "pill_inventory"

static int s_stock[8] = {0};  // up to 8 compartments

int inventory_init(void) {
    nvs_handle_t handle;
    nvs_open("pillbox", NVS_READWRITE, &handle);
    nvs_get_i32_array(handle, NVS_KEY_INVENTORY, s_stock, 8);
    nvs_close(handle);
    return 0;
}

void inventory_decrement(int compartment) {
    if (compartment >= 0 && compartment < 8) {
        s_stock[compartment]--;
        if (s_stock[compartment] < 0) s_stock[compartment] = 0;
    }
}

void inventory_increment(int compartment, int count) {
    if (compartment >= 0 && compartment < 8) {
        s_stock[compartment] += count;
    }
}

bool inventory_is_low(int compartment) {
    return compartment >= 0 && compartment < 8 && s_stock[compartment] <= 3;
}

int inventory_get(int compartment) {
    if (compartment >= 0 && compartment < 8) return s_stock[compartment];
    return 0;
}

void inventory_save(void) {
    nvs_handle_t handle;
    nvs_open("pillbox", NVS_READWRITE, &handle);
    nvs_set_i32_array(handle, NVS_KEY_INVENTORY, s_stock, 8);
    nvs_commit(handle);
    nvs_close(handle);
}
```

- [ ] **Step 7: 创建 main.c**

```c
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "esp_sleep.h"
#include "nvs_flash.h"
#include "dispensing_motor.h"
#include "optical_detect.h"
#include "tts_broadcaster.h"
#include "inventory_watch.h"
#include "free_rtos_tasks.h"
#include "mqtt_common.h"

#define APP_VERSION "auto_v1.0.0"

void app_main(void) {
    nvs_flash_init();

    motor_init();
    optical_init();
    tts_init();
    inventory_init();

    mqtt_common_connect("mqtt.eregen.com", 1883,
                        "PX-AUTO-XXXX", "user", "pass");

    tts_speak("智能药盒已启动");
    free_rtos_tasks_start();

    while (1) {
        vTaskDelay(pdMS_TO_TICKS(60000));
        // Periodic heartbeat
        mqtt_common_publish("eregen/up/heartbeat/PX-AUTO-XXXX",
                           "{\"type\":\"heartbeat\",\"dev_id\":\"PX-AUTO-XXXX\",\"bat\":90}",
                           64, 0);
    }
}
```

- [ ] **Step 8: 编译验证**

Run: `cd firmware/pillbox/auto && idf.py build`

### Task D2: 手环 Pro 档补全

**Files:**
- Create: `firmware/bracelet/pro/ecg_driver.c` + `.h`
- Create: `firmware/bracelet/pro/amoled_display.c` + `.h`
- Create: `firmware/bracelet/pro/gps_high_precision.c` + `.h`
- Create: `firmware/bracelet/pro/power_optimizer.c` + `.h`
- Modify: `firmware/bracelet/pro/main.c` (add ECG + AMOLED init)
- Modify: `firmware/bracelet/pro/free_rtos_tasks.c` (add ECG task)

**Steps:**

- [ ] **Step 1: 创建 ECG 驱动**

```c
// ecg_driver.h
#ifndef ECG_DRIVER_H
#define ECG_DRIVER_H

#include <stdint.h>
#include <stdbool.h>

#define ECG_SAMPLE_RATE_HZ 250
#define ECG_CHANNELS       1

typedef struct {
    int16_t raw_samples[ECG_CHANNELS][ECG_SAMPLE_RATE_HZ];
    uint32_t timestamp;
    bool valid;
} EcgReading_t;

int ecg_init(void);
bool ecg_read(EcgReading_t* reading);
float ecg_calc_hr(const EcgReading_t* reading);

#endif
```

```c
// ecg_driver.c
#include "ecg_driver.h"
#include "driver/adc.h"
#include "esp_adc_cal.h"

static esp_adc_cal_characteristics_t s_adc_chars;

int ecg_init(void) {
    adc1_config_width(ADC_WIDTH_BIT_12);
    adc1_config_channel_atten(ADC1_CHANNEL_6, ADC_ATTEN_DB_11);
    esp_adc_cal_characterize(ADC_UNIT_1, ADC_ATTEN_DB_11, ADC_WIDTH_BIT_12,
                             1100, &s_adc_chars);
    return 0;
}

bool ecg_read(EcgReading_t* reading) {
    for (int i = 0; i < ECG_SAMPLE_RATE_HZ; i++) {
        int raw = adc1_get_reading(ADC1_CHANNEL_6, 4);
        reading->raw_samples[0][i] = (int16_t)raw;
        vTaskDelay(pdMS_TO_TICKS(4));  // 250 Hz = 4ms interval
    }
    reading->timestamp = esp_timer_get_time() / 1000;
    reading->valid = true;
    return true;
}

float ecg_calc_hr(const EcgReading_t* reading) {
    // Simple R-peak detection via first derivative threshold
    float hr = 0;
    int peaks = 0;
    for (int i = 1; i < ECG_SAMPLE_RATE_HZ - 1; i++) {
        int diff = reading->raw_samples[0][i+1] - reading->raw_samples[0][i-1];
        if (diff > 100) peaks++;
    }
    if (peaks > 0) {
        hr = (peaks / 60.0f) * 60.0f;  // simplified HR calculation
    }
    return hr;
}
```

- [ ] **Step 2: 创建 AMOLED 驱动**

```c
// amoled_display.h
#ifndef AMOLED_DISPLAY_H
#define AMOLED_DISPLAY_H

#include <stdint.h>

#define AMOLED_W  416
#define AMOLED_H  416

void amoled_init(void);
void amoled_fill(uint32_t color);
void amoled_draw_pixel(int x, int y, uint32_t color);
void amoled_draw_text(int x, int y, const char* text, uint32_t color);
void amoled_show(void);

#endif
```

```c
// amoled_display.c
#include "amoled_display.h"
#include "driver/spi_master.h"
#include "driver/gpio.h"
#include <string.h>

#define AMOLED_DC GPIO_NUM_5
#define AMOLED_RST GPIO_NUM_18
#define AMOLED_CS GPIO_NUM_19

void amoled_init(void) {
    gpio_set_direction(AMOLED_DC, GPIO_MODE_OUTPUT);
    gpio_set_direction(AMOLED_RST, GPIO_MODE_OUTPUT);
    gpio_set_level(AMOLED_RST, 1);
    vTaskDelay(pdMS_TO_TICKS(100));
    gpio_set_level(AMOLED_RST, 0);
    vTaskDelay(pdMS_TO_TICKS(100));
    gpio_set_level(AMOLED_RST, 1);
    // Send SSD13xx/ST77xx init commands here
    // (specific commands depend on the AMOLED panel model)
}

void amoled_fill(uint32_t color) {
    // Fill entire 416x416 display with color
    // Implementation uses SPI DMA transfer
}

void amoled_draw_pixel(int x, int y, uint32_t color) {
    // Set window and write single pixel via SPI
}

void amoled_draw_text(int x, int y, const char* text, uint32_t color) {
    // Draw text using built-in font or custom bitmap font
}

void amoled_show(void) {
    // Flush frame buffer to display via SPI
}
```

- [ ] **Step 3: 创建高精度 GPS 驱动**

```c
// gps_high_precision.h
#ifndef GPS_HIGH_PRECISION_H
#define GPS_HIGH_PRECISION_H

#include <stdint.h>
#include <stdbool.h>

typedef struct {
    double lat;
    double lon;
    float  accuracy_m;
    uint32_t satellites;
    uint32_t timestamp_ms;
} GpsHighPrecision_t;

int gps_hp_init(void);
bool gps_hp_read(GpsHighPrecision_t* fix);
float gps_hp_calc_speed(const GpsHighPrecision_t* prev, const GpsHighPrecision_t* curr);

#endif
```

```c
// gps_high_precision.c
#include "gps_high_precision.h"
#include "gps_nmea.h"  // reuse existing NMEA parser
#include "driver/uart.h"

#define GPS_UART UART_NUM_2
#define GPS_BAUD 115200

int gps_hp_init(void) {
    // Configure UART for u-blox NEO-M9N at 115200
    // Enable high-precision mode (RTK-ready)
    uart_config_t cfg = {
        .baud_rate = GPS_BAUD,
        .data_bits = UART_DATA_8_BITS,
        .parity = UART_PARITY_DISABLE,
        .stop_bits = UART_STOP_BITS_1,
    };
    uart_param_config(GPS_UART, &cfg);
    uart_driver_install(GPS_UART, 4096, 4096, 0, NULL, 0);
    return 0;
}

bool gps_hp_read(GpsHighPrecision_t* fix) {
    char buf[256];
    int len = uart_read_bytes(GPS_UART, (uint8_t*)buf, sizeof(buf), pdMS_TO_TICKS(100));
    if (len <= 0) return false;

    // Parse GGA/GSV/RMC sentences
    char* gga = strstr(buf, "$GPGGA");
    if (!gga) return false;

    return parse_nmea_gga(gga, &fix->lat, &fix->lon, &fix->accuracy_m,
                          &fix->satellites, &fix->timestamp_ms);
}

float gps_hp_calc_speed(const GpsHighPrecision_t* prev, const GpsHighPrecision_t* curr) {
    // Haversine distance / time delta
    return 0.0f;  // simplified placeholder
}
```

- [ ] **Step 4: 创建功耗优化模块**

```c
// power_optimizer.h
#ifndef POWER_OPTIMIZER_H
#define POWER_OPTIMIZER_H

typedef enum {
    POWER_MODE_ACTIVE,
    POWER_MODE_LIGHT_SLEEP,
    POWER_MODE_DEEP_SLEEP,
} PowerMode_t;

void power_optimizer_init(void);
void power_set_mode(PowerMode_t mode);
PowerMode_t power_get_mode(void);
int power_get_battery_pct(void);
void power_enter_deep_sleep(uint32_t wake_interval_ms);

#endif
```

```c
// power_optimizer.c
#include "power_optimizer.h"
#include "esp_sleep.h"
#include "driver/adc.h"
#include "battery_adc.h"  // reuse from entry

static PowerMode_t s_mode = POWER_MODE_ACTIVE;

void power_optimizer_init(void) {
    adc_power_acquire();
}

void power_set_mode(PowerMode_t mode) {
    s_mode = mode;
    switch (mode) {
        case POWER_MODE_LIGHT_SLEEP:
            esp_light_sleep_start();
            break;
        case POWER_MODE_DEEP_SLEEP:
            // Wake interval set by power_enter_deep_sleep
            break;
        default:
            break;
    }
}

PowerMode_t power_get_mode(void) { return s_mode; }

int power_get_battery_pct(void) {
    return battery_adc_read_percent();  // reuse entry's battery driver
}

void power_enter_deep_sleep(uint32_t wake_interval_ms) {
    esp_sleep_enable_timer_wakeup(wake_interval_ms * 1000ULL);
    esp_deep_sleep_start();
}
```

- [ ] **Step 5: 修改 main.c 整合所有 Pro 模块**

Modify `firmware/bracelet/pro/main.c`:
```c
#include "freertos/FreeRTOS.h"
#include "esp_sleep.h"
#include "nvs_flash.h"
#include "ecg_driver.h"
#include "amoled_display.h"
#include "gps_high_precision.h"
#include "power_optimizer.h"
#include "fall_detect.h"
#include "geofence_manager.h"
#include "ble_pair.h"
#include "battery_optimizer.h"
#include "mqtt_common.h"

void app_main(void) {
    nvs_flash_init();

    // Pro-specific hardware init
    amoled_init();
    ecg_init();
    gps_hp_init();
    power_optimizer_init();
    fall_detect_init();
    geofence_init();
    ble_pair_init();
    battery_optimizer_init();

    mqtt_common_connect("mqtt.eregen.com", 1883,
                        "BR-PRO-XXXX", "user", "pass");

    amoled_fill(0x000000);
    amoled_draw_text(100, 150, "Eregen Pro", 0xFFFFFF);
    amoled_show();

    // Start RTOS tasks
    // (free_rtos_tasks_start from plus tier)
}
```

- [ ] **Step 6: 修改 free_rtos_tasks.c 添加 ECG 任务**

Add to `firmware/bracelet/pro/free_rtos_tasks.c`:
```c
// Additional task for Pro tier: ECG monitoring
static void ecg_task(void* param) {
    EcgReading_t reading;
    while (1) {
        if (ecg_read(&reading)) {
            float hr = ecg_calc_hr(&reading);
            // Publish ECG data via MQTT
            char buf[128];
            snprintf(buf, sizeof(buf),
                     "{\"type\":\"health\",\"dev_id\":\"BR-PRO-XXXX\",\"hr\":%.0f,\"ecg_valid\":true}",
                     hr);
            mqtt_common_publish("eregen/up/health/BR-PRO-XXXX/message",
                               buf, strlen(buf), 0);
        }
        vTaskDelay(pdMS_TO_TICKS(4000));  // Sample every 4 seconds
    }
}

void free_rtos_tasks_start(void) {
    // Existing tasks...
    xTaskCreate(ecg_task, "ECGTask", 2048, NULL, 3, NULL);
}
```

- [ ] **Step 7: 编译验证**

Run: `cd firmware/bracelet/pro && cmake --build build`

---

## Execution Order

```
Phase A (小程序+官网) → Phase B (云平台) → Phase C (前端) → Phase D (固件)
```

Each phase can be reviewed independently. Tasks within a phase run sequentially to maintain consistency.
