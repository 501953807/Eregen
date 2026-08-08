# Eregen 系统改进任务计划

> 编制日期：2026-07-26
> 状态：执行中

---

## 用户反馈的5个核心问题

### 1. 语言强制：所有对话必须使用中文

**现状：** 模型在几天后忘记了中文指令。

**解决方案：**
- 在所有 Flutter 应用的 UI 文本、提示消息、错误信息中强制使用中文
- 在 Go 后端的服务端错误消息中使用中文
- 在 Vue 管理后台的所有界面文本中使用中文
- 在数据库注释和 API 响应中使用中文

**涉及文件：**
- `apps/family-app/lib/common/theme.dart` — 主题色等常量
- `apps/family-app/lib/api/client.dart` — API 客户端
- `apps/family-app/lib/screens/**/*.dart` — 所有页面
- `cloud/admin-api/internal/handler/*.go` — 所有处理器
- `apps/admin-web/src/**/*.vue` — 所有管理后台页面

**验证方法：**
- 运行 `grep -r "login\|error\|success" apps/family-app/lib/ --include="*.dart" | grep -v "//"` 检查是否有英文文本
- 运行 `grep -r "\"error\"\|\"success\"" cloud/admin-api/ --include="*.go"` 检查 Go 代码

---

### 2. 数据库策略：MVP 阶段仅使用 SQLite，但保证可迁移到 PostgreSQL

**现状：** 部分代码已使用 SQLite，但未系统化设计迁移路径。

**解决方案：**
- 统一 `Store` 层抽象接口，定义所有数据访问方法
- 实现 `SQLiteStore`（MVP）和 `PostgreSQLStore`（生产）两个实现
- 所有业务逻辑通过接口访问 Store，不直接操作数据库
- 使用工厂模式根据环境变量选择 Store 实现
- 提供迁移脚本从 SQLite 到 PostgreSQL

**涉及文件：**
- `cloud/api-server/internal/store/store.go` — Store 接口定义
- `cloud/api-server/internal/store/sqlite_store.go` — SQLite 实现
- `cloud/api-server/internal/store/postgres_store.go` — PostgreSQL 实现
- `cloud/api-server/cmd/server/main.go` — 启动时选择 Store

**验证方法：**
- 单元测试覆盖 Store 接口
- 集成测试验证 SQLite → PostgreSQL 迁移脚本

---

### 3. 缺失功能：家属 APP 可查看住院病人每日用药、检查、治疗情况

**现状：** 医用电子腕带监管业务中，此功能未在任何子系统中实现。

**解决方案：**
- 在 `cloud/admin-api` 中添加 REST API 端点：
  - `GET /api/v1/medical/patient/:id/daily-summary` — 获取某患者某日的汇总
  - `GET /api/v1/medical/patient/:id/medications` — 获取用药清单
  - `GET /api/v1/medical/patient/:id/tests` — 获取检查清单
  - `GET /api/v1/medical/patient/:id/treatments` — 获取治疗清单
- 在 `apps/family-app` 中添加新页面 `HospitalizationPage`：
  - 显示当前住院父母的基本信息
  - Tab 切换：用药 / 检查 / 治疗
  - 每条记录显示：项目名、时间、状态、执行人
- 在 `apps/miniprogram` 中添加简化版页面

**涉及文件：**
- `cloud/admin-api/internal/handler/medical_handler.go` — 新增 API 处理器
- `cloud/admin-api/internal/service/medical_svc.go` — 新增服务层
- `cloud/admin-api/internal/store/sqlite.go` — 新增查询方法
- `apps/family-app/lib/screens/hospitalization/hospitalization_page.dart` — 新页面
- `apps/family-app/lib/api/client.dart` — 新增 API 调用方法

**验证方法：**
- API 测试：curl 调用端点验证返回格式
- Flutter 测试：打开 HospitalizationPage 验证 UI 渲染
- 集成测试：端到端流程（入院 → 录入 → 查看）

---

### 4. 文档同步：阅读 AnaReports 并更新系统文档

**现状：** `docs/specs/` 和各 README.md 未包含最新的医用腕带和社区老人方案。

**解决方案：**
- 读取所有 AnaReports 目录下的设计文档
- 更新 `docs/specs/` 中的系统设计文档
- 更新各子系统 README.md
- 更新根目录 README.md

**涉及文件：**
- `docs/specs/01-system-overview.md` — 更新系统架构
- `docs/specs/02-hardware-specs.md` — 更新硬件规格
- `docs/specs/03-cloud-backend.md` — 更新云后端设计
- `docs/specs/04-family-app.md` — 更新家属 APP 设计
- `docs/specs/05-admin-web.md` — 更新管理后台设计
- `docs/specs/06-miniprogram.md` — 更新小程序设计
- `apps/family-app/README.md` — 更新家属 APP 说明
- `apps/admin-web/README.md` — 更新管理后台说明
- `cloud/admin-api/README.md` — 更新云后端说明
- `README.md` — 更新根文档

**验证方法：**
- 对比 AnaReports 文档和更新后的 specs 文档，确保所有新模块已涵盖
- 检查每个 README.md 是否包含最新的功能列表

---

### 5. 子系统独立性：每个子系统可单独运行和测试

**现状：** 部分子系统启动脚本存在问题（如启动所有服务而非单个服务）。

**解决方案：**
- 确保每个子系统有独立的 `start.sh` 或 `Makefile`
- 确保每个子系统可以独立测试（单元 + 集成）
- 提供 Docker Compose 配置用于本地开发环境
- 编写 CI/CD 工作流验证每个子系统独立性

**涉及文件：**
- `cloud/api-server/start.sh` — 修复启动脚本
- `apps/family-app/start.sh` — 添加 Flutter web 启动脚本
- `apps/admin-web/start.sh` — 添加 Vue 启动脚本
- `.github/workflows/ci.yml` — 添加 CI 工作流
- `docker-compose.dev.yml` — 添加开发环境配置

**验证方法：**
- 分别启动每个子系统验证无冲突
- 分别运行每个子系统的测试套件
- 验证 Docker Compose 配置可启动完整开发环境

---

## 实施批次

```
第一批 (T+1天):   语言强制 + 子系统独立性
├── 修复所有 Flutter UI 文本为中文
├── 修复所有 Go 服务端错误消息为中文
├── 修复 start.sh 脚本确保只启动对应服务
└── 验证每个子系统独立运行

第二批 (T+2天):   数据库策略
├── 定义 Store 接口
├── 实现 SQLiteStore
├── 实现 PostgreSQLStore（骨架）
└── 编写迁移脚本

第三批 (T+3天):   缺失功能
├── 添加住院每日汇总 API
├── 添加 HospitalizationPage
└── 端到端联调

第四批 (T+4天):   文档同步
├── 更新 docs/specs/
├── 更新各子系统 README.md
└── 更新根 README.md
```

---

## 完成标准

每项任务完成后需满足：
1. 代码通过静态分析（Dart analyze / Go vet）
2. 单元测试通过率 ≥ 90%
3. 集成测试通过
4. 相关文档已更新
5. 用户确认验收
