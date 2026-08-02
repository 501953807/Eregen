# Eregen 数据库安全审查报告 - Phase 6

**审查日期：** 2026-07-28
**审查范围：** cloud/api-server/internal/store/*.go, cloud/admin-api/internal/store/*.go

---

## 一、SQL注入风险分析

### Critical级发现：UpdateDeviceConfig存在严重SQL注入漏洞

**文件：** `cloud/admin-api/internal/store/postgres.go` 第173-182行

```go
func (s *PostgresStore) UpdateDeviceConfig(ctx context.Context, deviceID string, config map[string]interface{}) error {
    settings := `{}`
    for k, v := range config {
        if v != nil {
            settings += fmt.Sprintf(`"%s":%v,`, k, v)  // ← SQL注入点！
        }
    }
    _, err := s.db.ExecContext(ctx, `UPDATE devices SET settings = settings || $1::jsonb WHERE device_id = $2`, settings, deviceID)
    return err
}
```

**风险描述：** config map中的key值直接通过fmt.Sprintf拼接进SQL字符串，未作任何转义。攻击者可通过注入恶意键名（如 `"x' UNION SELECT * FROM users--"`）执行任意SQL查询或数据篡改。

**修复建议：**
- 使用预编译的map转义函数，对JSON key进行严格白名单验证
- 或使用标准的json.Marshal转换整个配置对象，而非逐字段拼接
- 示例修复：
```go
data, _ := json.Marshal(config)
_, err := s.db.ExecContext(ctx, `UPDATE devices SET settings = settings || $1::jsonb WHERE device_id = $2`, data, deviceID)
```

**影响评级：** CRITICAL — 可导致任意设备配置被篡改，包括固件更新URL、传感器参数等关键设置

---

### High级发现：多处使用字符串格式化构建WHERE子句

**文件位置：**
- `cloud/api-server/internal/store/sqlite.go` 第97-104行 (ListUsers)
- `cloud/api-server/internal/store/sqlite.go` 第131-140行 (ListDevices)  
- `cloud/api-server/internal/store/postgres.go` 第259-267行 (ListDevices)
- `cloud/api-server/internal/store/postgres.go` 第465-473行 (AdminStatsAlertTrend) — `where %s`拼接
- `cloud/api-server/internal/store/postgres.go` 第1166-1184行 (AdminDeviceList)

虽然这些查询参数均使用`$n`或`?`占位符传递值，但WHERE子句本身（表名、列名）通过字符串拼接，若将来过滤条件增加用户输入字段（如按部门筛选设备），将造成SQL注入。当前代码中`devType`, `tier`, `status`等参数经过硬编码比较，尚属安全，但存在潜在隐患。

---

## 二、Schema设计与完整性问题

### 1. MedStatusRecord模型与协议不匹配

**协议定义：** `{"type":"med_status","dev_id":"PX-XXXX","compartment":3,"taken":true,"ts":xxx}`

**数据库实现：** `med_status_records`表包含：
```sql
(id, rule_id, elderly_id, taken_at, taken, missed_at)
```

**缺失字段：**
- `compartment`（药盒分格编号）— 药盒特有关键字段完全缺失
- `device_id`（上报设备的dev_id）— 无法追溯数据来源设备
- `ts`时间戳用`taken_at`代替，语义不清

**影响：** 无法追踪是哪个药盒的分格在什么时间发生了服药事件，违背了协议设计原则。

---

### 2. HealthRecord模型缺少完整健康指标

**协议预期：** 心率(HR)、血氧(SPO2)、步数(Steps)、睡眠(Sleep Hours)等

**Database表结构：**
```sql
CREATE TABLE health_records (
    id TEXT PRIMARY KEY,
    elderly_id TEXT,
    timestamp TIMESTAMP,
    hr INT,
    spo2 INT,
    steps BIGINT,
    sleep_hours REAL,
    bp_systolic INT,
    bp_diastolic INT
)
```

**观察：** 模型中HR/SPO2/Steps等为指针类型(*int)，允许NULL，符合实际。但在GetHealthSummary中使用MAX(timestamp)作为group by键，同时取AVG(hr), MIN(spo2)等聚合，可能导致跨不同时间点的错误关联——同一天的不同记录中MAX()可能与MIN()来自不同行。建议改用窗口函数或先排序再分组。

---

### 3. 设备配置settings字段类型不一致

- **SQLite store:** `settings TEXT DEFAULT '{}'`（存储为JSON文本）
- **Postgres api-server store:** `settings TEXT`（同样存储JSON文本，非PostgreSQL的JSONB）
- **Postgres admin-api store:** `settings JSONB`（使用原生JSONB类型）

**问题：** api-server的Postgres实现本应使用JSONB却使用了TEXT，导致无法使用PostgreSQL特有的JSON操作符（->, ->>, etc.）。admin-store正确使用JSONB，但两者不一致。

---

### 4. missing elderly_devices关联表

**发现：** 在admin-api的postgres.go中，GetDeviceByID方法引用了`elderly_devices`表：
```sql
LEFT JOIN elderly_devices ed WHERE ed.elderly_id = e.id
```

但在api-server/store/sqlite.go的migrate函数和postgres.go的schema中，**完全没有创建elderly_devices表**。该表本应关联老人档案和设备，是"绑定设备到老人"功能的关键外键。

**后果：** GetDeviceByID功能在postgres store中将失败（表不存在），且api-server完全缺少老人-设备关联关系支持，导致老人无法正确绑定设备。

---

### 5. 用药规则缺少关联约束

medication_rules表中通过elderly_id关联到用户，但缺少外键约束到elderly_profiles表。理论上elderly_id应引用elderly_profiles.id，若老人被删除而用药规则仍存在，会导致孤儿数据。

---

## 三、连接池与事务管理

### 1. Postgres连接池配置过小

**文件：** `cloud/api-server/cmd/main.go` 第35行
```go
pgConfig.MaxConns = 10
```

在高并发场景下（百级以上设备并发上报），10个连接可能成为瓶颈，导致请求阻塞或超时。建议根据CPU核心数调整：`MaxConns = 2 * numcores + 5`。

---

### 2. DeleteUser使用全量事务但未控制批量操作风险

**文件：** `cloud/api-server/internal/store/postgres.go` 第1089-1158行

函数DeleteUser遍历用户关联的所有老人记录，逐个DELETE geofences、alerts、health_records等表。**无任何批量限制或分批逻辑**：
- 若某个老人有上万条健康记录，单条DELETE可能锁表长时间
- 事务中逐表循环DELETE，rollback风险大，易产生中间状态

**建议：** 改用CASCADE删除（如果外键已添加约束），或对大数据量表执行分页删除并在日志中记录进度。

---

### 3. Redis Token Store缺乏重放攻击保护

**文件：** `cloud/api-server/internal/store/redis.go` 第101-123行

ValidateRefreshToken仅检查token是否存在于Redis，**不标记已消耗**。攻击者可窃取refresh token后重复使用直到过期，每次都能获取新的access token。

**修复策略：** 
- 使用后删除token（一次性消费）
- 或维护token黑名单+rotation机制（每次刷新生成新token并失效旧token）

---

## 四、健康检查缺失深度依赖检测

**文件：** `cloud/api-server/router/router.go` 第33-35行
```go
r.GET("/api/v1/health", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ok"}})
})
```

该端点仅返回静态响应，**未验证**：
- PostgreSQL是否可达
- Redis是否正常
- NATS连接是否存活
- SMTP/FCM推送服务是否可用

在生产环境中，负载均衡器或Kubernetes liveness probe会误判服务为"健康"，实则后端依赖已不可用。

---

## 五、B2B服务数据库凭证硬编码

**文件回顾（从之前审查已确认）：**
- `b2b/hospital-api/cmd/main.go` 第21行
- `b2b/insurance-integration/cmd/main.go` 第21行  
- `b2b/community-platform/cmd/main.go` 第21行

```go
dbDSN := os.Getenv("DB_DSN")
if dbDSN == "" {
    dbDSN = "postgres://postgres:password@localhost:5432/eregen_hospital_api" // ← 硬编码默认凭证！
}
```

默认密码`postgres/password`是业界已知最弱组合之一。若环境变量未被设置，服务将以最高权限运行。

---

## 六、审计日志设计缺陷

**文件：** `cloud/admin-api/internal/store/postgres.go` 审计相关方法

audit功能仅有CreateAuditLog等方法，但：
1. **audit_logs表中没有记录操作前/后的值差异**，仅记录"做了什么"，无法追溯变更前后的状态
2. **无IP地址记录**，无法定位操作来源
3. **无角色变更审计**（UpdateUserRecords虽然存在，但实际未验证调用者是否有权限）

---

## 七、总结与建议清单

| # | 问题 | 严重程度 | 文件 | 修复建议 |
|---|------|---------|------|----------|
| 1 | UpdateDeviceConfig拼接SQL key名 | CRITICAL | admin-api/store/postgres.go:173-178 | 用json.Marshal替代string拼接 |
| 2 | MedStatusRecord缺少compartment/device_id字段 | HIGH | model.go + migrate | 扩展schema兼容协议定义 |
| 3 | elderly_devices表缺失（引用但不存在） | HIGH | store/sqlite.go / postgres.go | 创建表并添加外键约束 |
| 4 | RefreshToken无一次性消耗 | MEDIUM | store/redis.go:100-117 | 消费后立即删除token |
| 5 | /health端点无依赖检查 | MEDIUM | router/router.go:33-35 | 添加DB/NATS/Redis探测 |
| 6 | B2B服务硬编码DB凭据 | CRITICAL | b2b/*/cmd/main.go:21 | 移除hardcoded，env缺失时panic退出 |
| 7 | Postgres MaxConns=过小 | MEDIUM | cmd/main.go:35 | 根据机器规格调至20-50 |
| 8 | DeleteUser无分页/批量控制 | MEDIUM | postgres.go:1089-1158 | 增加batch删除或cascade |
| 9 | SQLite ListUsers/Filters WHERE拼接 | LOW-MEDIUM | sqlite.go:97, 131 | 保持当前方式但避免引入动态列名 |
|10 | JWT密钥硬编码见外部config | CRITICAL | gateway/internal/config/config.go | 从环境变量读取（非db审查但相关） |

---

**数据库审查完成度：** ✅ store层主要文件已覆盖，schema对比已完成，安全漏洞识别明确。下一步可结合前端应用测试实际查询行为验证所有发现。