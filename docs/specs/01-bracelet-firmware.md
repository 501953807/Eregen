# ① 手环固件 — 详细设计文档

> 生成日期：2026-07-17  
> 对应子系统：① 手环固件 (GD32E230C8T3 + FreeRTOS)  
> 语言：C | RTOS：FreeRTOS | 通信：Cat1 MQTT over AT

---

## 1. 概述

### 1.1 职责

手环固件负责采集老人健康数据（心率、血氧、步数、体温、ECG）、定位数据（GPS/北斗）、告警触发（SOS按钮、跌倒检测），通过 Cat1 蜂窝网络经 MQTT 上报至云平台，同时接收云端下发的配置更新和 OTA 升级指令。

> **注意：** 住院场景下，患者腕带由独立子系统⑨（医用电子腕带，ESP32-S3 + BLE GATT）承担，手环固件不处理住院患者核验逻辑。

### 1.2 三档差异化

| 功能模块 | Entry (Starter) | Plus (中端) | Pro (高端) | Pro+ (医疗级) |
|---------|:---:|:---:|:---:|:---:|
| 心率/血氧 (PPG) | ✅ | ✅ | ✅ | ✅ |
| SOS 按钮 | ✅ | ✅ | ✅ | ✅ |
| GPS 定位 | ✅ | ✅ | ✅ (GNSS 高精度) | ✅ (GNSS 高精度) |
| IMU 跌倒检测 | ❌ | ✅ (专用算法) | ✅ | ✅ |
| 电子围栏 | ❌ | ✅ | ✅ | ✅ |
| BLE 配网 | ❌ | ✅ | ✅ | ✅ |
| 电池优化 | ❌ | ✅ (动态采样) | ✅ | ✅ |
| ECG 心电 | ❌ | ❌ | ✅ | ✅ |
| AMOLED 屏 | ❌ | ST7789 LCD | ✅ (高刷 AMOLED) | ✅ (高刷 AMOLED) |
| 金属机身 | ❌ | 塑料 | ✅ | ✅ |
| 电化学检测 (血糖/尿酸) | ❌ | ❌ | ❌ | ✅ (专用试纸模块) |
| BLE 血压计配件 | ❌ | ❌ | ❌ | ✅ |

### 1.3 输入输出

| 类型 | 数据源/目标 | 协议 |
|------|-----------|------|
| **输入** | PPG 传感器 (汇顶 GT320) | I2C |
| **输入** | IMU 传感器 (ICM-42670-P) | SPI |
| **输入** | GPS 模组 (和芯星通/u-blox) | UART NMEA |
| **输入** | ECG 传感器 (Pro/Pro+) | I2S |
| **输入** | 电化学检测模块 (Pro+) | SPI (试纸识别) |
| **输入** | BLE 血压计配件 (Pro+) | BLE GATT |
| **输入** | SOS 按钮 | GPIO |
| **输入** | OLED/AMOLED 显示屏 | SPI/I2C |
| **输出** | Cat1 模组 (广和通 L610-CM) | UART AT 指令 |
| **输出** | BLE 外设 (Plus/Pro/Pro+ 配网) | BLE 5.0 |

---

## 2. 核心数据结构

### 2.1 消息协议 (shared/protocol)

```c
// 所有上行消息统一格式
typedef struct {
    char type[32];      // "heartbeat" / "location" / "health" / "sos" / "fall"
    char dev_id[16];    // "BR-XXXX"
    uint64_t ts;        // Unix timestamp (秒)
} MessageHeader;

// 心跳包
typedef struct {
    MessageHeader header;
    int battery_pct;    // 0-100
} HeartbeatMessage;

// 定位数据
typedef struct {
    MessageHeader header;
    double lat;         // 纬度
    double lon;         // 经度
    float accuracy;     // 定位精度(米)
} LocationMessage;

// 健康数据
typedef struct {
    MessageHeader header;
    int hr;             // 心率 (bpm)
    int spo2;           // 血氧饱和度 (%)
    int64_t steps;      // 步数
    float temperature;  // 体温 (℃)
} HealthMessage;

// SOS 告警
typedef struct {
    MessageHeader header;
    double lat;
    double lon;
} SosMessage;

// 跌倒检测
typedef struct {
    MessageHeader header;
    float confidence;   // 跌倒置信度 0.0-1.0
    double lat;
    double lon;
} FallMessage;
```

### 2.2 设备状态机

```
┌─────────────────────────────────────────────────┐
│                  POWER_OFF                        │
│                  (深度休眠)                        │
└────────────────────┬────────────────────────────┘
                     │ 开机/唤醒
                     ▼
              ┌──────────────┐
              │  BOOT        │ ← 验证 OTA 镜像签名
              │  (Entry/Pro) │   失败则回滚到旧版本
              └──────┬───────┘
                     │ 启动成功
                     ▼
              ┌──────────────┐
              │  INIT        │ ← 传感器初始化
              │              │ ← GPS 冷启动/热启动
              │              │ ← Cat1 网络注册
              └──────┬───────┘
                     │
                     ▼
              ┌──────────────┐
              │  RUN         │ ← 主循环：采集→编码→发送
              │              │
              │  低功耗模式:  │ ← 空闲时进入 Stop 模式
              │  定时唤醒     │
              └──────┬───────┘
                     │
          ┌──────────┼──────────┐
          ▼          ▼          ▼
    ┌──────────┐ ┌────────┐ ┌──────────┐
    │ OTA_UPD  │ │ FALL   │ │ SOS_TRIG │
    │          │ │ DETECT │ │          │
    └──────────┘ └────────┘ └──────────┘
```

---

## 3. 功能模块说明

### 3.1 传感器采集层

| 文件 | 模块 | 说明 |
|------|------|------|
| `sensors_ppg.c/h` | PPG 驱动 | I2C 读取 GT320 心率/血氧原始数据 |
| `sensors_imu.c/h` | IMU 驱动 | SPI 读取 ICM-42670 加速度/陀螺仪 |
| `health/health_collector.c/h` | 健康采集器 | 融合 PPG+IMU 数据，计算心率/血氧/步数 |
| `gps_nmea.c/h` | GPS 解析 | NMEA GGA/RMC 语句解析为经纬度 |
| `ecg_driver.c/h` (Pro/Pro+) | ECG 驱动 | 单导联心电信号 ADC 采集 |
| `electrochem.c/h` (Pro+) | 电化学检测驱动 | SPI 试纸识别，葡萄糖/尿酸电化学信号采集 |
| `ble_blood_pressure.c/h` (Pro+) | BLE 血压计客户端 | 连接外接 BLE 血压计配件，同步血压数据 |

### 3.2 算法层

| 文件 | 模块 | 说明 |
|------|------|------|
| `algorithms/sliding_window.c/h` | 滑动窗口 | 通用滑动窗口统计，用于步数累计和信号平滑 |
| `algorithms/fall_detect.c/h` (Entry) | 基础跌倒检测 | 基于加速度阈值的简单判断 |
| `plus/fall_detect.c/h` (Plus) | 增强跌倒检测 | 加速度+角速度+方向综合判断 |
| `battery_adc.c/h` | 电池电量 | ADC 读取锂电池电压，折算百分比 |
| `power/power_mgmt.c/h` | 电源管理 | Stop 模式切换、定时器唤醒策略 |
| `plus/battery_optimizer.c/h` (Plus) | 电池优化器 | 动态调整采样间隔，延长续航 |

### 3.3 通信层

| 文件 | 模块 | 说明 |
|------|------|------|
| `cat1_at.c/h` | Cat1 AT 指令 | UART 与广和通 L610 通信，AT+MQTT 发布/订阅 |
| `protocol/message_encode.c/h` | 消息编码 | C 结构体 → JSON 字符串 |
| `protocol/message_decode.c/h` | 消息解码 | JSON 字符串 → C 结构体 (下行指令) |
| `protocol/heartbeat.c/h` | 心跳管理 | 定时生成心跳包并上报 |
| `ble_pair.c/h` (Plus/Pro) | BLE 配网 | BLE GATT 服务，手机 APP 辅助 WiFi/Cat1 配网 |

### 3.4 告警层

| 文件 | 模块 | 说明 |
|------|------|------|
| `sos_button.c/h` | SOS 按钮 | GPIO 中断检测，防抖处理，长按 3s 触发 |
| `location/geofence.c/h` (Entry) | 电子围栏 (Entry 版) | 简易距离判断 |
| `plus/geofence_manager.c/h` (Plus) | 电子围栏管理器 | 多边形围栏，进出触发告警 |
| `gps_manager.c/h` | GPS 管理器 | 定位周期控制、坐标缓存、精度评估 |

### 3.5 显示层

| 文件 | 模块 | 说明 |
|------|------|------|
| `display_st7789.c/h` (Entry) | LCD 驱动 | SPI 驱动 ST7789 1.8寸 LCD |
| `display_amoled.c/h` (Pro) | AMOLED 驱动 | SPI 驱动 AMOLED 高刷屏 |
| `board_init.c/h` (Entry) | 板级初始化 | Entry 版本 GPIO/SPI/I2C 引脚配置 |
| `board_pro.c/h` (Pro) | Pro 板级初始化 | Pro 版本额外引脚 (ECG/GNSS) 配置 |

### 3.6 OTA 升级

| 文件 | 模块 | 说明 |
|------|------|------|
| `ota/ota_download.c/h` | OTA 下载 | HTTP 下载差分升级包到 Flash 空闲区 |
| `ota/ota_verify.c/h` | OTA 校验 | SHA-256 签名验证 + CRC32 完整性检查 |
| `ota/boot_switch.c/h` | 启动切换 | 修改 Boot 标志位，下次重启进入新版本 |

### 3.7 公共组件

| 文件 | 模块 | 说明 |
|------|------|------|
| `common/crc16.c/h` | CRC16 | Modbus CRC16 校验 |
| `common/ring_buffer.c/h` | 环形缓冲区 | UART/MQTT 数据收发缓冲 |
| `common/log.c/h` | 日志系统 | 分级日志 (DEBUG/INFO/WARN/ERROR) |

---

## 4. FreeRTOS 任务划分

### Entry 版本任务

| 任务名 | 优先级 | Stack | 职责 |
|--------|--------|-------|------|
| `MainTask` | 最高 | 512 字 | 系统初始化、任务创建 |
| `SensorTask` | 高 | 384 字 | PPG/IMU 数据采集，周期 1s |
| `GPSTask` | 中 | 384 字 | GPS 定位获取，周期 30s |
| `HealthTask` | 中 | 512 字 | 健康数据处理，周期 5s |
| `CommTask` | 高 | 768 字 | MQTT 连接维持、消息发送/接收 |
| `SOSTask` | 最高 | 256 字 | SOS 按钮检测，GPIO 中断 |
| `PowerTask` | 低 | 256 字 | 电池电量监测，周期 60s |
| `DisplayTask` | 低 | 384 字 | 屏幕刷新，显示状态信息 |

### Plus/Pro/Pro+ 额外任务

| 任务名 | 优先级 | Stack | 职责 |
|--------|--------|-------|------|
| `FallDetectTask` (Plus) | 高 | 512 字 | IMU 数据跌倒分析 |
| `GeofenceTask` (Plus) | 中 | 384 字 | 电子围栏判断 |
| `BLETask` (Plus/Pro/Pro+) | 中 | 512 字 | BLE 配网服务 |
| `ECGTask` (Pro/Pro+) | 高 | 512 字 | ECG 心电采集 |
| `AMOLEDTask` (Pro/Pro+) | 低 | 512 字 | AMOLED 高刷显示 |
| `ElectrochemTask` (Pro+) | 高 | 768 字 | 电化学检测采集，试纸识别+信号处理 |
| `BLEPressureTask` (Pro+) | 中 | 384 字 | BLE 血压计配件连接与数据同步 |

---

## 5. 接口定义

### 5.1 上行消息 (设备 → 云平台)

```json
// 心跳包
{"type":"heartbeat","dev_id":"BR-0001","bat":85,"ts":1720000000}

// 定位数据
{"type":"location","dev_id":"BR-0001","lat":31.2304,"lon":121.4737,"acc":5,"ts":1720000000}

// 健康数据
{"type":"health","dev_id":"BR-0001","hr":72,"spo2":98,"step":3456,"temp":36.5,"ts":1720000000}

// SOS 告警
{"type":"sos","dev_id":"BR-0001","lat":31.2304,"lon":121.4737,"ts":1720000000}

// 跌倒检测
{"type":"fall","dev_id":"BR-0001","conf":0.95,"lat":31.2304,"lon":121.4737,"ts":1720000000}

// 心电图数据 (Pro/Pro+)
{"type":"ecg","dev_id":"BR-0001","sample_rate":250,"duration":30,"ts":1720000000}

// 电化学检测 (Pro+) — 血糖/尿酸单次测量
{"type":"chronic_test","dev_id":"BR-0001","marker":"glucose","value":5.4,"unit":"mmol/L","mode":"fasting","ts":1720000000}

// 电化学检测 — 尿酸
{"type":"chronic_test","dev_id":"BR-0001","marker":"uric_acid","value":320,"unit":"umol/L","mode":"random","ts":1720000000}

// BLE 血压计数据 (Pro+)
{"type":"blood_pressure","dev_id":"BR-0001","sbp":120,"dbp":80,"pulse":72,"ts":1720000000}
```

### 5.2 下行指令 (云平台 → 设备)

```json
// 配置更新
{"type":"config","dev_id":"BR-0001","settings":{"interval":30,"volume":80}}

// OTA 升级
{"type":"ota","dev_id":"BR-0001","url":"https://...","hash":"sha256:..."}

// 电化学检测启动指令 (Pro+)
{"type":"chronic_test","dev_id":"BR-0001","action":"start","marker":"glucose","mode":"fasting"}

// BLE 血压计配对指令 (Pro+)
{"type":"config","dev_id":"BR-0001","settings":{"bp_device_addr":"AA:BB:CC:DD:EE:FF","bp_interval":300}}
```

---

## 6. 编译与烧录

### 6.1 环境要求

```bash
# macOS
brew install arm-none-eabi-gcc openocd

# Linux (Ubuntu/Debian)
sudo apt install arm-none-eabi-gcc openocd dfu-util
```

### 6.2 编译 Entry 版本

```bash
cd firmware/bracelet/entry
mkdir -p build && cd build
cmake .. -DCMAKE_TOOLCHAIN_FILE=../toolchain/arm-none-eabi.cmake
make -j$(nproc)
# 输出: build/bracelet_entry.bin
```

### 6.3 编译 Plus 版本

```bash
cd ../../plus && mkdir -p build && cd build
cmake .. && make
# 输出: build/bracelet_plus.bin
```

### 6.4 编译 Pro 版本

```bash
cd ../../pro && mkdir -p build && cd build
cmake .. && make
# 输出: build/bracelet_pro.bin
```

### 6.5 烧录 (J-Link OB)

```bash
JLinkExe -device GD32E230C8T3 -if SWD -speed 4000
# J-Link 交互界面:
loadbin bracelet_entry.bin, 0x08000000
reset
exit
```

---

## 7. 测试策略

### 7.1 单元测试

每个模块配有独立测试文件 (`test_*.c`)，在 x86 主机上编译运行验证逻辑正确性：

```bash
# 编译测试
cd firmware/bracelet/entry/common
gcc -DTEST_MODE test_crc16.c crc16.c -o test_crc16
./test_crc16

# 跌倒检测算法测试
gcc -DTEST_MODE ../algorithms/test_fall_detect.c algorithms/fall_detect.c algorithms/sliding_window.c -lm -o test_fall
./test_fall
```

### 7.2 集成测试

- 传感器数据环测：连接真实 PPG/IMU/GPS 模组，验证数据链路
- MQTT 端到端：设备 → EMQX → gateway → NATS → api-server，全链路验证
- OTA 升级测试：下发升级包，验证下载→校验→切换→回滚流程

---

## 8. 功耗设计

| 模式 | 电流 | 唤醒源 | 适用场景 |
|------|------|--------|---------|
| 运行模式 | ~25mA | — | 数据采集/发送中 |
| 空闲等待 | ~2mA | 定时器/GPIO | 两次采集间隔 |
| Stop 模式 | ~20μA | RTC 闹钟/外部中断 | 夜间/低频率采集 |
| 深度休眠 | ~1μA | SOS 按钮 | 长时间不活动 |

**续航估算：** 350mAh 电池，Entry 版本平均电流 8mA → 约 43 小时；Plus 版本电池优化器可延长至 72 小时。

---

## 9. Pro+ 变体详细设计

### 9.1 硬件架构扩展

Pro+ 在 Pro 基础上增加两大医疗检测能力，硬件架构扩展如下：

```
GD32E230C8T3
  ├── [ECG 模块] (与 Pro 共享)
  │     └── I2S → ADS1292R 单导联心电 ADC
  ├── [电化学检测模块] (Pro+ 新增)
  │     ├── SPI 接口 → 试纸识别芯片 (ADS1298 多通道同步)
  │     ├── GPIO → 试纸插入检测 (interrupt)
  │     └── I2C → 电化学传感器信号调理 (ADA4522 零漂移仪表放大器)
  ├── [BLE 血压计客户端] (Pro+ 新增)
  │     └── BLE 5.0 → 外接袖带式血压计配件 (GATT Service: 0x1810/0x2A35)
  └── [AMOLED 屏] (与 Pro 共享)
        └── SPI → 驱动 IC ST7796 / Samsung AMOLED
```

### 9.2 试纸检测模块规格

| 参数 | 规格 |
|------|------|
| 检测指标 | 血糖 (Glucose)、尿酸 (Uric Acid) |
| 试纸类型 | 一次性酶电极试纸 (葡萄糖氧化酶 / 尿酸酶法) |
| 采样原理 | 电化学安培法，恒电位 -0.5V vs Ag/AgCl |
| 量程 | 血糖：1.1–33.3 mmol/L；尿酸：60–1200 μmol/L |
| 采样周期 | 单次测量约 8 秒（含滴血识别→信号稳定→结果输出） |
| 接口协议 | SPI 主模式，8MHz，试纸 ID 预识别防误用 |
| 功耗 | 测量峰值 15mA，待机 <1μA |
| 试纸库存检测 | SPI 读取试纸仓霍尔传感器，低库存告警 |

**试纸仓硬件：**
- 内置霍尔传感器检测试纸剩余量
- 插入式检测机制：试纸推入自动唤醒电化学模块
- 防重复使用：单次测量后试纸仓机械闭锁

### 9.3 固件新增模块结构

```
firmware/bracelet/proplus/
├── electrochem/
│   ├── electrochem_driver.c/h    # 电化学模块 SPI/I2C 驱动
│   ├── strip_detect.c/h          # 试纸插入检测 + 类型识别
│   ├── glucose_measure.c/h       # 血糖测量算法（信号滤波+浓度换算）
│   └── uric_acid_measure.c/h     # 尿酸测量算法
├── ble_pressure/
│   ├── ble_pressure_client.c/h   # BLE GATT 血压计客户端
│   └── pressure_sync.c/h         # 血压数据同步+本地缓存
├── proplus_init.c/h              # Pro+ 特有板级初始化（GPIO/SPI/I2C 复用）
├── chron_task.c/h                # 慢病检测主流程任务（电化学+血压）
└── CMakeLists.txt
```

### 9.4 编译 Pro+ 版本

```bash
cd firmware/bracelet/proplus
mkdir -p build && cd build
cmake .. -DCMAKE_TOOLCHAIN_FILE=../toolchain/arm-none-eabi.cmake
make -j$(nproc)
# 输出: build/bracelet_proplus.bin
```

### 9.5 功耗影响说明

| 模式 | Pro+ 功耗影响 |
|------|-------------|
| 电化学待机 | +0.5μA（霍尔传感器持续监测） |
| 单次血糖测量 | +120mJ（8秒，峰值15mA） |
| BLE 血压计连接 | +8mA（扫描+连接周期维持） |
| 典型日常 | 续航影响 < 2%（电化学模块低频使用） |

> **注意：** 电化学检测为医疗级功能，测量结果仅供参考，不作为诊断依据。固件中应包含测量有效性校验，异常值需标记并通知云端 AI 分析引擎复核。

---

© 2026 Eregen (颐贞). All rights reserved.
