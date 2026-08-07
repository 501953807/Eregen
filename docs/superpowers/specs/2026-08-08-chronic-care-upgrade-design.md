# Eregen 慢性病专项升级 — 设计方案

> 编制日期：2026-08-08
> 状态：设计方案（已确认）
> 对应子系统：① 手环Pro+版硬件升级 + ④ 家属APP慢病管理模块
> 关联文档：`docs/specs/project_total_construction_scheme_v2.md` (v2.2), `docs/specs/04-family-app.md` (v2.0)

---

## 一、设计背景与目标

### 1.1 背景

Eregen 项目在第一阶段完成了手环（Entry/Plus/Pro）、药盒、云平台、家属APP等基础架构建设。为提升产品差异化竞争力，满足老年人慢性病管理的真实需求，需要在 Pro 版手环上增加血糖/尿酸检测能力，并升级家属APP为"银发健康管家"平台。

### 1.2 设计目标

1. **手环Pro+版**：新增可拆卸电化学检测模块，支持血糖/尿酸试纸检测
2. **血压配件**：外置蓝牙血压计配件，数据同步到手环和APP
3. **家属APP升级**：新增慢病管理专项模块，借鉴糖护士APP和第乐健康APP的设计模式
4. **后端能力**：新增慢病数据模型、API接口和AI分析引擎
5. **商业模式**：试纸耗材订阅制，形成长期收入来源

### 1.3 设计原则

- **技术可行性优先**：不追求无创血糖检测（技术不成熟），采用成熟的试纸电化学检测方案
- **专利保护导向**：自研试纸电极设计，不与市售试纸兼容，避免专利风险
- **适老化设计**：所有界面遵循适老化原则，大字体、高对比度、简单操作
- **生态闭环**：检测→数据→分析→建议→执行→反馈的完整闭环

---

## 二、硬件升级方案：手环Pro+版

### 2.1 硬件架构

```
┌─────────────────────────────────────────────────────────────┐
│  GD32E230C8T3 (ARM Cortex-M4, FreeRTOS)                     │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐          │
│  │ PPG     │ │ IMU     │ │ GPS     │ │ ECG     │          │
│  │ 心率/血氧│ │ 跌倒检测 │ │ 定位    │ │ 心电    │          │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘          │
│       │           │           │           │                │
│  ┌────┴───────────┴───────────┴───────────┴──────────┐    │
│  │              🆕 可拆卸检测模块插槽                   │    │
│  │  · 电化学传感器阵列（金/碳三电极）                  │    │
│  │  · 微流控通道（毛细进样）                           │    │
│  │  · 信号放大IC + 温度补偿NTC                         │    │
│  │  · 试纸类型识别（阻抗谱分析）                       │    │
│  │  · 自动排废机构（可选）                             │    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐   │
│  │  🆕 蓝牙 5.3 血压计配件接口                         │   │
│  │  · BLE Central角色连接外置血压计                    │   │
│  │  · 测量完成后自动上传数据                           │   │
│  └────────────────────────────────────────────────────┘   │
│  Cat1通信 → EMQX MQTT → NATS总线                         │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 试纸检测模块技术规格

| 参数 | 规格 | 说明 |
|------|------|------|
| 检测项目 | 血糖 + 尿酸（一机双测） | 不同试纸，自动识别 |
| 酶法 | FAD-GDH（脱氢酶法） | 抗干扰能力强，适合用降糖药的老人 |
| 血糖范围 | 1.1 ~ 33.3 mmol/L | 覆盖低血糖到高血糖全范围 |
| 尿酸范围 | 60 ~ 900 μmol/L | 覆盖临床全范围 |
| 测试时间 | ≤ 10秒 | 与糖护士SPUG一致 |
| 采血量 | 0.3 ~ 0.6 μL | 微量采血，适合老人 |
| 电极类型 | 三电极（金工作电极 + 碳对电极 + Ag/AgCl参比） | 行业标准方案 |
| 试纸尺寸 | 自定义设计，申请专利 | 不与市售试纸兼容 |
| 温度补偿 | 内置NTC，-10~50°C补偿 | 环境温度适应 |
| 试纸识别 | 插入时检测电极阻抗特征，自动识别血糖/尿酸 | 防止误用 |
| 检测功耗 | 峰值50mA，待机<1μA | Cat1电池可持续使用 |
| 医疗注册 | 二类医疗器械（与试纸一起注册） | 走NMPA注册流程 |

### 2.3 血压配件方案

| 参数 | 规格 |
|------|------|
| 配件类型 | 外置上臂式蓝牙血压计（可拆卸，独立定价） |
| 连接方式 | 蓝牙 5.3（与手环直连） |
| 测量原理 | 示波法（医疗级标准，精度高于脉搏波法） |
| 测量范围 | 压力 0~280 mmHg，脉搏 40~180 次/分 |
| 精度 | 压力 ±3 mmHg，脉搏 ±5% |
| 数据同步 | 测量完成后自动上传 → 手环 → 云端 → 家属APP |
| 提醒逻辑 | 手环定时提醒"该测血压了"，测量后APP记录 |

### 2.4 固件新增模块

```
firmware/bracelet/
├── pro_plus/
│   ├── main.c                    # 主循环（更新：增加检测模块调度）
│   ├── sensors/
│   │   ├── ppg.c/h               # 心率/血氧（现有）
│   │   ├── imu.c/h               # IMU跌倒检测（现有）
│   │   ├── gps.c/h               # GPS定位（现有）
│   │   ├── ecg.c/h               # ECG心电（现有）
│   │   └── electrochemical.c/h   # 🆕 电化学检测模块驱动
│   ├── bt_peripheral/
│   │   ├── bp_device.c/h         # 🆕 蓝牙血压计配件连接
│   │   └── test_strip_module.c/h # 🆕 检测模块通信协议
│   ├── app/
│   │   ├── chronic_manager.c/h   # 🆕 慢病管理任务调度
│   │   └── test_strip_workflow.c/h  # 🆕 试纸检测流程控制
│   └── protocol/
│       ├── strip_type.h          # 🆕 试纸类型定义
│       └── electrochemical_msg.h # 🆕 检测数据消息格式
```

### 2.5 电化学检测驱动核心流程

```c
// 检测主流程
detection_result_t electrochemical_detect(void) {
    // 1. 检测试纸插入（GPIO中断触发）
    if (!strip_inserted()) return (detection_result_t){0};

    // 2. 识别试纸类型（阻抗谱分析）
    strip_type_t type = identify_strip_type();

    // 3. 启动微流控（毛细进样）
    start_microfluidic();

    // 4. 读取温度
    float temp = read_temperature();

    // 5. 电化学测量（恒电位仪读取nA级电流）
    float raw_signal = read_current_signal();

    // 6. 温度补偿 + 浓度换算
    float compensated = apply_temp_compensation(raw_signal, temp);
    float value = current_to_concentration(compensated, type);

    // 7. 信号质量验证
    if (signal_quality(raw_signal) < THRESHOLD) {
        return (detection_result_t){.type = STRIP_ERROR};
    }

    // 8. 清理微流控通道
    cleanup_channel();

    // 9. 上传数据到云端
    upload_detection_result(type, value, temp);

    return (detection_result_t){.type = type, .value = value, ...};
}
```

### 2.6 MQTT消息格式扩展

```json
// 试纸检测数据上报
{
  "type": "chronic_test",
  "dev_id": "BR-XXXX",
  "test_type": "glucose",
  "value": 6.8,
  "unit": "mmol/L",
  "test_mode": "fasting",
  "temperature": 25.3,
  "quality": 0.92,
  "ts": 1723000000
}

// 血压数据上报
{
  "type": "blood_pressure",
  "dev_id": "BR-XXXX",
  "systolic": 135,
  "diastolic": 85,
  "pulse": 72,
  "ts": 1723000000
}
```

### 2.7 试纸供应链与商业模式

**供应链架构：**
```
ODM代工生产 → Eregen自有品牌分装 → 订阅配送 → 老人使用
  ├── 电极基材：PET薄膜 + 金/碳电极印刷
  ├── 酶制剂：FAD-GDH（脱氢酶）
  ├── 微流控层：毛细通道设计
  ├── 试纸切割：自定义尺寸（申请专利）
  └── 包装：铝箔袋 + 干燥剂，有效期12个月
```

**定价策略：**
| 产品 | 定价 | 说明 |
|------|------|------|
| Pro+版手环 | 比Pro版加价300-500元 | 含检测模块 |
| 血糖试纸 × 50条 | 150元/盒 | ~3元/条 |
| 尿酸试纸 × 50条 | 200元/盒 | ~4元/条 |
| 组合装（各25条） | 180元/盒 | 推荐入门装 |
| 血压配件 | 300-500元 | 外置上臂式蓝牙血压计 |

**订阅模式：**
- 月度订阅：每月初自动配送试纸
- 智能估算：根据检测频率自动调整配送量
- 积分抵扣：健康积分可兑换试纸折扣

---

## 三、家属APP改造方案

### 3.1 导航架构升级

```
现有底部导航（4 Tab）→ 升级后（5 Tab）：

┌──────────────────────────────────────────────────────────┐
│  🏠首页  📊健康  ⚠️告警  💊用药  🩺慢病管理  ⚙️设置       │
│  (现有)                                  (新增核心差异化)  │
└──────────────────────────────────────────────────────────┘
```

### 3.2 新增页面列表

| 页面 | 路由 | 核心功能 | 设计参考 |
|------|------|---------|---------|
| 慢病管理主页 | `/chronic` | 血糖/尿酸趋势卡片 + 每日任务 + AI建议 | 糖护士数据管理 + 第乐健康任务体系 |
| 血糖详情页 | `/chronic/blood-sugar` | 趋势图 + 检测记录 + 异常标记 + 导出报告 | 糖护士多维度数据管理 |
| 尿酸详情页 | `/chronic/uric-acid` | 趋势图 + 检测记录 + 饮食建议联动 | 糖护士尿酸管理 |
| 血压详情页 | `/chronic/blood-pressure` | 收缩压/舒张压双折线图 + 热力图 | 糖护士血压监测 |
| 饮食记录页 | `/chronic/diet` | 食物数据库 + 碳水计算 + AI建议 | 糖护士饮食运动指导 |
| 运动追踪页 | `/chronic/exercise` | 手环数据联动 + 运动计划 + 消耗统计 | 第乐健康个性化运动 |
| 健康报告页 | `/chronic/report` | 周报/月报/年报 + AI综合建议 | 第乐健康报告体系 |

### 3.3 改造页面

| 页面 | 改造内容 |
|------|---------|
| 首页 | 新增血糖/尿酸/血压快速入口卡片 |
| 健康看板 | 扩充趋势图：增加血糖/尿酸/血压折线图 |
| 用药管理 | 与慢病任务体系打通，服药提醒+检测提醒联动 |

### 3.4 慢病管理主页设计

```
┌─────────────────────────────────────────┐
│  👤 [老人切换器]    📅 2026-08-08 周六    │
├─────────────────────────────────────────┤
│  🩸 血糖    📊 尿酸    🩺 血压          │
│  6.8 ↑      380       135/85            │
│  空腹       正常      偏高              │
│  [查看趋势]  [查看趋势] [查看趋势]       │
├─────────────────────────────────────────┤
│  📋 每日任务                            │
│  ☐ 测空腹血糖   07:00                   │
│  ☐ 测餐后血糖   12:00                   │
│  ☐ 测血压       08:00                   │
│  ☐ 服用降压药   08:00                   │
│  ☐ 散步30分钟   18:00                   │
│  [完成 2/5] → 进度条                    │
├─────────────────────────────────────────┤
│  🤖 AI健康建议                          │
│  "今日血糖略偏高，建议午餐减少精制碳水，  │
│   增加蔬菜和蛋白质摄入"                  │
│  [查看详情]                             │
├─────────────────────────────────────────┤
│  📷 快速入口                            │
│  [饮食记录] [运动记录] [健康报告]        │
└─────────────────────────────────────────┘
```

### 3.5 UI设计语言

**颐贞品牌色系统：**

| 用途 | 颜色 | 色值 |
|------|------|------|
| 主色 | 颐贞绿 | #2E7D32 |
| 辅助色 | 暖橙 | #FF8F00 |
| 成功/正常 | 绿色 | #4CAF50 |
| 警告/偏高 | 橙色 | #FF9800 |
| 危险/异常 | 红色 | #F44336 |
| 背景 | 米白 | #FAFAF5 |
| 文字 | 深灰 | #37474F |

**适老化设计原则：**
- 字体 ≥ 16sp（主要文字），≥ 14sp（辅助文字）
- 触摸目标 ≥ 48dp
- 高对比度（WCAG AA级）
- 关键操作大按钮
- 减少输入，增加选择
- 异常值用颜色+图标双重提示

**借鉴糖护士+第乐健康的设计要素：**
- 糖护士：数据卡片式布局、多维度数据管理、AI建议卡片
- 第乐健康：任务体系+奖励徽章、周期报告、个性化计划

### 3.6 任务体系与奖励机制

**任务类型：**
```
├── 检测任务：测血糖/尿酸/血压（定时提醒）
├── 用药任务：按时服药
├── 运动任务：每日步数/运动时长目标
├── 饮食任务：记录三餐
└── 健康任务：测量前等待10分钟等准备动作
```

**奖励机制：**
```
├── 完成每日任务 → 获得"健康积分"
├── 连续7天达标 → 解锁"自律达人"徽章
├── 连续30天达标 → 解锁"健康卫士"徽章
├── 月度达标 → 生成优秀报告 + 积分奖励
└── 积分可兑换试纸折扣（5%~20%）
```

### 3.7 AI建议引擎规则

| 场景 | 条件 | 建议类型 | 建议内容示例 |
|------|------|---------|-------------|
| 空腹血糖偏高 | > 7.0 mmol/L | 饮食建议 | "建议减少早餐精制碳水，增加蛋白质" |
| 餐后血糖偏高 | 2hPG > 10.0 | 运动建议 | "建议餐后30分钟散步20分钟" |
| 血糖偏低 | < 3.9 mmol/L | 紧急提醒 | "立即补充15g快速碳水，30分钟后复测" |
| 尿酸偏高 | > 420 μmol/L | 饮食建议 | "建议减少海鲜、啤酒摄入，多喝水" |
| 血压偏高 | 收缩压 > 140 | 就医建议 | "建议休息10分钟后复测，如仍高请咨询医生" |
| 连续达标7天 | 各项指标正常 | 鼓励 | "本周各项指标达标，继续保持！" |

**推送时机：**
- 检测数据异常 → 立即推送（P1级别）
- 任务提醒 → 定时推送（P2级别）
- 周期报告 → 每周日/每月1日自动生成

---

## 四、后端升级方案

### 4.1 新增数据模型

```sql
-- 血糖检测记录
CREATE TABLE chronic_glucose_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    elderly_id UUID NOT NULL REFERENCES elders(id),
    value DECIMAL(5,2) NOT NULL,
    unit VARCHAR(10) DEFAULT 'mmol/L',
    test_mode VARCHAR(20),  -- fasting/postprandial/random
    measurement_time TIMESTAMP NOT NULL,
    detected_at TIMESTAMP DEFAULT NOW(),
    source VARCHAR(20) DEFAULT 'test_strip',  -- test_strip/bt_device/imported
    quality DECIMAL(3,2),  -- 信号质量 0~1
    temperature DECIMAL(4,1)  -- 检测时温度
);

-- 尿酸检测记录
CREATE TABLE chronic_uric_acid_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    elderly_id UUID NOT NULL REFERENCES elders(id),
    value DECIMAL(6,1) NOT NULL,
    unit VARCHAR(10) DEFAULT 'μmol/L',
    measurement_time TIMESTAMP NOT NULL,
    detected_at TIMESTAMP DEFAULT NOW(),
    source VARCHAR(20) DEFAULT 'test_strip'
);

-- 血压记录
CREATE TABLE chronic_bp_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    elderly_id UUID NOT NULL REFERENCES elders(id),
    systolic INT NOT NULL,
    diastolic INT NOT NULL,
    pulse INT,
    measurement_time TIMESTAMP NOT NULL,
    detected_at TIMESTAMP DEFAULT NOW()
);

-- 饮食记录
CREATE TABLE chronic_diet_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    elderly_id UUID NOT NULL REFERENCES elders(id),
    meal_type VARCHAR(20) NOT NULL,  -- breakfast/lunch/dinner/snack
    food_items JSONB NOT NULL,  -- [{"name":"米饭","portion_g":150,"carbs_g":39}]
    total_carbs DECIMAL(6,1),
    total_calories DECIMAL(6,1),
    recorded_at TIMESTAMP DEFAULT NOW()
);

-- 运动记录
CREATE TABLE chronic_exercise_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    elderly_id UUID NOT NULL REFERENCES elders(id),
    type VARCHAR(50) NOT NULL,
    duration_min INT,
    calories DECIMAL(6,1),
    avg_hr INT,
    max_hr INT,
    recorded_at TIMESTAMP DEFAULT NOW()
);

-- 每日任务
CREATE TABLE chronic_daily_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    elderly_id UUID NOT NULL REFERENCES elders(id),
    task_type VARCHAR(30) NOT NULL,  -- bg_test/bp_test/medication/exercise/diet
    scheduled_time TIME NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP,
    task_date DATE DEFAULT CURRENT_DATE,
    UNIQUE(elderly_id, task_type, task_date)
);

-- 周期报告
CREATE TABLE chronic_health_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    elderly_id UUID NOT NULL REFERENCES elders(id),
    report_type VARCHAR(20) NOT NULL,  -- weekly/monthly/annual
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    data_summary JSONB,  -- 汇总统计数据
    ai_recommendations JSONB,  -- AI建议列表
    generated_at TIMESTAMP DEFAULT NOW()
);
```

### 4.2 新增API端点

```
慢性健康管理API（/api/v1/chronic/*）

血糖：
  POST   /api/v1/chronic/glucose           录入血糖值
  GET    /api/v1/chronic/glucose           获取血糖趋势列表（支持时间范围）
  GET    /api/v1/chronic/glucose/trend     获取趋势聚合数据（用于图表）
  POST   /api/v1/chronic/test-strip/read   试纸检测数据上报（从手环）

尿酸：
  POST   /api/v1/chronic/uric-acid         录入尿酸值
  GET    /api/v1/chronic/uric-acid         获取尿酸趋势列表

血压：
  POST   /api/v1/chronic/blood-pressure    录入血压值
  GET    /api/v1/chronic/blood-pressure    获取血压趋势列表
  POST   /api/v1/chronic/bp-device/sync    血压计数据同步（从蓝牙配件）

饮食：
  POST   /api/v1/chronic/diet              记录饮食
  GET    /api/v1/chronic/diet              获取饮食记录
  POST   /api/v1/chronic/diet/analyze      饮食分析（AI建议）

运动：
  POST   /api/v1/chronic/exercise          记录运动
  GET    /api/v1/chronic/exercise          获取运动记录

任务：
  GET    /api/v1/chronic/daily-tasks       获取当日任务列表
  PUT    /api/v1/chronic/daily-tasks/:id   标记任务完成

报告：
  GET    /api/v1/chronic/report/:type      获取周期报告（weekly/monthly/annual）
  POST   /api/v1/chronic/report/generate   手动生成报告

AI建议：
  GET    /api/v1/chronic/recommendations   获取综合AI建议
  POST   /api/v1/chronic/recommendations/feedback  反馈建议效果
```

### 4.3 后端目录结构

```
cloud/
├── api-server/
│   └── handler/
│       └── chronic/                    # 🆕 慢病管理handler
│           ├── glucose_handler.go
│           ├── uric_acid_handler.go
│           ├── bp_handler.go
│           ├── diet_handler.go
│           ├── exercise_handler.go
│           ├── task_handler.go
│           └── report_handler.go
├── admin-api/
│   └── handler/
│       └── chronic/                    # 🆕 管理后台慢病管理
│           ├── stats_handler.go
│           └── report_handler.go
├── data-pipeline/
│   └── service/
│       └── chronic/                    # 🆕 慢病分析引擎
│           ├── glucose_analyzer.go
│           ├── uric_acid_analyzer.go
│           ├── bp_analyzer.go
│           ├── diet_analyzer.go
│           ├── exercise_analyzer.go
│           └── recommendations.go
└── push-service/
    └── handler/
        └── chronic/                    # 🆕 慢病提醒推送
            ├── task_reminder.go
            └── abnormal_alert.go
```

---

## 五、实施计划

### 5.1 阶段划分

```
第四批（新增）：慢性病专项升级
├── Phase 4.1（月1-3）：手环Pro+固件开发 + 试纸模块原型
├── Phase 4.2（月2-4）：家属APP慢病管理模块开发
├── Phase 4.3（月3-5）：后端API + AI分析引擎 + 数据模型
├── Phase 4.4（月4-6）：试纸供应链对接 + 二类医疗器械注册申请
├── Phase 4.5（月5-6）：血压配件开发 + 集成测试
└── Phase 4.6（月6-8）：端到端联调 + UI原型确认 + 用户测试
```

### 5.2 详细任务分解

**Phase 4.1：手环Pro+固件开发**
- [ ] 电化学传感器选型与采购（AD5933阻抗分析仪 + 三电极试纸）
- [ ] 微流控通道设计（毛细进样结构）
- [ ] 温度补偿算法实现
- [ ] 试纸阻抗识别算法
- [ ] BLE 5.3血压计配件连接协议
- [ ] 固件编译测试（Pro+变体）

**Phase 4.2：家属APP慢病管理模块**
- [ ] UI原型设计（7个新页面 + 3个改造页面）
- [ ] 慢病管理主页开发
- [ ] 血糖/尿酸/血压详情页开发
- [ ] 饮食记录页开发（含食物数据库）
- [ ] 运动追踪页开发
- [ ] 健康报告页开发
- [ ] 任务体系与奖励机制
- [ ] AI建议引擎UI集成

**Phase 4.3：后端API + AI分析引擎**
- [ ] 数据库迁移脚本（7张新表）
- [ ] 慢病管理API（15个端点）
- [ ] 血糖分析器（趋势计算 + 异常检测）
- [ ] 尿酸分析器
- [ ] 血压分析器
- [ ] 饮食分析器（碳水计算 + 结构分析）
- [ ] 综合建议引擎
- [ ] 推送服务扩展（慢病提醒）

**Phase 4.4：试纸供应链 + 医疗器械注册**
- [ ] 试纸ODM代工对接
- [ ] 试纸电极专利设计
- [ ] 二类医疗器械注册申请
- [ ] 试纸包装设计
- [ ] 订阅配送系统开发

**Phase 4.5：血压配件开发**
- [ ] 外置血压计选型/定制
- [ ] BLE连接协议开发
- [ ] 数据同步测试
- [ ] 与手环Pro+集成测试

**Phase 4.6：端到端联调 + 用户测试**
- [ ] 试纸检测端到端测试（试纸→手环→云端→APP）
- [ ] 血压配件端到端测试
- [ ] UI原型确认（7个新页面）
- [ ] 用户测试（5-10个目标用户）
- [ ] 性能优化

### 5.3 成本估算

| 项目 | 成本 | 说明 |
|------|------|------|
| 试纸模块硬件BOM | ~150元/套 | 电化学传感器+信号IC+微流控+温度传感器 |
| 试纸生产成本 | ~2元/条 | ODM代工量产后 |
| 二类医疗器械注册 | ~30-50万元 | 与试纸一起注册 |
| 试纸供应链（初期） | ~20万元 | 试纸生产模具 + 首批库存 |
| 血压配件开发 | ~50万元 | 外置蓝牙血压计，独立注册 |
| APP开发（慢病模块） | ~3人月 | Flutter开发 + UI设计 |
| 后端开发（慢病API） | ~2人月 | Go后端 + AI分析引擎 |
| **合计** | **~120-150万元** | 含硬件+软件+注册 |

### 5.4 专利申请方向

| 专利方向 | 技术方案 | 预期保护范围 |
|---------|---------|-------------|
| 手环式试纸检测模块 | 可拆卸电化学检测插槽 + 微流控 + 温度补偿 | 模块结构 + 检测流程 |
| 试纸阻抗识别算法 | 插入时通过电极阻抗特征自动识别试纸类型 | 识别算法 |
| 慢病数据与手环联动 | 检测数据与运动/饮食/用药数据的关联分析 | 数据关联方法 |
| 试纸耗材订阅系统 | 基于检测频率的智能估算 + 自动配送 | 商业模式 + 系统架构 |

---

## 六、与现有系统的整合

### 6.1 与现有手环固件的整合

- Entry/Plus版本**不受影响**，保持现有功能
- Pro版**升级**为Pro+版，新增检测模块
- 条件编译隔离：`#ifdef TARGET_PRO_PLUS` 控制新模块代码
- 现有API兼容：检测数据通过新topic上报，不影响现有数据流

### 6.2 与现有家属APP的整合

- 现有4个Tab**保留**，新增第5个Tab"慢病管理"
- 首页新增快捷入口卡片（不破坏现有布局）
- 健康看板新增趋势图（在现有图表下方追加）
- 用药管理与慢病任务体系联动

### 6.3 与现有云平台的整合

- 新增API路径 `/api/v1/chronic/*`，不影响现有路径
- 新增数据表，不影响现有表结构
- 新增MQTT topic `eregen/chronic/#`，不影响现有topic
- 新增分析引擎模块，不影响现有规则引擎

---

## 七、风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 电化学传感器研发周期长 | 硬件延期 | 先采用现成传感器模块验证算法，再设计专用模块 |
| 二类医疗器械注册耗时 | 上市延迟 | 提前准备注册资料，与CRO合作加速 |
| 试纸供应链不稳定 | 交付风险 | 选择2-3家ODM供应商，建立备选方案 |
| 试纸检测精度不达标 | 用户体验差 | 与糖护士SPUG技术对标，确保精度一致 |
| APP改造影响现有功能 | 回归风险 | 渐进式开发，每新增功能单独测试 |
| 用户接受度不确定 | 市场风险 | 先做UI原型用户测试，验证需求真实性 |

---

## 八、成功标准

| 维度 | 指标 | 目标值 |
|------|------|-------|
| 硬件 | 试纸检测精度 | 与糖护士SPUG一致（±0.5mmol/L血糖，±15%尿酸） |
| 硬件 | 测试时间 | ≤ 10秒 |
| 软件 | 慢病管理模块用户留存率 | > 60%（30天） |
| 软件 | 任务完成率 | > 70% |
| 商业 | 试纸订阅转化率 | > 30%（购买Pro+版的用户） |
| 商业 | 试纸复购率 | > 50%（季度） |
| 注册 | 二类医疗器械注册 | 拿到注册证 |
| 专利 | 专利申请 | 至少2项发明/实用新型专利 |

---

## 九、附录

### 9.1 糖护士产品参考

- SPUG血糖尿酸测试仪：一机双测，FAD脱氢酶技术，38×28×7mm，无电池
- InsulinK胰岛素笔环：蓝牙连接，记录注射时间和剂量
- M100多参数监测仪：血糖+血压+心电图+血氧，机构用
- 糖护士APP：多品牌血糖仪对接，IDSS引擎，饮食运动指导
- 第乐健康APP：专属控糖计划，任务+奖励体系，周期报告

### 9.2 试纸技术参考

- 电极类型：三电极系统（工作电极/对电极/参比电极）
- 酶法：FAD-GDH（脱氢酶）vs GOD（葡萄糖氧化酶）
- 采血方式：指尖血（毛细血管全血）
- 测试时间：10秒级
- 温度影响：需补偿，工作温度10-40°C

### 9.3 相关文档

- `docs/specs/01-bracelet-firmware.md` — 手环固件原有设计
- `docs/specs/04-family-app.md` — 家属APP原有设计
- `docs/specs/2026-08-07-system-improvement-outline.md` — 系统完善大纲
