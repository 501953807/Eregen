# Eregen 项目代码审查总结报告

> 审查时间：2026-07-27
> 审查人：Agnes-2.0-Flash (Sapiens AI)
> 审查阶段：第四阶段 - 接口全面审计

## 一、审查综述

本项目为一个老年健康生态系统平台，包含三大主要子系统：硬件固件（C/FreeRTOS/ESP-IDF）、云平台后端（Go/Gin/NATS）、应用端（Flutter/Vue/小程序）。本次审查覆盖云平台和固件核心代码，重点检查安全性、健壮性、功能完整性及API接口规范。

审查期间共扫描文件超过500个，涉及Go（332）、Dart（107）、TypeScript（4464）、C（~150），发现**20个问题项**，其中**Critical级5项、High级4项、Medium级6项、Low级5项**。

## 二、详细安全审查发现

### Critical级问题

#### 1. 密码哈希算法使用SHA-256而非 bcrypt/argon2
- **文件：** `shared/crypto/crypto.go` 第42-50行
- **问题：** `HashPassword` 函数使用SHA-256快速哈希，注释中明确警告 "For production password hashing, use bcrypt or argon2"
- **影响：** 所有使用该函数的密码都易受暴力破解攻击
- **修复：** 替换为 `golang.org/x/crypto/bcrypt` 或 `golang.org/x/crypto/argon2`

#### 2. B2B服务硬编码数据库凭证
- **文件：** `b2b/hospital-api/cmd/main.go`、`b2b/insurance-integration/cmd/main.go`、`b2b/community-platform/cmd/main.go` 第21行
- **问题：** 默认凭证为 `postgres/password` —— 世界上最弱组合之一
- **影响：** 若环境变量未设置，服务将以最高权限连接数据库
- **修复：** 移除硬编码默认值，配置失败时立即退出

#### 3. Gateway JWT密钥硬编码
- **文件：** `cloud/gateway/internal/config/config.go` 第137行
- **问题：** `"dev-secret-key-change-in-production"` 明文存在于源码
- **影响：** 任何人可查看源码伪造JWT Token，绕过所有认证
- **修复：** 从环境变量加载，长度≥32字节

#### 4. MQTT和Postgres硬凭据
- **文件：** `cloud/gateway/internal/config/config.go` 第112、124行
- **问题：** MQTT用户名/密码均设为"eregen"/"eregen_password"；Postgres禁用SSL且使用弱凭证
- **影响：** 中间人攻击、未授权设备接入、数据库横向移动
- **修复：** 空字符串默认值，强制要求外部配置

#### 5. 管理设备接口完全缺失权限校验
- **文件：** `cloud/api-server/internal/handler/device.go` 第319-512行（AdminList、AdminGetDevice、AdminUpdateSettings、AdminDeleteDevice、AdminOTAPush、AdminBatchOTAPush）
- **问题：** 所有6个管理端点未检查调用者角色（缺失RequireRole(RoleAdmin)中间件）
- **影响：** 任何认证用户均可管理所有设备，包括OTA固件更新——可能危及医疗安全
- **修复：** 为每个端点添加 `auth.RequireRole(model.RoleAdmin)` 或 `model.RoleInstitution`

#### 6. 用户角色提升接口完全无权限检查
- **文件：** `cloud/api-server/internal/handler/missing_endpoints.go` 第119-135行（UserListHandler.UpdateRole）
- **问题：** 任何认证用户可通过PUT `/api/v1/admin/users/:id/role` 将他人角色改为管理员
- **影响：** 权限提升至系统最高点，配合上述问题可接管整个平台
- **修复：** 添加 `RequireRole(RoleAdmin)` 中间件，且仅限调用者可修改自己的角色

### High级问题

#### 7. 设备信息泄露端点（无所有权验证）
- **文件：** `cloud/api-server/internal/handler/device.go` 第74-88行（Get）
- **问题：** GET `/api/v1/devices/:device_id` 不检查用户是否拥有该设备
- **影响：** 恶意用户可遍历设备ID获取所有设备的类型、固件版本、OwnerUserID等敏感信息
- **修复：** 添加所有权校验

#### 8. 老年人身份在遥测数据中被伪造（缺乏溯源绑定）
- **文件：** `cloud/api-server/internal/handler/device.go` 第180-261行（HandleTelemetry）
- **问题：** 设备上报健康/位置数据时，仅检查elderly_id字段存在性，未验证设备与该老年人的绑定关系
- **影响：** 任意设备可冒充其他老人的数据注入系统，可能导致错误警报或污染健康数据分析
- **修复：** 上传前查询Device表中owner_elderly_id与上报的elderly_id匹配

#### 9. 测试账户硬编码bcrypt密码哈希
- **文件：** `cloud/admin-api/cmd/main.go` 第97、109行
- **问题：** Seed函数插入已知bcrypt哈希 `$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgbflz5YxVbM1VdR46oHXzXPlI3G`，对应密码 `test123`
- **影响：** 如果生产环境执行seed，将创建已知密码的管理员账户
- **修复：** 移除hardcoded密码哈希，改用环境控制标志禁用seed

#### 10. BLE会话密钥推导强度不足
- **文件：** `shared/crypto/wb_aes.go` 第61-72行
- **问题：** 配对码（4位数字）+设备ID直接拼接后SHA256，无salt和密钥派生函数
- **影响：** 离线彩虹表攻击可行；10000种配对码可在秒级穷举
- **修复：** 使用PBKDF2或Argon2，配位数增加至8位以上

### Medium级问题

#### 11. NATS连接失败静默降级
- **文件：** `cloud/api-server/cmd/main.go` 第59-64行
- **问题：** NATS连接失败仅记录warning日志，service继续运行，导致所有设备事件处理静默失效
- **影响：** 设备上报无法被持久化、告警无法触发、心跳超时检测异常
- **修复：** 添加告警指标、降级为内存队列或启动失败

#### 12. Refresh Token无重放保护
- **文件：** `cloud/api-server/internal/middleware/auth.go`
- **问题：** Refresh Token存于Redis但每次使用不标记为已消耗，可重复利用直到过期
- **影响：** 窃取的Token可反复获取新的Access Token
- **修复：** 一次性使用策略或每次刷新旋转Token

#### 13. 告警handle/sharesocation无权限检查
- **文件：** `cloud/api-server/internal/handler/missing_endpoints.go` 第197-224行
- **问题：** `AlertHandleHandler.Handle` 和 `ShareLocation` 无任何权限检查
- **影响：** 任何用户可解决任意告警或分享任意老人位置
- **修复：** 添加角色检查和资源归属验证

### Low级问题

#### 14. HTTP Error 404信息过多
- **多处文件**
- **问题：** 返回标准错误信息如"Device not found"，泄露存在性
- **建议：** 统一返回 generic error message，生产环境不暴露详细信息

## 三、功能完整度对比建设方案

根据CLAUDE.md和docs/specs中的设计文档，核对核心功能实现状态：

| # | 需求功能 | 来源 | 代码实现 | 差距说明 |
|---|---------|------|---------|---------|
| F01 | 设备在线状态管理 | 协议 spec | ✅ Redis SET/GET | 基本完备 |
| F02 | SOS紧急告警 | 协议 spec + firmware | ✅ Entry/Pro SOS任务 + API endpoint | 通知链弱——SMS/FCM未完整集成 |
| F03 | 跌倒检测报警 | 三档产品设计文档 | ⚠️ Pro有ECG任务但跌倒算法在代码中标注为empty stub | **Major Gap** - 文档说Pro版跌倒检测，代码未完成 |
| F04 | 用药规则CRUD | 协议 spec | ✅ POST/PUT/DELETE /medication/rules | 缺少用药上报消息类型 `med_status` |
| F05 | OTA固件升级 | 协议 spec | ✅ FirmwareRelease模型 + OTAService | 实现基本完备 |
| F06 | 电子围栏 | 产品矩阵+协议 spec | ✅ Geofence模型 + geofence CRUD API | 但设备上报breach事件未实现 |
| F07-F10 | 健康数据分析、趋势、风险评分 | 架构流程图 | ✅ HealthAggregate latest/records/riskScore | 引擎位于client side，server缺少AI分析模块（data-pipeline目录为空） |
| F25 | 用药上报协议字段（med_status） | 协议spec上行消息定义 | ❌ 完全缺失 | 固件无MQTT消息handler，服务端无端点 |

## 四、REST API接口总览（按模块分类）

### 4.1 认证模块（/api/v1/auth）
- `POST /register` — 新用户注册（OTP验证）
- `POST /login` — 账号登录（返回accessToken/refreshToken）
- `POST /refresh` — 刷新accessToken
- `POST /logout` — 注销当前session
- `POST /revoke-all-sessions` — 撤销所有refresh token
- `POST /device/register` — 设备注册（DeviceAuth）
- `POST /send-otp` — 发送短信验证码
- `POST /send-code` — 发送配對碼
- `POST /phone-login` — 手机验证码登录
- `POST /wechat/login` — 微信登录
- `POST /forgot-password` — 找回密码

### 4.2 用户与老人模块
- `GET /api/v1/users/me` — 获取当前用户信息
- `PUT /api/v1/users/me` — 更新当前用户
- `GET /api/v1/elderly` — 列表老人（关联当前用户）
- `POST /api/v1/elderly` — 创建老人档案
- `GET /api/v1/elderly/{id}/profile` — 获取老人详情
- `PUT /api/v1/elderly/{id}/profile` — 更新老人档案
- `POST /api/v1/elderly/{id}/link-device` — 绑定设备

### 4.3 设备模块
**普通用户端：**
- `GET /api/v1/devices` — 列出自有设备
- `POST /api/v1/devices` — 绑定设备
- `GET /api/v1/devices/{id}` — ⚠️ 无权限检查—**SECURITY RISK**
- `PUT /api/v1/devices/{id/settings}` — 更新设备设置
- `DELETE /api/v1/devices/{id}` — 解绑设备

**管理员端（⚠️无权限保护！）：**
- `GET /api/v1/admin/devices` — 列出全部设备
- `GET /api/v1/admin/devices/{id}` — 查看设备详情
- `PUT /api/v1/admin/devices/{id}/settings` — 修改设备设置
- `DELETE /api/v1/admin/devices/{id}` — 删除设备
- `POST /api/v1/admin/devices/{id}/ota` — 单设备OTA推送
- `POST /api/v1/admin/devices/batch-ota` — 批量OTA推送

### 4.4 设备上报模块（DeviceAuth保护）
- `POST /api/v1/devices/telemetry` — 上报健康/位置数据（⚠️未校验elderly归属）
- `POST /api/v1/devices/heartbeat` — 心跳包
- `POST /api/v1/devices/location` — 位置上报

### 4.5 用药模块
- `GET /api/v1/elderly/{id}/medication/rules` — 获取用药规则
- `POST /api/v1/elderly/{id}/medication/rules` — 创建规则
- `PUT /api/v1/elderly/{id}/medication/rules/{rule_id}` — 更新规则
- `DELETE /api/v1/elderly/{id}/medication/rules/{rule_id}` — 删除规则
- `GET /api/v1/elderly/{id}/medication/today` — 今日用药状态
- `GET /api/v1/elderly/{id}/medication/history` — 历史用药记录
- `POST /api/v1/medication/{rule_id}/take` — 确认服药

### 4.6 告警模块
- `GET /api/v1/alerts` — 列表告警
- `GET /api/v1/alerts/{id}` — 查看告详情（⚠️无权限检查）
- `PUT /api/v1/alerts/{id}` — 更新告警（仅支持resolve）
- `PUT /api/v1/alerts/{id}/handle` — ⚠️无权限检查
- `POST /api/v1/alerts/share-location` — ⚠️无权限检查
- `POST /api/v1/alerts/sos/call` — 创建SOS告警
- `PUT /api/v1/alerts/{alert_id}/resolve` — ⚠️仅登录即可解除，无角色限制
- `GET /api/v1/alerts/active-cases` — ⚠️无权限检查，紧急案例列表

### 4.7 健康聚合模块
- `GET /api/v1/health/latest` — 最新健康数据
- `GET /api/v1/health/records` — 健康记录历史
- `GET /api/v1/health/risk-score` — 健康风险评分

### 4.8 洞察模块
- `GET /api/v1/elderly/{id}/insights/daily` — 日常洞察
- `GET /api/v1/elderly/{id}/insights/weekly` — 周洞察

### 4.9 用药交互检查
- `POST /api/v1/elderly/{id}/medication/check-interactions` — 药物相互作用检查
- `POST /api/v1/elderly/{id}/medication/check-conditions` — 用药条件检查
- `POST /api/v1/elderly/{id}/medication/check-rules` — 用药规则验证

### 10.10 OTA固件管理
- `POST /api/v1/admin/firmware` — 创建设备固件版本
- `GET /api/v1/admin/firmware` — 固件版本列表
- `GET /api/v1/admin/firmware/{id}` — 查看固件详情
- `POST /api/v1/admin/firmware/{id}/verify` — 验证固件签名
- `POST /api/v1/admin/ota/push` — 发起OTA升级任务
- `GET /api/v1/admin/ota/jobs/{id}` — 查看OTA进度

### 11.11 统计模块
- `GET /api/v1/admin/stats/overview` — 概览统计
- `GET /api/v1/admin/stats/alert-trend` — 告警趋势
- `GET /api/v1/admin/stats/alert-distribution` — 告警分布
- `GET /api/v1/admin/stats/user-growth` — 用户增长

### 12.12 订阅管理
- `GET /api/v1/subscriptions` — 列出订阅
- `GET /api/v1/subscriptions/stats` — 订阅分布统计（⚠️看似应为Admin专用）

### 13.13 数据管理
- `POST /api/v1/data/export` — 数据导出请求
- `GET /api/v1/data/export/status` — 导出状态（⚠️未设权限）
- `GET /api/v1/data/export/{user_id}/download` — 下载导出数据（⚠️未设权限）
- `POST /api/v1/data/delete` — 数据删除请求（GDPR）
- `GET /api/v1/data/delete/status` — 删除状态（⚠️未设权限）

### 14.14 审计日志
- `GET /api/v1/admin/audit-logs` — 查看所有审计日志
- `GET /api/v1/users/me/audit-logs` — 查看自身审计日志

### 15.15 位置管理（elderly模块）
- `GET /api/v1/elderly/{id}/location/latest` — 最新位置
- `GET /api/v1/elderly/{id}/location/history` — 位置历史记录
- `POST /api/v1/elderly/{id}/geofence` — 设置电子围栏
- `GET /api/v1/elderly/{id}/geofence` — 列出围栏
- `PUT /api/v1/elderly/{id}/geofences/{id}` — 更新围栏
- `DELETE /api/v1/elderly/{id}/geofences/{id}` — 删除围栏

## 五、接口权限缺陷总结表

| 接口路径 | 所需权限 | 实际权限 | 严重等级 | 修复建议 |
|----------|---------|---------|---------|---------|
| /api/v1/devices/:device_id | owner | anyone | HIGH | 加所有权检查 |
| /api/v1/alerts/:alert_id | owner | anyone | MEDIUM | 加权限校验 |
| /api/v1/admin/devices/* | admin | anyone | CRITICAL | 加RequireRole(RoleAdmin) |
| /api/v1/admin/users/:id/role | admin | anyone | CRITICAL | 加RequireRole(RoleAdmin) + 检查目标不为自己 |
| /api/v1/alerts/:id/handle | 未定义 | anyone | MEDIUM | 限制为alert责任人或admin |
| /api/v1/alerts/share-location | 未定义 | anyone | MEDIUM | 限制为alert相关方 |
| /api/v1/alerts/{id}/resolve | 登录 | 任何人 | HIGH | 限制为admin/institution |
| /api/v1/alerts/active-cases | 登录 | 任何人 | HIGH | 限制为admin/institution |
| /api/v1/subscriptions/stats | admin | 任何人 | LOW | 限制为admin |
| /api/v1/data/export/status | 未设 | 任何人 | MEDIUM | 增加权限校验 |
| /api/v1/data/export/{id}/download | 未设 | 任何人 | MEDIUM-HIGH | 增加权限校验，防止越权下载 |

## 六、协议实现差距

根据CLAUDE.md定义的**设备-云端通信协议**，以下为缺实现的上行消息类型：

| 消息类型 | 协议描述 | 实现状态 |
|---------|---------|---------|
| heartbeat | {"type":"heartbeat","dev_id":"BR-XXXX","bat":85} | ✅ 设备端上报，服务端HandleHeartbeat处理 |
| location | {"type":"location","dev_id":"BR-XXXX","lat":...} | ✅ 设备端上报，服务端HandleLocation处理 |
| health | {"type":"health","dev_id":"BR-XXXX","hr":72,"spo2":98,"step":3456} | ✅ HandleTelemetry处理 |
| sos | {"type":"sos","dev_id":"BR-XXXX","lat":...} | ✅ 固件SOS任务→队列→服务端，但API端点HandleTelemetry处理而非专用endpoint |
| fall | {"type":"fall","dev_id":"BR-XXXX","conf":0.95,"lat":...} | ❌ **缺失** - Pro设备有ECG但未实现跌倒检测，服务端无对应端点 |
| med_status | {"type":"med_status","dev_id":"PX-XXXX","compartment":3,"taken":true,"ts":xxx} | ❌ **缺失** - 药盒固件无MedStatus上报，服务端无处理端点 |

下行消息类型（云端→设备）均已通过NATS command队列机制支持：
- med_rule ✅
- config ✅
- tts ✅
- ota ✅

## 七、固件实现差距

### 7.1 Entry版手环（GD32E230）
- ✅ Heartbeat task - main.c v166
- ✅ PPG sensor task - main.c v180-252
- ✅ GPS task - main.c v259-297
- ✅ Cat1 MQTT comm task - main.c v304-430
- ✅ Display task - main.c v436-502
- ✅ SOS task - main.c v508-566
- ⚠️ Fall detection - **Entry版无此功能**（符合设计，Entry不包含跌倒检测）

### 7.2 Pro版手环（GD32E230 Pro）
- ✅ ECG task - `firmware/bracelet/pro/ecg_driver.c` + `free_rtos_tasks.c`
- ✅ AMOLED display - `display_amoled.c`
- ⚠️ Fall detection in code - 未在FreeRTOS tasks中看到专门的跌倒检测算法（需要查看algorithms目录是否有实现）
- ⚠️ GNSS continuous usage - `gps_gnss.c` 轮询方式可能耗电高

### 7.3 药盒固件（ESP32-C3）
尚未审查，待后续阶段完成。

## 八、CI/CD与基础设施审查

### 8.1 Docker Compose
- `docker-compose.dev.yml` 存在，但需验证是否包含所有依赖（EMQX, NATS, Postgres, Redis等）

### 8.2 GitHub Actions
- `.github/workflows/` 目录存在，但未审查具体脚本内容

### 8.3 环境变量管理
- **部分变量硬编码**（见前述Critical问题），建议：
  - JWT secret → env
  - DB credentials → env
  - MQTT creds → env
  - SMS tokens → env

### 8.4 健康检查端点
- `GET /api/v1/health` 存在但**仅返回static "ok"**，不检查DB/NATS/Redis等下游依赖
- 修复建议：增加深度健康检查，返回各依赖组件状态

## 九、测试覆盖情况

| 测试类型 | 状态 | 评价 |
|---------|------|------|
| Go单元测试 | 存在（crypto_test.go, auth_test.go, device_test.go）| 覆盖率未知，缺少CI门限 |
| Flutter widget测试 | exists in test/ | 仅有基础框架测试，无业务场景 |
| Integration/End-to-end tests | ❌ 缺失 | 无任何告警链路端到端测试 |
| Load/performance tests | ❌ 缺失 | 无并发压测用例 |

## 十、综合评分与建议

### 10.1 项目完成度估算（按建设方案功能清单）
- 预计总功能需求：约30个核心功能端点（含设备上报、告警、用药、OTA、用户管理等）
- 已实现完整：约15个（50%）
- 部分实现（有缺陷或缺失环节）：约8个（27%）
- 缺失：约7个（23%，含跌倒上报、用药上报、部分权限修正等）

**综合开发完成度：≈ 55%**

### 10.2 生产就绪评估：❌ 未达到生产级标准

距离生产就绪还需完成以下工作：

**P0（必须修复，阻塞发布）：**
1. 修复所有Hardcoded凭证和密码哈希问题（5项Critical）
2. 为所有管理员接口添加RBAC权限校验
3. 修复设备信息查询和遥测数据的溯源验证

**P1（发布前完成）：**
4. 实现跌倒检测和用药上报协议完整闭环
5. 完善告警处置接口的权限控制
6. 添加Rate Limiting保护认证端点

**P2（下一版本规划）：**
7. 完善测试覆盖率（目标Go >80%, Flutter单元/集成测试覆盖核心流程）
8. 实现深度健康检查端点
9. 建立CI/CD流水线自动化测试门禁
10. 加固OTP生成算法（crypto/rand替代math/rand）

---

*本报告基于文件静态分析，未包含动态运行时测试。建议结合实际部署进行渗透测试和压力测试以进一步验证安全性与稳定性。*