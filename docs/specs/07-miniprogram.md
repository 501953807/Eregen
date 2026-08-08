# ⑦ 微信小程序 — 详细设计文档

> 生成日期：2026-07-17  
> 对应子系统：⑦ 微信小程序 (原生 WXML/WXSS)  
> 框架：微信原生 | 基础库：2.44+

---

## 1. 概述

### 1.1 职责

微信小程序是轻量版家属端，聚焦核心监护功能：老人位置查看、用药提醒接收、紧急告警通知。相比 Flutter 家属 APP，小程序功能精简但触达更广——无需下载安装，通过微信即可使用。

### 1.2 与家属 APP 的差异

| 功能 | 家属 APP (Flutter) | 微信小程序 |
|------|-------------------|-----------|
| 实时定位 | ✅ 高德地图 | ✅ 腾讯地图 |
| 健康数据趋势 | ✅ 完整图表 | ⚠️ 仅最新值 |
| SOS 告警 | ✅ 实时推送 | ✅ 订阅消息 |
| 用药管理 | ✅ 远程配置 | ⚠️ 仅接收提醒 |
| 多老人绑定 | ✅ 完整管理 | ⚠️ 最多 3 位 |
| 电子围栏 | ✅ 多边形围栏 | ❌ 仅圆形 |
| 订阅管理 | ✅ 完整功能 | ❌ 跳转 H5 |
| 离线缓存 | ✅ Hive 本地存储 | ⚠️ wxStorage (5MB) |

---

## 2. 功能模块

### 2.1 核心页面

| 页面 | 路径 | 核心元素 | 原型文件 |
|------|------|---------|---------|
| 首页 | `pages/home/home` | 老人位置(腾讯地图)、快速状态卡片、SOS 按钮 | `miniprogram-home.html` |
| 用药提醒 | `pages/medication/medication` | 今日用药列表、服药确认按钮、提醒时间 | `miniprogram-medication.html` |
| 告警中心 | `pages/alerts/alerts` | P0/P1 告警列表、处理记录 | `miniprogram-alerts.html` |
| 住院治疗 | `pages/hospitalization/hospitalization` | 住院期间腕带信息、每日诊疗记录、核验历史 | `miniprogram-hospitalization.html` |
| 我的 | `pages/mine/mine` | 老人管理、设置、帮助 | `miniprogram-mine.html` |

### 2.2 项目结构

```
apps/miniprogram/
├── pages/
│   ├── home/
│   │   ├── home.wxml          # 首页模板
│   │   ├── home.wxss          # 首页样式
│   │   ├── home.js            # 首页逻辑
│   │   └── home.json          # 页面配置
│   ├── medication/
│   │   ├── medication.wxml
│   │   ├── medication.wxss
│   │   ├── medication.js
│   │   └── medication.json
│   ├── alerts/
│   │   ├── alerts.wxml
│   │   ├── alerts.wxss
│   │   ├── alerts.js
│   │   └── alerts.json
│   ├── hospitalization/
│   │   ├── hospitalization.wxml
│   │   ├── hospitalization.wxss
│   │   ├── hospitalization.js
│   │   └── hospitalization.json
│   └── mine/
│       ├── mine.wxml
│       ├── mine.wxss
│       ├── mine.js
│       └── mine.json
├── components/                  # 自定义组件
│   ├── status-card/            # 状态卡片
│   ├── alert-item/             # 告警条目
│   └── med-list/               # 用药列表
├── utils/
│   ├── api.js                  # API 请求封装
│   ├── auth.js                 # 微信登录 + JWT 获取
│   └── storage.js              # 本地存储封装
├── app.js                       # 小程序入口
├── app.json                     # 全局配置
├── app.wxss                     # 全局样式
└── sitemap.json                 # 索引配置
```

---

## 3. 技术实现

### 3.1 微信登录流程

```
用户打开小程序
    ↓
wx.login() → 获取 code
    ↓
POST /api/v1/auth/wechat/login {code}
    ↓
api-server 返回 JWT Token
    ↓
wx.setStorageSync('token', jwt_token)
    ↓
后续请求自动携带 Authorization: Bearer <token>
```

### 3.2 订阅消息推送

```javascript
// 请求订阅消息授权
wx.requestSubscribeMessage({
  tmplIds: [
    'alert_notification_id',   // 告警通知
    'medication_reminder_id',   // 用药提醒
    'location_update_id',       // 位置更新
  ],
  success(res) {
    console.log('订阅结果:', res)
  }
})
```

云端推送流程：
```
pipeline AI 分析发现告警
    ↓
NATS → push-service
    ↓
push-service 判断用户有微信订阅授权
    ↓
调用微信 subscribe_message.send API
    ↓
用户收到微信消息通知
    ↓
点击通知 → 打开小程序指定页面
```

### 3.3 腾讯地图集成

```javascript
// pages/home/home.js
onLoad() {
  // 加载腾讯地图插件
  const mapPlugin = requirePlugin('TencentMapPlugin')
  this.setData({
    latitude: 31.2304,
    longitude: 121.4737,
    markers: [{
      icon: '/assets/location-bracelet.png',
      callout: { content: '王秀英', fontSize: 14 }
    }]
  })
  this.fetchLatestLocation()
},

async fetchLatestLocation() {
  const token = wx.getStorageSync('token')
  const res = await api.get('/location/latest', {
    elderly_id: this.data.currentElderlyId,
    headers: { Authorization: `Bearer ${token}` }
  })
  this.setData({
    latitude: res.lat,
    longitude: res.lon,
    accuracy: res.accuracy
  })
}
```

### 3.4 新增页面

- **`pages/login/index.{js,wxml,wxss,json}`** — 手机号+验证码登录
  - `wx.login()` → 获取 code → `POST /api/v1/auth/wechat/login {code}` → 存储 token
- **`pages/add-elderly/index.{js,wxml,wxss,json}`** — 添加老人档案
  - 表单：姓名、身份证号、与本人关系
  - `POST /api/v1/elderly` 创建后自动跳转首页

---

## 4. 接口定义

小程序复用 api-server 的 REST API，以下为核心接口：

```
POST /api/v1/auth/wechat/login          # 微信登录
GET  /api/v1/elderly/:id                # 老人档案
GET  /api/v1/devices                    # 设备列表 (最多 3 位)
GET  /api/v1/location/latest            # 最新位置
GET  /api/v1/health?elderly_id=         # 最新健康数据
GET  /api/v1/medication/rules           # 用药规则
GET  /api/v1/alerts?status=pending      # 待处理告警

# 住院治疗（新增）
GET  /api/v1/medical/hospitalization?elderly_id=  # 当前住院信息
GET  /api/v1/medical/daily-entries?patient_id=    # 每日诊疗记录
GET  /api/v1/medical/verifications?patient_id=    # 核验历史
```

---

## 5. 编译与发布

### 5.1 开发者工具

1. 下载并安装 [微信开发者工具](https://developers.weixin.qq.com/miniprogram/dev/devtools/download.html)
2. 打开 `apps/miniprogram` 目录
3. 在 `app.json` 中配置 `appid` (申请的小程序 ID)
4. 点击编译即可预览

### 5.2 上传代码

```bash
# 使用微信开发者工具 CLI
npx miniprogram-ci build \
  --project ./apps/miniprogram \
  --appid YOUR_APP_ID \
  --secret YOUR_APP_SECRET \
  --output ./dist
```

### 5.3 版本管理

| 步骤 | 操作 |
|------|------|
| 1. 开发 | 开发者工具本地调试 |
| 2. 上传 | 开发者工具 → 上传 → 填写版本号和备注 |
| 3. 审核 | 微信公众平台 → 版本管理 → 提交审核 |
| 4. 发布 | 审核通过后点击"发布" |
| 5. 灰度 | 可选灰度发布 (先对 10% 用户开放) |

---

## 6. 注意事项

| 限制项 | 上限 | 应对策略 |
|--------|------|---------|
| 包体积 | 主包 2MB | 图片压缩、分包加载 |
| 总大小 | 20MB | 使用分包 (每个 ≤ 2MB) |
| 网络请求 | HTTPS 域名白名单 | 配置服务器域名 |
| 本地存储 | 10MB | 仅缓存 token + 必要数据 |
| 定位精度 | 约 10 米 | 足够电子围栏使用 |

---

© 2026 Eregen (颐贞). All rights reserved.
