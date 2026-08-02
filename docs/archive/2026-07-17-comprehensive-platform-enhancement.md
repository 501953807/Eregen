# Eregen 全平台功能完善实施计划

> **生成日期：** 2026-07-17
> **目标：** 将所有规划的任务全部执行，所有子系统功能全面完善，系统功能深化应用
> **范围：** 云平台后端、B2B服务、固件(手环/药盒)、前端(admin-web/family-app/小程序)

---

## 全局约束

- 开源许可：仅 MIT/BSD-3/Apache-2.0/ISC，禁用 GPL/AGPL/LGPL
- 技术选型：Go+Gin(云)、Flutter(家属APP)、Vue3+TS+Element Plus(管理后台)、微信小程序、FreeRTOS+C(手环)、ESP-IDF v5.3+C(药盒)
- 安全：JWT验证、Redis限流、PII脱敏、输入校验、bcrypt密码、TLS MinVersion:TLS12
- GitHub仓库仅存代码，不存文档/配置

---

## 批次划分

| 批次 | 内容 | 优先级 |
|------|------|--------|
| **Batch 1** | 云平台NATS主题修复 + Data-Pipeline启动 + Push-Service用户获取 | P0-阻塞性 |
| **Batch 2** | API-Server Service层数据库对接 + Handler死代码清理 | P0-核心 |
| **Batch 3** | Admin-API路由注册 + Dashboard完整实现 | P1 |
| **Batch 4** | B2B Hospital API健康数据真实处理 | P1 |
| **Batch 5** | 固件MQTT通信完整实现(Entry/Plus/Pro + Basic/Smart/Auto) | P1 |
| **Batch 6** | 前端页面API全面对接(admin-web/family-app/小程序) | P2 |
| **Batch 7** | 小程序修复(Alerts/Medication/Mine) | P2 |

---

## Batch 1: 云平台NATS主题修复 + Data-Pipeline启动 + Push-Service用户获取

### Task 1.1: 修复NATS主题不一致问题

**文件：**
- Modify: `cloud/gateway/internal/nats/publisher.go`
- Modify: `cloud/api-server/internal/service/nats_client.go`
- Modify: `cloud/data-pipeline/internal/subscriber/nats_handler.go`

**接口：**
- Consumes: 无
- Produces: 统一的 `eregen.event.{type}` 主题体系

**步骤：**

- [ ] **Step 1: 统一NATS主题规范**

阅读 `cloud/gateway/internal/nats/publisher.go:19` 确认当前发布主题为 `eregen.event.{type}`。

将 `cloud/api-server/internal/service/nats_client.go:17` 的订阅主题从 `device.events.` 改为 `eregen.event.`。

将 `cloud/data-pipeline/internal/subscriber/nats_handler.go:37` 的订阅主题从 `device.events` 改为 `eregen.event.`。

运行 `grep -rn "device.events\|eregen.event" cloud/ --include="*.go"` 验证所有引用已统一。

- [ ] **Step 2: 验证编译通过**

```bash
cd cloud/api-server && go build ./...
cd cloud/data-pipeline && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add cloud/gateway/internal/nats/publisher.go cloud/api-server/internal/service/nats_client.go cloud/data-pipeline/internal/subscriber/nats_handler.go
git commit -m "fix: unify NATS event subjects to eregen.event.{type} across all services"
```

### Task 1.2: 激活Data-Pipeline NATS订阅器

**文件：**
- Modify: `cloud/data-pipeline/internal/main.go`
- Modify: `cloud/data-pipeline/internal/subscriber/nats_handler.go`

**接口：**
- Consumes: `nats_handler.Handler` struct with `Start()` method
- Produces: 启动的NATS事件处理循环

**步骤：**

- [ ] **Step 1: 阅读 nats_handler.go 确认 Handler 结构**

读取 `cloud/data-pipeline/internal/subscriber/nats_handler.go` 确认 `Handler` 的结构体定义、`Start()` 方法签名、以及依赖的字段(config, store, healthAnalyzer, riskScoreCalc)。

- [ ] **Step 2: 修改 main.go 实例化并启动 Handler**

修改 `cloud/data-pipeline/internal/main.go`，将当前被丢弃的 `_ = HealthAnalyzer` 和 `_ = RiskScoreCalculator` 改为：

```go
handler := subscriber.NewNATSHandler(natsConn, pgStore, dataPipelineCfg.Analyzer, healthAnalyzer, riskScoreCalc)
log.Info("starting data pipeline NATS subscriber")
if err := handler.Start(ctx); err != nil {
    log.Fatal("failed to start NATS subscriber:", err)
}
```

确保在 `defer cancel()` 之前启动，在 `log.Info("shutting down")` 之后停止。

- [ ] **Step 3: 验证编译通过**

```bash
cd cloud/data-pipeline && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add cloud/data-pipeline/internal/main.go
git commit -m "feat: activate data-pipeline NATS subscriber for device event processing"
```

### Task 1.3: 实现Push-Service用户获取(DB查询)

**文件：**
- Create: `cloud/push-service/internal/store/postgres.go` (新建)
- Modify: `cloud/push-service/internal/publisher/nats_subscriber.go`

**接口：**
- Consumes: `router.FamilyMember` struct
- Produces: `func GetFamilyMembersByElderlyID(elderlyID string) ([]router.FamilyMember, error)`

**步骤：**

- [ ] **Step 1: 阅读现有 router.go 确认 FamilyMember 结构**

读取 `cloud/push-service/internal/router/router.go` 确认 `FamilyMember` 包含 `UserID`, `DeviceToken`, `OpenID`, `Phone` 字段。

- [ ] **Step 2: 创建 postgres.go Store**

创建 `cloud/push-service/internal/store/postgres.go`：

```go
package store

import (
    "context"
    "github.com/uptrace/bun"
    "eregen/push-service/internal/router"
)

type PostgresStore struct {
    db *bun.DB
}

func NewPostgresStore(db *bun.DB) *PostgresStore {
    return &PostgresStore{db: db}
}

func (s *PostgresStore) GetFamilyMembersByElderlyID(ctx context.Context, elderlyID string) ([]router.FamilyMember, error) {
    var members []router.FamilyMember
    err := s.db.NewSelect().
        Table("users").
        Column("id", "device_token", "openid", "phone").
        Where("parent_user_id = ?", elderlyID).
        Where("role = 'family'").
        Scan(ctx, &members)
    if err != nil {
        return nil, err
    }
    return members, nil
}
```

- [ ] **Step 3: 修改 nats_subscriber.go 使用真实DB查询**

修改 `cloud/push-service/internal/publisher/nats_subscriber.go:125-127`，将：

```go
members := []router.Member{{UserID: ev.ElderlyID}} // TODO: fetch from DB
```

替换为：

```go
members, err := s.store.GetFamilyMembersByElderlyID(ctx, ev.ElderlyID)
if err != nil {
    log.Warn("failed to fetch family members", zap.String("elderly_id", ev.ElderlyID), zap.Error(err))
    return // skip push if no members found
}
if len(members) == 0 {
    log.Warn("no family members found for alert", zap.String("elderly_id", ev.ElderlyID))
    return
}
```

同时需要在 `NewNATSSubscriber` 中接收 `*store.PostgresStore` 参数。

- [ ] **Step 4: 修改 main.go 传递 store**

读取 `cloud/push-service/internal/main.go`，在创建 `NewNATSSubscriber` 时传入 `postgresStore`。

- [ ] **Step 5: 验证编译通过**

```bash
cd cloud/push-service && go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add cloud/push-service/internal/store/postgres.go cloud/push-service/internal/publisher/nats_subscriber.go cloud/push-service/internal/main.go
git commit -m "feat: implement real DB-based family member fetching in push-service"
```

### Task 1.4: 启用Data-Pipeline地理围栏检查

**文件：**
- Modify: `cloud/data-pipeline/internal/subscriber/nats_handler.go`

**接口：**
- Consumes: `GeofenceStore` interface (需新建)
- Produces: 地理围栏越界告警发布到NATS

**步骤：**

- [ ] **Step 1: 创建 geofence_store.go**

创建 `cloud/data-pipeline/internal/store/geofence_store.go`：

```go
package store

import (
    "context"
    "eregen/data-pipeline/internal/model"
)

type GeofenceStore struct {
    db *bun.DB
}

func NewGeofenceStore(db *bun.DB) *GeofenceStore {
    return &GeofenceStore{db: db}
}

func (s *GeofenceStore) GetActiveGeofences(ctx context.Context, elderlyID string) ([]model.Geofence, error) {
    var fences []model.Geofence
    err := s.db.NewSelect().
        Model(&fences).
        Where("elderly_id = ?", elderlyID).
        Where("active = true").
        Scan(ctx)
    return fences, err
}
```

- [ ] **Step 2: 添加 Geofence model**

在 `cloud/data-pipeline/internal/model/model.go` 中添加：

```go
type Geofence struct {
    ID        uuid.UUID `bun:"id,pk,type:uuid"`
    ElderlyID string    `bun:"elderly_id,type:uuid"`
    Name      string    `bun:"name"`
    Latitude  float64   `bun:"latitude"`
    Longitude float64   `bun:"longitude"`
    Radius    float64   `bun:"radius"` // meters
    Active    bool      `bun:"active"`
    CreatedAt time.Time `bun:"created_at,autoNOW"`
}
```

- [ ] **Step 3: 实现 haversine 距离计算**

在 `cloud/data-pipeline/internal/analyzer/time_helper.go` 旁新建 `cloud/data-pipeline/internal/analyzer/geofence.go`：

```go
package analyzer

import "math"

const earthRadius = 6371000.0 // meters

// HaversineDistance returns distance in meters between two lat/lng points.
func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
    dLat := (lat2 - lat1) * math.Pi / 180.0
    dLng := (lng2 - lng1) * math.Pi / 180.0
    a := math.Sin(dLat/2)*math.Sin(dLat/2) +
        math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
            math.Sin(dLng/2)*math.Sin(dLng/2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return earthRadius * c
}

func IsInsideGeofence(lat, lng, fenceLat, fenceLng, radius float64) bool {
    return HaversineDistance(lat, lng, fenceLat, fenceLng) <= radius
}
```

- [ ] **Step 4: 取消注释并实现地理围栏检查逻辑**

修改 `cloud/data-pipeline/internal/subscriber/nats_handler.go:131-139`，将注释掉的代码替换为：

```go
func (h *Handler) checkGeofence(lat, lng float64, elderlyID string) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    fences, err := h.geofenceStore.GetActiveGeofences(ctx, elderlyID)
    if err != nil {
        log.Warn("failed to get geofences", zap.String("elderly_id", elderlyID), zap.Error(err))
        return
    }

    for _, f := range fences {
        if !IsInsideGeofence(lat, lng, f.Latitude, f.Longitude, f.Radius) {
            log.Info("geofence breach detected",
                zap.String("elderly_id", elderlyID),
                zap.Float64("lat", lat),
                zap.Float64("lng", lng),
                zap.String("fence", f.Name),
            )
            // Publish breach event to NATS for push-service consumption
            h.publishAlert(elderlyID, "geofence_breach", map[string]interface{}{
                "fence_name": f.Name,
                "latitude":   lat,
                "longitude":  lng,
                "distance":   HaversineDistance(lat, lng, f.Latitude, f.Longitude),
            })
        }
    }
}
```

- [ ] **Step 5: 在 processLocation 中调用 checkGeofence**

在 `processLocation` 函数中，处理完心跳后调用 `h.checkGeofence(payload.Lat, payload.Lon, elderlyID)`。

- [ ] **Step 6: 修改 main.go 传递 geofenceStore**

在 `cloud/data-pipeline/internal/main.go` 中创建 `NewGeofenceStore(pg)` 并传给 `NewNATSHandler`。

- [ ] **Step 7: 验证编译通过**

```bash
cd cloud/data-pipeline && go build ./...
```

- [ ] **Step 8: Commit**

```bash
git add cloud/data-pipeline/internal/store/geofence_store.go cloud/data-pipeline/internal/model/model.go cloud/data-pipeline/internal/analyzer/geofence.go cloud/data-pipeline/internal/subscriber/nats_handler.go cloud/data-pipeline/internal/main.go
git commit -m "feat: implement geofence breach detection in data-pipeline"
```

### Task 1.5: 实现Data-Pipeline告警发布到NATS

**文件：**
- Modify: `cloud/data-pipeline/internal/subscriber/nats_subscriber.go` (应为 nats_handler.go)

**接口：**
- Consumes: `*nats.Conn` (已在 Handler 中)
- Produces: 告警事件发布到 `eregen.event.alert` 主题

**步骤：**

- [ ] **Step 1: 实现 publishAlert 方法**

在 `cloud/data-pipeline/internal/subscriber/nats_handler.go` 中添加：

```go
func (h *Handler) publishAlert(elderlyID, alertType string, details map[string]interface{}) {
    alertMsg := map[string]interface{}{
        "type":       "alert",
        "elderly_id": elderlyID,
        "alert_type": alertType,
        "details":    details,
        "timestamp":  time.Now().Unix(),
    }
    data, _ := json.Marshal(alertMsg)
    _ = h.nats.Publish("eregen.event.alert", data)
}
```

- [ ] **Step 2: 在风险评分超过阈值时调用 publishAlert**

在 `processHealth` 或 `processRiskScore` 中，当风险评分 > P0/P1 阈值时，调用 `h.publishAlert(elderlyID, "high_risk_score", ...)`。

- [ ] **Step 3: 验证编译通过**

```bash
cd cloud/data-pipeline && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add cloud/data-pipeline/internal/subscriber/nats_handler.go
git commit -m "feat: publish alert events to NATS for push-service consumption"
```
