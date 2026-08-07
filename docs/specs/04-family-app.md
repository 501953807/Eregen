# ④ 家属APP — 详细设计文档

> 生成日期：2026-07-17
> 更新日期：2026-08-08（v2.0 — 慢性病管理模块升级）
> 对应子系统：④ 家属APP (Flutter 3.24+)
> 语言：Dart | 框架：Flutter | 平台：iOS/Android/Web

---

## 1. 概述

### 1.1 职责

家属APP 是子女查看老人状态、接收告警、管理用药的核心应用。支持多老人绑定、实时定位查看、健康数据趋势分析、SOS 告警即时响应、用药规则远程配置等功能。

**v2.0 新增：慢病管理模块** — 支持血糖/尿酸/血压检测数据管理、饮食运动记录、AI健康建议、任务体系与周期报告，借鉴糖护士APP和第乐健康APP的设计模式。

### 1.2 输入输出

| 类型 | 来源/目标 | 说明 |
|------|-----------|------|
| **输入** | 用户操作 (登录、查看、配置) | Flutter UI 事件 |
| **输入** | WebSocket 实时告警推送 | FCM 触发 + WSS 直连 |
| **输入** | 微信订阅消息通知 | 后台状态更新 |
| **输出** | REST API 调用 | 数据查询/配置下发 → api-server |
| **输出** | 本地缓存 | Hive 存储离线数据 |

---

## 2. 功能模块

### 2.1 核心页面

| 页面 | 路由 | 核心元素 | 原型文件 |
|------|------|---------|---------|
| 首页 | `/` | 高德地图定位、快速状态卡片、老人切换器、SOS 紧急按钮 | `family-app-home.html` |
| 健康看板 | `/health` | 心率/血氧/步数趋势图、异常提示、数据详情 | `family-app-health.html` |
| 告警中心 | `/alerts` | P0/P1/P2 分级告警列表、处理记录、一键呼叫 | `family-app-alert.html` |
| 用药管理 | `/medication` | 今日用药提醒、服药记录、远程配置用药规则 | `family-app-medication.html` |
| 住院治疗 | `/hospitalization` | 老人住院期间的腕带信息、每日诊疗记录、核验历史 | `family-app-hospitalization.html` |
| **登录页** | **`/login`** | **手机号+验证码登录** | **新建** |
| **设备绑定** | **`/bind-device`** | **扫码绑定手环/药盒** | **新建** |
| **慢病管理主页** | **`/chronic`** | **血糖/尿酸趋势卡片 + 每日任务 + AI建议** | **新建** |
| **血糖详情页** | **`/chronic/blood-sugar`** | **趋势图 + 检测记录 + 异常标记 + 导出报告** | **新建** |
| **尿酸详情页** | **`/chronic/uric-acid`** | **趋势图 + 检测记录 + 饮食建议联动** | **新建** |
| **血压详情页** | **`/chronic/blood-pressure`** | **收缩压/舒张压双折线图 + 热力图** | **新建** |
| **饮食记录页** | **`/chronic/diet`** | **食物数据库 + 碳水计算 + AI建议** | **新建** |
| **运动追踪页** | **`/chronic/exercise`** | **手环数据联动 + 运动计划 + 消耗统计** | **新建** |
| **健康报告页** | **`/chronic/report`** | **周报/月报/年报 + AI综合建议** | **新建** |

### 2.2 共享组件

| 组件 | 文件 | 说明 |
|------|------|------|
| 底部导航栏 | `widgets/bottom_nav_bar.dart` | 首页/健康/告警/用药/慢病管理 五个 Tab |
| 老人选择器 | `widgets/elderly_selector.dart` | 下拉切换已绑定的多位老人 |
| 地图组件 | `widgets/map_section.dart` | 高德地图集成，显示老人位置+电子围栏 |
| 状态卡片 | `widgets/quick_status_card.dart` | 电量、在线状态、最新数据摘要 |
| 告警列表 | `widgets/recent_alerts_list.dart` | 最近 10 条告警，支持标记已读 |
| SOS 按钮 | `widgets/sos_button.dart` | 红色大按钮，长按触发手动告警 |
| 主题 | `common/theme.dart` | 适老化设计：大字体、高对比度、暖色调 |

---

## 3. 技术架构

### 3.1 项目结构

```
apps/family-app/
├── lib/
│   ├── main.dart                    # 入口，初始化依赖
│   ├── common/
│   │   └── theme.dart               # 全局主题 (颜色/字体/间距)
│   ├── screens/
│   │   ├── home/home_page.dart            # 首页 - 定位+状态
│   │   ├── health/health_page.dart        # 健康数据看板
│   │   ├── alerts/alerts_page.dart        # 告警中心
│   │   ├── medication/medication_page.dart # 用药管理
│   │   ├── hospitalization/hospitalization_page.dart  # 住院治疗
│   │   ├── chronic/                       # 🆕 慢病管理模块
│   │   │   ├── chronic_home_page.dart     # 慢病管理主页
│   │   │   ├── blood_sugar_page.dart      # 血糖详情页
│   │   │   ├── uric_acid_page.dart        # 尿酸详情页
│   │   │   ├── blood_pressure_page.dart   # 血压详情页
│   │   │   ├── diet_page.dart             # 饮食记录页
│   │   │   ├── exercise_page.dart         # 运动追踪页
│   │   │   └── report_page.dart           # 健康报告页
│   │   ├── login/login_page.dart          # 登录页
│   │   └── bind_device/bind_device_page.dart  # 设备绑定页
│   ├── widgets/                     # 可复用组件
│   │   ├── bottom_nav_bar.dart
│   │   ├── elderly_selector.dart
│   │   ├── map_section.dart
│   │   ├── quick_status_card.dart
│   │   ├── recent_alerts_list.dart
│   │   └── sos_button.dart
│   └── models/                      # 数据模型 (与 api-server model.go 对齐)
├── ios/                             # iOS 原生配置
├── android/                         # Android 原生配置
├── assets/                          # 图片/图标/字体
└── test/                            # 单元测试 + Widget 测试
```

### 3.2 数据流

```
用户操作
    ↓
State Management (Provider/Riverpod)
    ↓
API Service (dio/http)
    ↓
api-server REST API / WebSocket
    ↓
Response → Model → State Update → UI 刷新
```

### 3.3 适老化设计要点

| 设计项 | 标准 |
|--------|------|
| 最小字体 | 16sp (正文)，20sp (标题) |
| 最小触控区域 | 48×48 dp |
| 主色调 | 暖橙色 (#FF6F00) + 白色背景 |
| 按钮高度 | ≥ 56dp |
| 图标尺寸 | ≥ 32dp |
| 对比度 | WCAG AA 级 (≥ 4.5:1) |
| 告警颜色 | 红色 (#D32F2F) 用于 P0，橙色 (#F57C00) 用于 P1 |

---

## 4. 接口定义

### 4.1 认证

```
POST /api/v1/auth/register
Request: {"phone": "13800138000", "password": "...", "otp_code": "123456", "name": "张明"}
Response: {"token": "eyJ...", "refresh_token": "...", "user": {...}}

POST /api/v1/auth/login
Request: {"identifier": "13800138000", "password": "..."}
Response: {"token": "...", "user": {...}}

POST /api/v1/auth/sms/send {phone: string} → 发送验证码
POST /api/v1/auth/sms/login {phone, code: string} → 短信验证码登录
```

### 4.1.1 API Client 封装

```dart
// lib/api/client.dart (新建)
import 'package:dio/dio.dart';

class ApiClient {
  static final ApiClient _instance = ApiClient._internal();
  factory ApiClient() => _instance;
  ApiClient._internal();

  late Dio _dio;
  String? _token;

  Future<void> init({String? token}) async {
    _token = token ?? await _getToken();
    _dio = Dio(BaseOptions(
      baseUrl: 'https://api.eregen.com/api/v1',
      headers: {'Authorization': 'Bearer $_token'},
      connectTimeout: const Duration(seconds: 10),
    ));
    _dio.interceptors.add(InterceptorsWrapper(
      onError: (err) async {
        if (err.response?.statusCode == 401) {
          // Token expired, clear and redirect to login
          await _clearToken();
        }
      },
    ));
  }

  Future<Map<String, dynamic>> get(String path, {Map<String, dynamic>? query}) async { ... }
  Future<Map<String, dynamic>> post(String path, {dynamic data}) async { ... }
}
```

### 4.1.2 登录页面

```dart
// lib/screens/login/login_page.dart (新建)
class LoginPage extends StatefulWidget {
  @override
  _LoginPageState createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _phoneController = TextEditingController();
  final _codeController = TextEditingController();
  bool _sendingCode = false;

  Future<void> _sendVerificationCode() async {
    // POST /auth/sms/send {phone: ...}
  }

  Future<void> _login() async {
    // POST /auth/sms/login {phone: ..., code: ...}
    // Store token, navigate to MainTabScreen
  }
}
```

### 4.1.3 设备绑定页面

```dart
// lib/screens/bind-device/bind_device_page.dart (新建)
class BindDevicePage extends StatelessWidget {
  // 1. QR Code scan → extract device_id from QR
  // 2. POST /devices/bind {device_id: ..., owner_user_id: ...}
  // 3. Show success/failure
}
```

### 4.2 老人档案

```
GET /api/v1/elderly/:id
Response: {"id": "...", "name": "王秀英", "birth_date": "1955-03-15", "health_tiers": ["hypertension", "diabetes"]}

PUT /api/v1/elderly/:id
Request: {"name": "王秀英", "health_tiers": ["hypertension"]}
```

### 4.3 设备管理

```
GET /api/v1/devices?owner_user_id=:uid
Response: [{"id": "...", "device_id": "BR-0001", "device_type": "bracelet", "tier": "plus", "status": "online", "last_seen": "...", "settings": {...}}]

PUT /api/v1/devices/:id/settings
Request: {"settings": {"interval": 30, "volume": 80}}
```

### 4.4 健康数据

```
GET /api/v1/health?elderly_id=:eid&start=2024-07-01&end=2024-07-15
Response: [
  {"id": "...", "elderly_id": "...", "timestamp": "...", "hr": 72, "spo2": 98, "steps": 3456},
  ...
]

GET /api/v1/health/trend?elderly_id=:eid&metric=hr&days=7
Response: {"metric": "hr", "data": [{"ts": "...", "value": 72}, ...], "avg": 75, "max": 98, "min": 62}
```

### 4.5 定位

```
GET /api/v1/location/latest?elderly_id=:eid
Response: {"lat": 31.2304, "lon": 121.4737, "accuracy": 5.0, "timestamp": "..."}

GET /api/v1/location/history?elderly_id=:eid&from=2024-07-01&to=2024-07-15
Response: [{"lat": ..., "lon": ..., "timestamp": ...}, ...]
```

### 4.6 用药管理

```
GET /api/v1/medication/rules?elderly_id=:eid
Response: [{"id": "...", "schedule_time": "08:00", "dose_count": 1, "pill_type": "capsule", "active": true}]

POST /api/v1/medication/rules
Request: {"schedule_time": "12:30", "dose_count": 2, "pill_type": "tablet", "days_of_week": [1,2,3,4,5]}

WS /api/v1/stream/alerts  ← WebSocket 实时推送
```

### 4.7 住院治疗（新增）

```
# 查看老人住院期间的腕带信息
GET /api/v1/medical/hospitalization?elderly_id=:eid&status=active
Response: {
  "patient_id": "...",
  "ward": "心内科3床",
  "admission_date": "2026-07-20",
  "wristband_id": "WB-0001",
  "attending_doctor": "张医生",
  "department": "cardiology"
}

# 查看每日诊疗记录
GET /api/v1/medical/daily-entries?patient_id=:pid&date=2026-07-20
Response: [
  {"time": "08:00", "type": "medication", "content": "降压药 1粒", "verified": true},
  {"time": "10:00", "type": "test", "content": "血常规检查", "verified": false}
]

# 查看近场核验历史
GET /api/v1/medical/verifications?patient_id=:pid&from=&to=
Response: [
  {"timestamp": "...", "nurse": "李护士", "action": "give_medication", "matched": true}
]
```

### 4.8 慢病管理（v2.0 新增）

```
# 血糖
POST   /api/v1/chronic/glucose            录入血糖值
GET    /api/v1/chronic/glucose            获取血糖趋势列表
GET    /api/v1/chronic/glucose/trend      获取趋势聚合数据（用于图表）
POST   /api/v1/chronic/test-strip/read    试纸检测数据上报（从手环）

# 尿酸
POST   /api/v1/chronic/uric-acid          录入尿酸值
GET    /api/v1/chronic/uric-acid          获取尿酸趋势列表

# 血压
POST   /api/v1/chronic/blood-pressure     录入血压值
GET    /api/v1/chronic/blood-pressure     获取血压趋势列表
POST   /api/v1/chronic/bp-device/sync     血压计数据同步（从蓝牙配件）

# 饮食
POST   /api/v1/chronic/diet               记录饮食
GET    /api/v1/chronic/diet               获取饮食记录

# 运动
POST   /api/v1/chronic/exercise           记录运动
GET    /api/v1/chronic/exercise           获取运动记录

# 任务
GET    /api/v1/chronic/daily-tasks        获取当日任务列表
PUT    /api/v1/chronic/daily-tasks/:id    标记任务完成

# 报告
GET    /api/v1/chronic/report/:type       获取周期报告（weekly/monthly/annual）
POST   /api/v1/chronic/report/generate    手动生成报告

# AI建议
GET    /api/v1/chronic/recommendations    获取综合AI建议
POST   /api/v1/chronic/recommendations/feedback  反馈建议效果
```

---

## 5. 编译与运行

```bash
cd apps/family-app

# 获取依赖
flutter pub get

# Web 调试 (浏览器)
flutter run -d chrome

# Android 真机
flutter devices          # 查看可用设备
flutter run -d <device_id>

# iOS 真机 (需 macOS + Xcode)
flutter run -d <device_id>
```

### 5.1 构建发布包

```bash
# Android APK
flutter build apk --release

# Android App Bundle (Google Play)
flutter build appbundle --release

# iOS (macOS only)
flutter build ios --release
```

---

## 6. 测试策略

| 测试类型 | 工具 | 覆盖范围 |
|---------|------|---------|
| 单元测试 | `flutter test` | Model、Service、工具函数 |
| Widget 测试 | `flutter test --widget` | 每个页面的 UI 渲染 |
| 集成测试 | `integration_test/` | 端到端：登录→查看定位→处理告警 |

---

© 2026 Eregen (颐贞). All rights reserved.
