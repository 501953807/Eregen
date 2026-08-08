# 子系统顺序验证设计方案（2026-07-26）

## 1. 目标与范围

**目标**：按照顺序启动并验证所有应用层子系统的可用性，确认每个子系统能否成功启动并在浏览器中正常显示界面。对于发现的问题，逐项补齐欠缺内容后重新验证。

**验证深度**：B) 启动+访问 — 通过 start.sh 启动各子系统 dev server，使用浏览器打开 URL 截图验证 UI 是否正常渲染。

**涉及子系统（按验证顺序）**：
1. admin-web (Vue 3 + Vite)
2. family-app (Flutter web-server)
3. nurse_terminal (Flutter web-server)
4. website (Hugo dev server)
5. miniprogram (WeChat DevTools — 需手动交互)
6. ui-prototypes (HTML 静态文件 — 检查完整性)
7. cloud backend (Go microservices — 依赖基础设施)

---

## 2. 前置条件检查

| 检查项 | 预期状态 | 当前状态 |
|--------|---------|---------|
| Flutter SDK | `flutter` 命令可用 | ✅ /opt/homebrew/bin/flutter |
| Node.js | `node`/`npm` 命令可用 | ✅ v26.3.0 / npm 11.16.0 |
| Hugo | `hugo` 命令可用 | ✅ v0.164.0+extended |
| .env 配置文件 | 存在且配置端口 | ✅ /Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/.env |
| 依赖安装完成 | node_modules 目录存在 | ✅ admin-web 已安装 |

---

## 3. 子系统验证方案

### 3.1 admin-web (Vue 3 管理后台)

**启动脚本**：`./apps/admin-web/start.sh`
**启动命令**：`(cd apps/admin-web && npx vite --host 0.0.0.0 --port 3001)`
**预期端口**：3001（来自 .env 中 PORT_ADMIN_WEB）
**预期 URL**：http://localhost:3001/
**验证要点**：
- Vite dev server 正常启动
- HTML 响应内容为 Vue 应用首页
- 页面加载无 404/5xx 错误

**欠缺内容补充策略**：
- 若 node_modules 缺失 → 执行 `npm install`
- 若端口被占用 → 调整 PORT_ADMIN_WEB 配置
- 若页面白屏 → 检查浏览器控制台错误日志

---

### 3.2 family-app (Flutter 家属端)

**启动脚本**：`./apps/family-app/start.sh`
**启动命令**：`flutter run -d web-server --web-port=5173`
**预期端口**：5173（来自 .env 中 PORT_FAMILY_APP）
**预期 URL**：http://localhost:5173/
**验证要点**：
- Flutter web-server 正常启动
- 应用界面可正常加载
- 主要页面路由可访问（首页、定位、健康等）

**欠缺内容补充策略**：
- 若 AMAP_KEY 未设置 → 添加 `--dart-define=AMAP_KEY=...`（或暂时Mock地图数据）
- 若依赖缺失 → 执行 `flutter pub get`
- 若端口冲突 → 调整 PORT_FAMILY_APP 配置

---

### 3.3 nurse_terminal (Flutter 护士终端)

**启动脚本**：`./apps/nurse_terminal/start.sh`
**启动命令**：`flutter run -d web-server --web-port=5174`
**预期端口**：5174（来自 .env 中 PORT_NURSE_TERMINAL）
**预期 URL**：http://localhost:5174/
**验证要点**：
- Flutter web-server 正常启动
- 登录页及主界面可正常显示
- API 集成相关功能可用（如登录成功）

**欠缺内容补充策略**：
- 同 family-app，处理 AMAP_KEY、依赖缺失等问题
- 若需要后端接口验证 → 提前 mock API 响应或暂时跳过

---

### 3.4 website (Hugo 官网)

**启动脚本**：`./apps/website/start.sh`
**启动命令**：`hugo server --bind 0.0.0.0 --port 1313`
**预期端口**：1313（来自 .env 中 PORT_WEBSITE）
**预期 URL**：http://localhost:1313/
**验证要点**：
- Hugo dev server 正常启动
- 官网首页正确渲染
- 静态资源无 404

**欠缺内容补充策略**：
- 若 Hugo 未安装 → `brew install hugo` 或下载安装
- 若配置错误 → 检查 hugo.toml 配置项

---

### 3.5 miniprogram (微信小程序)

**启动脚本**：`./apps/miniprogram/start.sh [open|build]`
**预期操作**：手动打开微信开发者工具，导入 `apps/miniprogram` 目录
**验证要点**：
- 项目目录结构完整（app.json, app.js, pages 等）
- 无语法报错
- 基础页面可预览

**欠缺内容补充策略**：
- 若缺少 WeChat CLI → 指导安装 `@wechat/miniprogram-cli`
- 若项目导入失败 → 检查目录结构和入口文件

---

### 3.6 ui-prototypes (HTML 原型文件)

**检查方式**：`ls apps/ui-prototypes/`
**验证要点**：
- HTML 原型文件是否存在且非空
- 关键界面原型是否齐全（根据 CLAUDE.md 要求列出的12个界面）

**欠缺内容补充策略**：
- 若有文件丢失 → 从 git history 恢复或重新制作高保真效果图
- 若界面缺失 → 按标准制作对应 HTML/CSS 文件

---

### 3.7 cloud backend (Go 微服务集群)

**启动方式**：`./scripts/start.sh start cloud` 或 Docker Compose
**预期服务**：api-server、push-service、data-pipeline、admin-api、gateway
**验证要点**：
- Go 代码编译通过 (`go build`)
- 各服务能正常启动（需外部依赖：PostgreSQL、Redis、EMQX、NATS）
- HTTP 端口可访问（若无则视为 MQTT 专有的 gateway）

**欠缺内容补充策略**：
- 若依赖服务未运行 → 先启动基础设施（docker compose up -d），或用 Mock 替代
- 若环境变量缺失 → 配置 `.env` 中的 PORT_* 变量
- 若编译错误 → 修复依赖版本问题

---

## 4. 验证工作流

```bash
# 步骤1: 启动 admin-web
./apps/admin-web/start.sh
sleep 10
curl -s http://localhost:3001 > admin_web_snapshot.html  # 保存快照
# 查看 admin_web_snapshot.html 确认内容正常

# 步骤2: 停止 admin-web, 启动 family-app
./apps/admin-web/stop.sh
./apps/family-app/start.sh
sleep 15
curl -s http://localhost:5173 > family_app_snapshot.html

# 步骤3: 停止 family-app, 启动 nurse-terminal
./apps/family-app/stop.sh
./apps/nurse_terminal/start.sh
sleep 15
curl -s http://localhost:5174 > nurse_terminal_snapshot.html

# 步骤4: 停止 nurse-terminal, 启动 website
./apps/nurse_terminal/stop.sh
./apps/website/start.sh
sleep 10
curl -s http://localhost:1313 > website_snapshot.html

# 步骤5: miniprogram - 手动验证
./apps/miniprogram/start.sh open  # 提示用户手动操作

# 步骤6: ui-prototypes - 检查文件完整性
ls -la apps/ui-prototypes/

# 步骤7: cloud backend - 依赖检查（可选，需额外环境）
./scripts/start.sh check-deps
```

每个步骤完成后，检查保存的快照文件：
- HTML 是否存在且非空
- 页面标题是否符合预期（应为对应系统名称）
- 无明显的错误或空白标记

发现问题立即补齐后重新启动验证，直至该子系统 OK。

---

## 5. 预期输出与验收标准

| 子系统 | 验收标准 | 负责人 |
|--------|---------|--------|
| admin-web | HTTP 200，标题含"颐贞 · 管理后台"，非空 HTML | 自动化 |
| family-app | HTTP 200，Flutter web启动无报错，有应用容器 | 自动化 |
| nurse_terminal | HTTP 200，登录界面可见，非空白页 | 自动化 |
| website | HTTP 200，Hugo站点渲染正常，有内容展示 | 自动化 |
| miniprogram | 微信开发者工具可成功导入项目，无报错 | 人工 |
| ui-prototypes | ≥8/12 个指定界面原型文件存在 | 人工检查 |
| cloud backend | Go编译通过，至少 api-server 可启动 | 自动化+人工 |

---

## 6. 补充内容清单（验证中发现）

根据初步检查发现的潜在欠缺：

1. **miniprogram**：缺少自动化启动脚本支持，需依赖 WeChat DevTools GUI
2. **ui-prototypes**：从 git diff 看有部分原型文件已被删除，需确认是否有意保留
3. **cloud backend**：依赖外部基础设施（PostgreSQL/Redis/EMQX），完整启动需要先搭建环境

这些问题将在验证过程中逐一发现并记录，逐个补齐后再进行回归验证。

---

## 7. 后续实施计划

本设计文档通过后，将调用 `writing-plans` 技能生成详细实施计划，包含：
- 精确到每一步的命令序列
- 异常情况的处理预案
- 验证结果的记录格式
- 欠缺内容的补充具体方案

设计者：Eregen 项目架构组
日期：2026-07-26
