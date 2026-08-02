# Eregen (颐贞) — 老年健康品牌平台

## 项目愿景

打造一个完整的老年健康生态系统，包含智能手环、智能药盒等硬件产品，以及云端中控系统、家属APP、运营管理后台、品牌官网、微信小程序和医院/社区对接系统。所有软硬件自主开发，申请专利保护。

**品牌：** 颐贞 (yí zhēn) / Eregen (Ease + Regen)
**Slogan：** "颐养正道，贞守安康"
**目标人群：** 60-85岁老人(使用者) + 40-55岁子女(购买者)
**项目性质：** 技术预研可行性开发 — 研究功能可行性和系统落地，不是市场性产品
**核心原则：** 设计全面细致不遗漏，UI界面先出高保真效果图再编码

---

## 绝对不可变更的技术选型

以下选型已经过充分评估，**整个项目周期内不再更改**。每次新会话启动时必须遵守这些选型。

### 硬件固件层

| 子系统 | 主控芯片 | RTOS/SDK | 语言 |
|--------|---------|----------|------|
| 手环(初/中/高端) | GD32E230C8T3 (ARM Cortex-M4) | FreeRTOS | C |
| 药盒(基础/智能/自动) | ESP32-C3 (RISC-V) | ESP-IDF v5.3 | C |

**原因：** FreeRTOS是MCU事实标准，90%厂商提供官方port。ESP32原生ESP-IDF生态最成熟。不用Zephyr/RT-Thread增加集成风险。

### 云平台后端

| 组件 | 选型 | 版本 |
|------|------|------|
| 语言/框架 | Go + Gin | Go 1.22+, Gin 1.9+ |
| MQTT Broker | EMQX | 5.x (开源版) |
| 消息总线 | NATS | 2.10+ |
| 用户数据库 | SQLite | — | MVP阶段全项目统一，零部署单文件存储 |
| 推送 | FCM + 阿里云SMS + 微信订阅消息 | — |

**为什么不用Java/Spring Boot：** Spring Boot太重，IoT网关需要高并发低延迟，Go goroutine天然适合。
**SQLite 迁移说明：** MVP 阶段采用 SQLite 替代 PostgreSQL/InfluxDB/Redis，实现单文件零部署、免运维的轻量方案。所有数据库操作封装在 `store/` 层，未来切换至 PostgreSQL 时只需替换存储实现代码，业务逻辑无需改动。

### 应用端

| 子系统 | 选型 | 版本 |
|--------|------|------|
| 家属APP | Flutter (Dart) | 3.24+ |
| 管理后台 | Vue 3 + TypeScript + Element Plus | Vue 3.4+, TS 5.4+, Element Plus 2.7+ |
| 微信小程序 | 原生WXML/WXSS | 基础库 2.44+ |
| 品牌官网 | Hugo + Tailwind CSS | Hugo 0.128+, Tailwind 3.4+ |

**为什么Flutter不用React Native：** RN Bridge长时运行有性能衰减，Flutter AOT编译性能可预测。
**为什么Vue不用React：** 小团队Vue开发效率更高，管理后台不需要React的虚拟DOM优势。

### 开源许可策略（专利申请导向）

- **允许：** MIT / BSD-3 / Apache-2.0 / ISC
- **禁止：** GPL / AGPL / LGPL (copyleft会强制你的代码也开源，影响专利申请)
- **核心业务逻辑、通信协议、AI算法全部自研闭源**
- **每个子项目必须维护 `THIRD-PARTY-LICENSES` 文件**

### CI/CD与基础设施

| 组件 | 选型 |
|------|------|
| CI/CD | GitHub Actions |
| 容器化 | Docker + Docker Compose |
| 日志 | Loki + Grafana |
| 监控 | Prometheus + Grafana |
| 域名/SSL | 阿里云DNS + Let's Encrypt |

---

## 三档产品线矩阵

| | 入门版 (Starter) | 中端版 (Plus) | 高端版 (Pro) |
|---|---|---|---|
| **手环** | 心率+血氧+SOS+GPS定位 | +电子围栏+跌倒检测+长续航 | +ECG心电+AMOLED屏+金属机身+高精度GPS |
| **药盒** | 基础分格药盒(无电子功能) | 定时语音提醒+APP联动 | 自动分药+光电检测+库存预警+TTS播报 |
| **医用腕带** | 身份识别+NFC核验+SOS+Cat1上报 | +跌倒检测+用药扫码 | +电子围栏+ECG+家属端查看 |
| **社区腕带** | 福利标签+签到激活 | +特病认证+民政补助 | +公交优惠+残疾补贴+GPS定位 |
| **研发BOM** | ~80-120元 | ~180-230元 | ~280-400元 |

---

## 实施批次与顺序

```
第一批 (月1-6):  ①手环固件(三档) → ③云平台后端 ← ②药盒固件(三档)
                     ↓            ↓            ↓
第二批 (月4-8):            ④家属APP    ⑤管理后台
                                    ↓
第三批 (月7-12):              ⑦小程序   ⑥官网
                                    ↓
                               ⑧B2B医院/社区对接
                                    ↓
第四批 (月10-18):     ⑨医用腕带监管闭环 + ⑩社区老人多维身份腕带
```

**关键路径：** ①手环固件 + ②药盒固件 + ③云平台后端 并行推进，第4个月开始④家属APP联调。第10个月开始⑨医用腕带监管闭环与⑩社区老人场景并行推进。

---

## 当前阶段实现状态 (2026-07-28)

代码审计与专项计划编制完成后，各子系统建设内容与执行方案如下：

### 第一批 → 已完成并进入维护优化阶段

| # | 子系统 | 建设内容 | 实施策略 | 专项计划文档 |
|---|--------|---------|----------|-------------|
| ③-1 | API Server REST API | 用户、设备、健康数据、告警等完整 CRUD 接口，JWT 认证中间件，错误信息统一封装 | Go Gin 微服务架构，分层处理逻辑 (handler/store/model)，依赖注入 | - |
| ③-2 | Admin API | 管理后台专用 API：老人档案、用药规则、固件版本、监管规则配置 | PostgreSQL/SQLite 双存储抽象层，JWT scope-based 权限控制 | - |
| ③-3 | Gateway MQTT | EMQX MQTT 客户端接入，设备自动注册，消息路由到 NATS 总线 | 独立服务通过 topic 路由识别设备类型，自动注册未绑定设备 | - |
| ③-4 | Push Service | FCM/APNs 推送通知通道，短信兜底（阿里云 SMS），微信订阅消息触发 | 基于 NATS 事件驱动架构，按 P0/P1/P2 优先级分发 | - |
| ③-5 | Data Pipeline | 时序数据存储，AI 分析引擎框架，健康异常检测模型占位符 | SQLite 时序表 + PostgreSQL 降采样 fallback，pgx 驱动预留 | - |
| ⑧-1 | Hospital API (B2B) | HIS 预留接口框架，入院/出院/护士核验终端绑定，监管规则评估 R_C01-R_C10 | APIKeyAuth 认证 + RequireAccess 访问控制，institution-aware | `shared/auth/middleware/auth.go` |
| ⑧-2 | Community Platform (B2B) | 社区老人档案管理、福利标签分配、批量支付结算、民政同步 | Same APIKeyAuth pattern, b2b_institutions table seeded via migration | `migrations/001_init_institutions.sql` |
| ⑧-3 | Insurance Integration (B2B) | 保险provider注册，费用结算对账，医保覆盖额度管理 | Same auth pattern, encrypted policy keys storage | `migrations/001_init_institutions.sql` |

### 第二批 → 已完成

| # | 子系统 | 建设内容 | 实施策略 | 参考文档 |
|---|--------|---------|----------|---------|
| ④ | 家属 App Flutter | 实时定位地图+SOS按钮+用药提醒+健康趋势图+告警中心 | Flutter模块化架构，Provider/Riverpod状态管理，dio API层 | `docs/specs/04-family-app.md` |
| ⑤ | 管理后台 Vue 3 | 仪表盘总览+设备列表+老人档案+用药管理+告警中心+系统设置 | Vue3 Composition API + Pinia + Element Plus + Axios interceptor | `docs/specs/05-admin-web.md` |

### 第三批 → 已完成

| # | 子系统 | 建设内容 | 实施策略 | 参考文档 |
|---|--------|---------|----------|---------|
| ⑦ | 微信小程序 | 轻量版家属端(地图定位+用药提醒+紧急呼叫) | 原生 WXML/WXSS, Tencent Map插件, 微信订阅消息 | `docs/specs/07-miniprogram.md` |
| ⑥ | 品牌官网 Hugo | 产品介绍三档对比页+购买引导+联系表单+关于页面 | Hugo静态生成, Tailwind CSS响应式布局 | `docs/specs/06-website.md` |
| ⑨ | 医用电子腕带 | NFC身份核验 + Cat1上报双模式，医疗护士终端交互协议 | ESP32-S3 + BLE/NFC双芯片设计，wb_ble.go 协议文档已更新 | — (protocol doc updated) |
| ⑩ | 护士核验终端 Flutter | PDA手持设备的NFC扫描验证 + 医嘱执行记录 | Flutter移动App，nfc_plus插件，完成wb_ble.go的VerificationReport映射 | `docs/specs/11-nurse-terminal.md` |

### 第四批 → 硬件采购后启动

| # | 子系统 | 建设内容 | 实施条件 | 备注 |
|---|--------|---------|---------|------|
| ① | 手环固件 (Entry/Plus/Pro) | GD32E230 FreeRTOS C工程，传感器驱动(GPS/PPG/IMU)/BLE低功耗通信 | 需采购GD32E230开发板+GPS模组+PPG传感器 | See firmware/bracelet/目录框架 |
| ② | 药盒固件 (Basic/Smart/Auto) | ESP32-C3 ESP-IDF C工程，电机控制/语音TTS/WiFi连接 | 需采购ESP32-C3开发板+步进电机+语音模块 | See firmware/pillbox/目录框架 |

> **重要说明**：以上任务顺序严格按 CLAUDE.md 原始实施批次安排。各子系统规格详见 `docs/specs/` 目录。已实施的专项计划文件已归档。

---

### 设备-云端通信协议（核心接口定义）

### 上行(设备→云端)

```json
// 心跳包
{"type":"heartbeat","dev_id":"BR-XXXX","bat":85}

// 定位数据
{"type":"location","dev_id":"BR-XXXX","lat":31.xxx,"lon":121.xxx,"acc":5,"ts":1720000000}

// 健康数据
{"type":"health","dev_id":"BR-XXXX","hr":72,"spo2":98,"step":3456}

// SOS告警
{"type":"sos","dev_id":"BR-XXXX","lat":xxx,"lon":xxx,"ts":xxx}

// 跌倒检测
{"type":"fall","dev_id":"BR-XXXX","conf":0.95,"lat":xxx,"lon":xxx}

// 药盒状态
{"type":"med_status","dev_id":"PX-XXXX","compartment":3,"taken":true,"ts":xxx}
```

### 下行(云端→设备)

```json
// 用药规则
{"type":"med_rule","dev_id":"PX-XXXX","rules":[{"time":"08:00","dose":1,"type":"capsule"}]}

// 配置更新
{"type":"config","dev_id":"BR-XXXX","settings":{"interval":30,"volume":80}}

// 语音播报
{"type":"tts","dev_id":"PX-XXXX","text":"爷爷，该吃降压药了"}

// OTA升级
{"type":"ota","dev_id":"BR-XXXX","url":"https://...","hash":"sha256:..."}
```

---

## 数据流闭环

```
[手环传感器]──Cat1蜂窝──→[EMQX MQTT]──→[Go设备接入]──→[NATS总线]
                                                        ↓
[药盒电机]────WiFi────────→[EMQX MQTT]──→[Go设备接入]──→[NATS总线]
                                                        ↓
                                            ┌───────────┼───────────┐
                                            ↓           ↓           ↓
                                      [SQLite]
                                      用户/设备/订阅  健康时序数据   在线状态
                                            ↓           ↓           ↓
                                            └───────────┼───────────┘
                                                        ↓
                                            [AI分析引擎]←──[实时数据流]
                                                ↓
                                    跌倒/异常→P0→WebSocket推送
                                    漏服药物→P1→短信兜底
                                    电子围栏→P1→APP推送
                                                ↓
                                          [推送分发器]
                                                ↓
                                    [家属APP] ←→ [管理后台]
```

---

## 技术预研成本（非市场性）

### 硬件采购清单（研发阶段）

| 物料 | 型号 | 数量 | 单价(元) | 总价(元) | 用途 |
|------|------|------|---------|---------|------|
| GD32E230C8T3开发板 | GD32E230G-START | 3 | 50 | 150 | 手环固件开发 |
| ESP32-C3开发板 | ESP32-C3-DevKitM-1 | 3 | 25 | 75 | 药盒固件开发 |
| GPS模组(国产) | 和芯星通UGN-7345 | 3 | 80 | 240 | 手环定位验证 |
| GPS模组(进口) | u-blox NEO-M9N | 2 | 350 | 700 | 高端版验证 |
| Cat1模组 | 广和通L610-CM | 3 | 60 | 180 | 手环通信验证 |
| PPG传感器 | 汇顶GT320 | 5 | 40 | 200 | 健康监测验证 |
| IMU传感器 | ICM-42670-P eval | 3 | 30 | 90 | 跌倒检测验证 |
| 步进电机 | 28BYJ-48 | 10 | 5 | 50 | 药盒分药验证 |
| 语音模块 | SYN5300 | 10 | 8 | 80 | 药盒TTS验证 |
| OLED屏 | SSD1306 0.96" | 5 | 12 | 60 | 药盒显示验证 |
| 锂电池 | 350mAh LiPo | 5 | 10 | 50 | 电源验证 |
| 3D打印外壳 | PLA材料 | 10套 | 20 | 200 | 结构验证 |
| **硬件采购合计** | | | | **~2075元** | |

### 软件基础设施成本

| 项目 | 方案 | 月成本(元) | 年成本(元) |
|------|------|-----------|-----------|
| 云服务器 | 阿里云轻量应用服务器2核4G | ~50 | ~600 |
| 域名 | 阿里云.com域名 | ~7 | ~84 |
| SSL证书 | Let's Encrypt免费 | 0 | 0 |
| MQTT Broker | EMQX Docker自建 | 含在服务器中 | — |
| 数据库 | SQLite（零部署） | — | — |
| 推送服务 | FCM免费 + 阿里云SMS按量 | ~10 | ~120 |
| **软件合计** | | **~67元/月** | **~804元/年** |

### 研发总成本

| 类别 | 金额 |
|------|------|
| 硬件采购 | ~2,075元 |
| 首年软件基础设施 | ~804元 |
| **总计** | **约3,000元** |

> 注意：这是纯技术预研可行性开发的成本。不含市场定价、ROI分析、订阅收入预测等市场性内容。
> 实际量产成本另计，本阶段只验证功能可行性。

---

## 服务启动与管理

所有子系统的启动、停止、状态查看、日志管理通过统一脚本系统完成：

```bash
# 环境配置（首次）
cp scripts/default-ports.env .env

# 依赖检查
./scripts/start.sh check-deps

# 端口冲突检测
./scripts/start.sh ports-check

# 启动服务
./scripts/start.sh start <service|--all|--group>    # 启动服务
./scripts/start.sh stop <service|--all|--group>     # 停止服务
./scripts/start.sh restart <service>                # 重启服务
./scripts/start.sh status [--all]                    # 查看状态
./scripts/start.sh logs <service|--all>              # 查看日志
./scripts/start.sh clean                             # 清理运行时文件
./scripts/start.sh start --docker                    # Docker 模式
```

**按组启动：** `cloud` (api-server, push-service, data-pipeline, admin-api, gateway) / `b2b` (hospital-api, community-platform, insurance-integration) / `apps` (family-app, admin-web, website) / `medical` (admin-api regulatory/, admin-api community_wb/, gateway community handlers)

**端口配置：** 编辑根目录 `.env` 中的 `PORT_*` 变量，或命令行 `--port X` 覆盖。

**子系统独立启动：** 每个子系统目录下有自己的 `start.sh`，如 `cd cloud/api-server && ./start.sh`。

详细使用说明见根目录 `README.md` §6。

### 高级特性

统一启动脚本增强以下功能：

- **端口冲突检测与清理**：启动前自动检测目标端口是否被占用，对冲突进程执行 `SIGTERM→SIGKILL` 级联清理（排除 `ssh`、`sshd`、`login` 等关键系统进程）。支持 `--force` 参数强制执行清理。
- **健康检查严格模式**：服务启动后轮询 `/api/v1/health` 端点，30 秒内连续 2 次成功返回 `{"data": {"status": "ok"}}` 则标记为就绪，否则判定启动失败。
- **PID 文件管理**：所有后台服务 PID 写入 `$HOME/.eregen/pids/<service>.pid`，支持基于 PID 文件优雅停止和 stale 进程清理。
- **多模式支持**：通过环境变量控制运行模式（`DEV`/`DEMO`/`CI/CD`/`PRODUCTION`），各模式具有不同的冲突处理策略和健康检查超时阈值。

---

## UI界面原型工作流

**所有应用层界面必须先出高保真HTML效果图，确认后才可以写Flutter/Vue代码。**

### 需出效果图的界面清单

| # | 子系统 | 界面名称 | 核心元素 |
|---|--------|---------|---------|
| 1 | 家属APP | 首页-实时定位 | 高德地图+老人位置标记+电子围栏+快速状态卡片 |
| 2 | 家属APP | 健康数据看板 | 心率/血氧/步数趋势图+异常告警提示 |
| 3 | 家属APP | SOS告警中心 | 紧急告警列表+处理记录+一键呼叫 |
| 4 | 家属APP | 用药管理 | 今日用药提醒+服药记录+远程配置 |
| 5 | 管理后台 | 仪表盘总览 | 设备在线率/告警统计/用户活跃度/订阅转化 |
| 6 | 管理后台 | 设备管理 | 设备列表+状态筛选+固件版本+OTA升级 |
| 7 | 管理后台 | 用户管理 | 老人/家属/机构用户列表+权限管理 |
| 8 | 管理后台 | 订阅管理 | 订阅状态+续费记录+降级原因 |
| 9 | 品牌官网 | 首页 | Eregen品牌展示+产品介绍+购买入口 |
| 10 | 小程序 | 首页 | 轻量版家属端-定位+用药提醒 |
| 11 | 管理后台 | 监管总览看板 | 在院患者摘要+异常告警列表+规则引擎状态+合规报表入口 |
| 12 | 管理后台 | 穿透审计详情页 | 患者全链路数据追溯（入院→核验→用药→围栏→出院） |
| 13 | 管理后台 | 社区老人档案页 | 老人信息+福利标签+签到记录+腕带绑定 |
| 14 | 家属APP | 父母福利页 | 福利标签状态+补助领取记录+签到提醒 |

### 执行顺序

1. 我编写HTML/CSS/JS高保真原型文件 → 放在 `apps/ui-prototypes/`
2. 你在浏览器中打开查看效果
3. 你确认满意或提出修改意见
4. 确认后，我基于原型编写Flutter/Vue实际代码
5. **未确认原型的界面绝对不写代码**

---

## 工作原则

1. **核心代码自研** — 所有业务逻辑、算法、协议自己写，不外包
2. **你负责硬件采购和实物** — 我负责所有代码编写
3. **专利保护优先** — 只用MIT/BSD/Apache-2.0/ISC开源许可，禁用GPL/AGPL/LGPL
4. **技术选型一次性确定** — 本文档中的选型整个项目周期不变
5. **每个子系统独立开发测试** — 明确的输入/输出边界，可并行推进
6. **硬件三档策略** — 入门版快速上市验证，高端版建立品牌溢价
7. **混合ODM起步** — Phase 1A公模快速验证，Phase 1B半定制差异化迭代

---

## 跨会话持久化机制

本文件(`CLAUDE.md`)是项目的唯一真相来源。每次新会话启动时：
1. 读取本文件确认技术选型和项目目标
2. 检查`docs/specs/`下的设计文档了解详细设计
3. 检查各子项目目录了解当前开发进度
4. 如有变更，先更新本文件和对应设计文档

**绝对不要**在没有更新本文件的情况下更改任何技术选型。

---

## GitHub 仓库管理准则（重要）

**GitHub 仓库仅用于代码版本管理。以下文件绝对禁止提交到远程仓库：**

- **所有文档文件**：`docs/` 目录下的设计文档、架构文档、规格说明、方案文档等（`.md`, `.txt`, `.png`, `.jpg` 等）
- **项目描述文件**：`README.md`、`slogan.md`、`product-overview.md` 等全盘文字描述项目的文件
- **工具配置文件**：`.claude/`、`.serena/`、`.codegraph/`、`.playwright-mcp/` 等 AI 工具和本地配置
- **CI/CD 工作流**：`.github/workflows/`（避免暴露内部工具链信息）
- **超能力相关**：`docs/superpowers/specs/` 目录下的设计决策（已实施部分）

**允许提交的：**
- 源代码文件（`.go`, `.dart`, `.ts`, `.vue`, `.js`, `.wxml`, `.wxss`, `.c`, `.h` 等）
- 构建配置文件（`go.mod`, `pubspec.yaml`, `package.json`, `tsconfig.json`, `CMakeLists.txt` 等）
- `.gitignore`

**核心原则：** 我们只使用 GitHub 的代码管理能力做版本迭代，所有涉及项目方案、架构设计、功能描述的文档只保留在本地。这是保证项目不泄密的关键准则。

## Agent skills

### Issue tracker

Issues tracked via GitHub Issues using the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Five canonical labels: needs-triage, needs-info, ready-for-agent, ready-for-human, wontfix. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout with `CONTEXT.md` at root and `docs/adr/` for ADRs. See `docs/agents/domain.md`.
