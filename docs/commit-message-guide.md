# 提交消息规范

## 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

## Type 类型

| Type | 说明 | 示例 |
|------|------|------|
| feat | 新功能 | feat(admin-web): add user management |
| fix | bug 修复 | fix(api-server): resolve timeout issue |
| docs | 文档更新 | docs: update README |
| style | 代码格式 | style: format Go code |
| refactor | 重构 | refactor(api): simplify auth logic |
| test | 测试 | test: add unit tests for auth |
| chore | 构建/工具 | chore: update dependencies |
| perf | 性能优化 | perf(api): optimize query performance |
| revert | 回滚 | revert: rollback broken commit |

## Scope 范围

| Scope | 说明 |
|-------|------|
| admin-api | 管理后台 API |
| api-server | 核心 API 服务 |
| gateway | MQTT 网关 |
| push-service | 推送服务 |
| data-pipeline | 数据分析 |
| admin-web | 管理后台前端 |
| family-app | 家属 APP |
| miniprogram | 微信小程序 |
| nurse-terminal | 护士终端 |
| website | 品牌官网 |
| hospital-api | 医院对接 API |
| community | 社区平台 |
| insurance | 保险对接 |
| shared | 共享库 |
| firmware | 硬件固件 |
| docs | 文档 |
| infra | 基础设施 |

## 示例

```
feat(admin-web): add medical wristband workflow

- Implement admission/discharge patient flow
- Add ward round completion tracking
- Integrate with hospital API endpoints

Closes #123
```

```
fix(api-server): resolve device heartbeat timeout

Devices were timing out after 30s instead of 60s.
Updated heartbeat interval configuration.

Fixes #456
```

```
docs: update ARCHITECTURE.md with new B2B services

Added hospital-api, community-platform, and
insurance-integration to system diagram.
```

---
最后更新：2026-08-02
