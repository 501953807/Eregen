# 家属APP 对齐原型设计 — 实施计划

> 目标：将 `apps/family-app` 所有页面样式、风格、配色、布局、操作方式严格对齐 `apps/ui-prototypes/family-app/` 中的HTML原型设计。

---

## 一、全局变更（跨页面共享）

### 1.1 品牌配色修正
| 当前值 | 原型值 | 影响文件 |
|--------|--------|---------|
| `primary: #E8734A` (暖橙) | `#4A90D9` (品牌蓝) | `common/theme.dart` |
| `headerGradient`: 橙色渐变 | 蓝色渐变 `#4A90D9 → #357ABD` | `common/theme.dart` |
| `bgScaffold: #FFF9F5` (暖白) | `#f5f6fa` (冷灰白) | `common/theme.dart` |
| SOS gradient: 红橙渐变 | 红色 `#FF6B6B → #EE5A24` | `common/theme.dart` |

### 1.2 底部导航栏重构
- **当前**：扁平排列，首页凸起仅通过margin负值实现，使用Material Icon
- **原型**：中间首页按钮圆形外溢-24px，蓝色背景白色图标，其余tab用emoji
- **修改文件**：`widgets/bottom_nav_bar.dart` — 改为emoji图标 + 首页凸起圆形样式

---

## 二、首页 (home_page.dart) — 完全重写

### 2.1 布局结构变更

| 原型区域 | 当前实现 | 需要改为 |
|---------|---------|---------|
| Header | 白色顶栏 + 横向滚动老人卡片 | 蓝色渐变header(4A90D9→357ABD)，老人选择器内嵌header中 |
| 地图 | Expanded全屏地图 | 固定200px高圆角地图卡片，带网格背景+电子围栏虚线环+弹跳定位图标 |
| 快捷状态 | 底部面板中的GridView 6项 | 地图下方4列独立卡片(心率/血氧/步数/电量)，每张含emoji+数值+标签+状态徽章 |
| SOS | 红色告警横幅(SOS事件通知) | 红色渐变SOS卡片，含脉冲动画按钮+长按3秒说明 |
| 告警列表 | 底部面板中显示计数 | 独立"最近告警"区域，3条告警卡片+查看全部入口 |
| 底部导航 | 扁平5tab | 首页凸起圆形导航 |

### 2.2 Widget级改造

**需重写的widget：**
1. `_buildTopBar()` → 改为渐变蓝色header，老人选择器内嵌
2. `MapSection` widget → 添加电子围栏虚线环、弹跳定位图标、位置信息浮层
3. `_statItems()` → 拆分为独立的4列QuickStatusCard网格（原型是grid: repeat(4, 1fr)）
4. `_buildSOSBanner()` → 改为红色渐变SOS卡片（非告警横幅），含脉冲动画
5. 新增 `_buildRecentAlerts()` → 独立告警列表区域
6. `_buildBottomNav()` → 改为原型样式的凸起导航

**需删除的组件：**
- `_showSOSBanner` 布尔变量及SOS Banner（原型无此告警横幅）
- `_cardExpanded` 可展开/收起逻辑（原型无此交互）
- `_healthTip()` 健康提示卡片（原型首页无此元素）
- `_batteryRow()` 电池行（移到快捷状态卡片中）
- `_quickActions()` 快速操作按钮（原型无）
- `_locationTooltip()` 和 `_mapControls()`（原型地图无这些控件）

### 2.3 数据绑定
- 保留现有的 `_fetchData()` API调用
- 快捷状态卡片值从API获取（当前硬编码在底部面板中，需提到header区域）

---

## 三、健康页 (health_page.dart) — 大幅调整

### 3.1 布局变更

| 原型区域 | 当前实现 | 需要改为 |
|---------|---------|---------|
| 风险评分 | 蓝紫渐变卡片+CircularProgressIndicator | 蓝色渐变卡片+SVG圆环(#4A90D9背景+#4ADE80进度条) |
| 时间范围 | pills切换 | 原型中是简单按钮组(#f0f0f5背景) |
| 心率 | CustomPainter折线图 | 迷你柱状图(12个bar，蓝色渐变) |
| 血氧 | 摘要卡片 | 迷你柱状图(7个bar，绿色渐变) |
| ~~血压~~ | 缺失 | 新增：收缩压/舒张压分开展示，橙色偏高徽章 |
| ~~睡眠~~ | 缺失 | 新增：深睡/浅睡/REM分段条形图 |
| 步数统计 | 柱状图+目标线 | 3项统计行(步数/千卡/活动分钟) |

### 3.2 需新增组件
1. 血压卡片 — 收缩压128/舒张压82分开展示
2. 睡眠质量卡片 — 深睡2.1h/浅睡3.8h/REM 1.3h彩色分段条
3. 迷你柱状图组件 — 替换现有CustomPainter折线图

### 3.3 需删除组件
1. `_buildHeartRateChart()` CustomPainter折线图
2. `_buildIntergenerationalComparison()` 代际对比（原型无此功能）
3. AI Insight Banner（原型无此元素）

---

## 四、告警页 (alerts_page.dart) — 结构调整

### 4.1 布局变更

| 原型区域 | 当前实现 | 需要改为 |
|---------|---------|---------|
| 统计卡片 | 白底彩色边框卡片 | 三色渐变背景卡片(红/橙/蓝) |
| 过滤 | 芯片式筛选 | Tab式过滤(全部/未处理/SOS/跌倒/健康) |
| SOS操作 | 独立红色卡片 | 嵌入每条告警的操作按钮区 |
| 告警卡片 | 图标+标题+操作按钮 | 类型徽章(SOS/跌倒/心率/围栏/用药)+优先级标签+操作按钮 |

### 4.2 需改动
1. `_buildStatsRow()` → 改为渐变背景统计卡片
2. `_buildFilterChips()` → 改为原型tab样式
3. `_buildAlertItem()` → 添加类型徽章+操作按钮(立即呼叫/查看位置/标记处理)
4. 删除独立的SOS Quick Action卡片，操作按钮合并到告警卡片内

---

## 五、用药页 (medication_page.dart) — 布局重构

### 5.1 布局变更

| 原型区域 | 当前实现 | 需要改为 |
|---------|---------|---------|
| 依从性 | 今日进度环形图 | 绿色渐变依从性卡片(85%白色进度环) |
| 统计行 | 无(合并到进度卡片) | 新增：已服用/漏服/迟到 3项统计 |
| 药品展示 | 列表+时段筛选 | 时间线布局(左侧时间+连接线圆点+右侧药盒卡片) |
| 远程配置 | 4个toggle | 保持toggle但样式改为原型样式 |
| ~~库存~~ | 有进度条 | 原型无此模块 |
| ~~日历~~ | 有依从性热力图 | 原型无此模块 |
| 添加用药 | 弹窗表单 | 原型中是"+ 添加"文字链接 |

### 5.2 需重写
1. `_buildTodaySummaryCard()` → 改为原型依从性环形图样式
2. 新增 `_buildMedStatsRow()` → 3项统计(已服用/漏服/迟到)
3. `_buildMedItem()` → 改为时间线布局(时间标签+连接线+药盒卡片)
4. `_buildInventoryCard()` → 删除
5. `_buildAdherenceHistory()` → 删除
6. `_buildRemoteConfig()` → 改为原型toggle样式
7. 新增添加用药按钮("+ 新增用药规则")

---

## 六、父母福利页 (welfare_page.dart)

该页无对应原型文件(`family-app-welfare-v2.html`存在但不在 `ui-prototypes/family-app/` 目录内)。当前实现相对完整，暂不纳入本次对齐范围，待后续补充原型后再评估。

---

## 七、实施顺序和依赖关系

```
Phase 1: 全局基础
  ├─ 1.1 修改 theme.dart (配色)
  └─ 1.2 修改 bottom_nav_bar.dart (导航样式)

Phase 2: 首页 (工作量最大)
  ├─ 2.1 重写 home_page.dart 布局
  ├─ 2.2 重写 map_section.dart (添加电子围栏/弹跳图标)
  ├─ 2.3 重写 sos_button.dart (脉冲动画+长按3秒)
  ├─ 2.4 重写 quick_status_card.dart (4列独立卡片)
  └─ 2.5 重写 recent_alerts_list.dart (告警列表)

Phase 3: 健康页
  ├─ 3.1 重写 health_page.dart
  ├─ 3.2 新增血压卡片
  ├─ 3.3 新增睡眠卡片
  └─ 3.4 新增迷你柱状图组件

Phase 4: 告警页
  ├─ 4.1 重写 alerts_page.dart 统计卡片
  ├─ 4.2 重写过滤tab
  └─ 4.3 重写告警卡片样式

Phase 5: 用药页
  ├─ 5.1 重写 medication_page.dart 时间线
  ├─ 5.2 重写依从性卡片
  └─ 5.3 删除库存和日历模块
```

---

## 八、验证标准

每个页面完成后，需在浏览器中打开对应的 `apps/ui-prototypes/family-app/*.html` 原型文件，与Flutter运行效果逐项对比：
1. 配色是否一致（hex值精确匹配）
2. 布局结构是否一致（卡片排列、间距、圆角）
3. 交互方式是否一致（点击/长按/滑动）
4. 字体大小和层级是否一致

---

## 九、注意事项

1. **保持现有API集成**：配色和布局改回原型，但 `_fetchData()` / WebSocket / 离线缓存等业务逻辑保持不变
2. **保留dark mode**：原型中无dark mode，但现有代码中有，作为增强功能保留
3. **Login页面不变**：login_page.dart 不在原型范围内，保持现状
4. **Settings页面不变**：settings_page.dart 不在原型范围内，保持现状
5. **所有mock数据替换为真实API数据**：对齐后逐步替换硬编码数据
