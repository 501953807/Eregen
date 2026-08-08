# 系统完善计划 — Spec

## Problem Statement

Eregen 平台经过第一阶段开发后，11 个子系统中存在规格文档与代码实现的偏差、部分功能仅骨架 stub、部分 API 路径不一致、以及关键模块缺失等问题。本计划明确所有需要完善的内容，按优先级和依赖关系排序，形成可执行的 issue 列表。

## Solution

按依赖关系排序的 17 个垂直切片，每个切片独立完成一个端到端功能：

1. **API 路径对齐**（P0，无依赖）— 修复 family-app 和 miniprogram 中医疗 API 路径前缀不一致问题
2. **规则引擎补全**（P0，依赖 #1）— 实现 R04/R05/R06/R08 + R_C01~C08 共 12 条规则
3. **护士终端 NFC 服务**（P0，依赖 #1）— 实现 nfc_wristband_service.dart 替换 BLE 方案
4. **医用腕带固件框架**（P0，依赖 #1）— 完善 medical-wristband 固件工程结构
5. **Entry 固件偏差修正**（P1）— 用条件编译分离 Entry/Plus/Pro 功能边界
6. **Basic 药盒 WiFi 矛盾修正**（P1）— 删除 Basic 版 WiFi 模块或更新 spec
7. **订阅管理 API 补全**（P1）— 后端补列表端点，前端补 store/API
8. **机构管理 CRUD 补全**（P1）— 后端补 CRUD API，前端补操作
9. **Medication.vue bug 修复**（P1）— 修复表单数据绑定类型安全问题
10. **CommunityWristband mock 替换**（P1，依赖 #1）— 替换药房/签到 mock 为真实 API
11. **推送服务优先级路由**（P2）— 确认并完善 P0/P1/P2 分级推送
12. **推送日志表完善**（P2，依赖 #11）— 确认 push_logs 写入逻辑
13. **Gateway medical/community topic**（P2，依赖 #1）— 确认并补全 medical_wb/community_wb topic 处理
14. **品牌官网内容完善**（P2）— 补充产品详情和法律页面
15. **审计日志完善**（P2，依赖 #1）— 补充操作追溯和权限变更审计
16. **Nurse terminal 服务层薄补强**（P2）— 扩充 verification/patient/ward_round service
17. **固件 OTA 模块补充**（P2）— 为 Plus/Pro/Smar 版本补全 OTA 模块

## User Stories

### P0 阻断性

1. As a family app user, I want the hospitalization page to call the correct API path (`/api/v1/medical/*`), so that patient data loads correctly from the backend.
2. As a nurse using the terminal, I want to scan a medical wristband via NFC (not BLE) so that I can read patient identity within 4cm proximity.
3. As a regulator, I want all 16 compliance rules (R01-R08 + R_C01-R_C08) to run every 5 minutes, so that hospital fraud and community welfare abuse are detected.
4. As a hospital administrator, I want the medical wristband firmware to have a complete build structure (NVS store, NFC server, OLED display, LED indicator, security), so that the hardware can be tested on real ESP32-S3 boards.

### P1 重要功能

5. As a firmware developer, I want Entry/Plus/Pro features to be clearly separated by compile-time flags, so that the spec's tier matrix is accurate in code.
6. As a product manager, I want the Basic pillbox spec to match its code (either remove WiFi or update the spec), so that there is no contradiction.
7. As an operations admin, I want to see a list of all subscriptions (not just stats), so that I can manage renewals and downgrades.
8. As a B2B manager, I want full CRUD for institutions (hospital/community), so that I can manage partner connections.
9. As an admin, I want the medication rule form to work without TypeScript errors, so that I can create and update medication rules correctly.
10. As a community health worker, I want the CommunityWristband page to show real pharmacy and sign-in data, so that I can track actual welfare distribution.

### P2 优化完善

11. As a family app user, I want P0 alerts (SOS/fall) to trigger SMS fallback when FCM fails, so that I never miss critical alerts.
12. As an operator, I want every push notification to be logged in the database, so that I can audit delivery status.
13. As a platform operator, I want the gateway to properly handle `eregen/medical/wb/#` and `eregen/community/wb/#` MQTT topics, so that wristband data flows to the correct NATS channels.
14. As a website visitor, I want detailed product pages for bracelets and pillboxes, so that I can understand the full specifications.
15. As a compliance auditor, I want all role changes and sensitive operations logged with timestamps and actor info, so that the audit trail is complete.
16. As a nurse, I want the verification and ward-round services to have full CRUD operations, so that I can manage patient interactions efficiently.
17. As a firmware developer, I want OTA module stubs for Plus/Pro/Smart variants, so that the build structure matches the spec's OTA requirements.

## Implementation Decisions

- **API paths**: Family app and miniprogram should use `/api/v1/medical/*` prefix to match admin-api router. The admin-api already exposes these endpoints under `/api/v1/medical/*`.
- **NFC vs BLE**: The nurse terminal spec requires NFC A protocol (106 kbps, 4cm range). The current `medical_wristband_ble_service.dart` uses BLE which is a different protocol. Must replace with `nfc_wristband_service.dart` using `nfc_manager` package.
- **Rule engine**: The 16 rules share a single `RuleEngine` struct with `map[string]RuleFunc`. Each rule is an independent function `EvaluateRXX(ctx, event) *Alert`. R04-R06/R08 and R_C01-R_C08 are implemented as empty/no-op or missing.
- **Firmware**: Medical wristband firmware is ESP32-S3 based, existing stubs only have `main.c`. Need to add patient_store, nfc_server, display_oled, led_indicator, nvs_manager, security modules.
- **Entry/Basic deviations**: These are intentional over-implementation (features beyond spec). Solution: add `#ifdef TARGET_TIER` guards rather than deleting code.
- **Subscription/Institution**: Backend has store layer but handler/router may be missing list endpoints. Check admin-api router registration.

## Testing Decisions

- Each ticket should have at minimum a unit test for the new rule/function
- NFC service: integration test with mock NfcTag
- Rule engine: test each rule independently, verify 16 rules all registered
- Family-app API path fix: verify HTTP requests use correct base path
-固件: compile test for each variant (entry/plus/pro/smart/auto/basic)

## Out of Scope

- B2B HIS 预留接口（spec §9 明确 MVP 暂不启用）
- Hardware procurement
- 微信小程序订阅消息模板 ID 配置（需微信后台操作）
- 生产环境部署配置

## Further Notes

- 所有 tickets 共享依赖 #1（API 路径对齐），因为它影响 family-app、miniprogram、nurse_terminal 三个子系统
- 固件开发（#4）需要实际硬件（ESP32-S3）才能验证，MVP 阶段以编译通过为验收标准
- Basic 药盒的 WiFi 矛盾：建议删除 `wifi_station.c` 或将 Basic 重新定义为"基础电子版（有 WiFi）"并更新 spec
