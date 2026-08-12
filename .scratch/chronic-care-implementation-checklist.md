# Eregen 慢病管理升级 - 实施清单

> 更新时间：2026-08-08  
> 项目：颐贞 (Eregen) 慢病管理功能全栈实现

---

## 已完成 (Completed)

### 后端服务
- [x] 后端数据库迁移（7张表）
- [x] 后端数据模型（chronic.go）
- [x] 后端API Handler（17条路由）
- [x] 后端路由注册
- [x] 分析引擎（血糖/尿酸/血压）

### 固件层
- [x] 固件Pro+模块（10个文件）

### 应用层
- [x] APP慢病主页
- [x] APP详情页（6个）

### 设计文档
- [x] UI原型（3个HTML页面）
- [x] 方案文档更新

---

## 待完成 (Pending)

### 合规与供应链
- [ ] 试纸ODM代工对接
- [ ] 二类医疗器械注册申请

### 硬件开发
- [ ] 血压配件硬件开发
- [ ] 固件BLE血压计集成

### 系统集成
- [x] 推送服务慢病提醒扩展
- [ ] 端到端硬件联调

### 用户验证
- [ ] 用户测试（5-10人）

---

## 验证清单 (Verification)

### 编译验证
- [x] API编译通过：`go build ./...`
- [x] Flutter分析通过：`flutter analyze`
- [x] 单元测试通过：`go test ./...`

### 功能验证
- [x] UI原型在浏览器中可访问
- [x] 后端API可通过curl测试

---

## 关键文件索引

### 后端服务
```
cloud/api-server/internal/
├── model/chronic*.go          # 慢病数据模型
├── store/chronic*.go          # 慢病存储层
├── service/chronic*.go        # 慢病业务逻辑
├── handler/chronic*.go        # 慢病API处理器
└── analyzer/                   # 分析引擎
```

### 固件层
```
firmware/bracelet/pro_plus/
├── src/
│   ├── chronic_main.c         # 主控制模块
│   ├── chronic_ble.c          # BLE协议模块
│   ├── chronic_glucose.c      # 血糖采集模块
│   ├── chronic_blood_pressure.c # 血压采集模块
│   ├── chronic_uric_acid.c    # 尿酸采集模块
│   ├── chronic_insulin.c      # 胰岛素管理模块
│   ├── chronic_alerts.c       # 告警模块
│   ├── chronic_sync.c         # 数据同步模块
│   ├── chronic_storage.c      # 本地存储模块
│   └── chronic_config.h       # 配置头文件
└── README.md                  # 模块说明文档
```

### 应用层
```
apps/family-app/lib/screens/chronic/
├── chronic_home_screen.dart     # 慢病主页
├── glucose_detail_screen.dart   # 血糖详情
├── blood_pressure_detail_screen.dart # 血压详情
├── uric_acid_detail_screen.dart # 尿酸详情
├── insulin_screen.dart          # 胰岛素管理
├── medication_screen.dart       # 用药提醒
└── chronic_stats_screen.dart    # 统计报表
```

### UI原型
```
apps/ui-prototypes/family-app/
├── chronic-home.html            # 慢病主页原型
├── chronic-detail-glucose.html  # 血糖详情原型
└── chronic-detail-blood-pressure.html # 血压详情原型
```

### 设计文档
```
docs/
├── superpowers/specs/
│   └── 2026-08-08-chronic-care-upgrade-design.md  # 设计方案文档
└── superpowers/plans/
    └── 2026-08-08-chronic-care-upgrade.md          # 实施计划
```

---

## 快速导航

### 查看后端实现
```bash
cd /Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/cloud/api-server
find . -name "chronic*.go" -type f
```

### 查看固件实现
```bash
cd /Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/firmware/bracelet/pro_plus
ls -la src/
```

### 查看APP实现
```bash
cd /Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/family-app
find lib/screens/chronic -name "*.dart" -type f
```

### 查看UI原型
```bash
open /Users/tangxiaochuan/AIWorkspace/ClaudeWorkspace/Eregen/apps/ui-prototypes/family-app/chronic-home.html
```

---

## 下一步行动

### 优先级 P0（本周）
1. 完成验证清单中的所有检查项
2. 修复任何编译或测试问题
3. 准备用户测试环境

### 优先级 P1（下周）
1. 启动用户测试（5-10人）
2. 收集反馈并迭代优化
3. 开始血压配件硬件开发

### 优先级 P2（本月）
1. 推进二类医疗器械注册申请
2. 对接试纸ODM代工
3. 完成端到端硬件联调

---

**文档维护说明：**
- 每完成一项待办事项，将对应的 `[ ]` 改为 `[x]`
- 如有新增任务，添加到对应分类中
- 完成全部实施后，将此文档归档至 `.scratch/` 目录
