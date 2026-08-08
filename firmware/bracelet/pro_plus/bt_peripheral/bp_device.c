/*
 * Eregen (颐贞) - Pro+ 试纸检测模块 · BLE 血压计设备连接实现
 * 通过 BLE GATT 与外接血压计外设通信，获取收缩压/舒张压/脉搏数据。
 *
 * 本实现为应用层框架，具体 BLE 栈调用由目标平台 SDK 提供。
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "bp_device.h"

#ifdef TARGET_PRO_PLUS

#include <string.h>
#include "../common/log.h"

/* 扫描超时（秒） */
#define BP_SCAN_TIMEOUT_S     30U

/* 默认连接超时（ms） */
#define BP_DEFAULT_CONN_TIMEOUT  5000U

/* ----------------------------------------------------------------
 * 内部状态
 * ---------------------------------------------------------------- */
static bp_result_cb_t   s_bp_cb      = NULL;
static void            *s_bp_cb_ctx  = NULL;

/* ----------------------------------------------------------------
 * 公共 API 实现
 * ---------------------------------------------------------------- */

bool bp_device_init(bp_device_t *dev)
{
    if (!dev) return false;

    memset(dev, 0, sizeof(*dev));
    dev->state = BP_STATE_DISCONNECTED;
    dev->bonded = false;
    memset(dev->peer_addr, 0, sizeof(dev->peer_addr));
    memset(&dev->last_result, 0, sizeof(dev->last_result));

    log_info("BPDev: Initialized (state=DISCONNECTED)");
    return true;
}

void bp_device_scan_start(bp_device_t *dev)
{
    if (!dev) return;
    dev->state = BP_STATE_SCANNING;
    log_info("BPDev: Scan started");
    /* 实际实现：调用 BLE stack scan_start() API */
}

void bp_device_scan_stop(bp_device_t *dev)
{
    if (!dev) return;
    dev->state = BP_STATE_DISCONNECTED;
    log_info("BPDev: Scan stopped");
}

bool bp_device_connect(bp_device_t *dev,
                       const uint8_t addr[6],
                       uint32_t timeout_ms)
{
    if (!dev || !addr) return false;

    dev->state = BP_STATE_CONNECTING;
    memcpy(dev->peer_addr, addr, 6);

    log_info("BPDev: Connecting to addr=%02X:%02X:%02X:%02X:%02X:%02X",
             addr[0], addr[1], addr[2], addr[3], addr[4], addr[5]);

    /*
     * 实际实现：
     * 1. 调用 BLE stack connect(addr, timeout)
     * 2. 等待连接事件回调
     * 3. 发现 GATT 服务（BP_SERVICE_UUID）
     * 4. 注册 BP_MEAS_CHAR_UUID 特征通知
     */

    /* 模拟：假设连接成功（实际需等待 BLE 栈回调） */
    dev->state = BP_STATE_CONNECTED;
    dev->bonded = true;

    log_info("BPDev: Connected");
    return true;
}

void bp_device_disconnect(bp_device_t *dev)
{
    if (!dev) return;
    dev->state = BP_STATE_DISCONNECTED;
    dev->bonded = false;
    log_info("BPDev: Disconnected");
}

bool bp_device_take_measurement(bp_device_t *dev,
                                bp_measurement_t *result,
                                uint32_t timeout_ms)
{
    if (!dev || !result) return false;
    if (dev->state != BP_STATE_CONNECTED) {
        log_warn("BPDev: Not connected, cannot measure");
        return false;
    }

    dev->state = BP_STATE_MEASURING;
    log_info("BPDev: Taking measurement (timeout=%u ms)", (unsigned)timeout_ms);

    /*
     * 实际实现：
     * 1. 写入 BP_MEAS_CHAR_UUID 特征值（触发测量命令）
     * 2. 等待 GATT 通知回调（测量结果）
     * 3. 解析 PPD (Professional Personal Data) 格式
     */

    /* 模拟结果：占位血压值 */
    result->systolic   = 120U;
    result->diastolic  = 80U;
    result->pulse_bpm  = 72U;
    result->timestamp_s = (uint32_t)xTaskGetTickCount() * (1000U / configTICK_RATE_HZ);
    result->valid      = true;
    result->quality    = 90U;

    dev->last_result = *result;
    dev->state = BP_STATE_CONNECTED;

    log_info("BPDev: Measurement result: SYS=%u DIA=%u PR=%u",
             result->systolic, result->diastolic, result->pulse_bpm);

    /* 回调通知 */
    if (s_bp_cb) {
        s_bp_cb(result, s_bp_cb_ctx);
    }

    return result->valid;
}

void bp_device_set_callback(bp_device_t *dev,
                            bp_result_cb_t cb,
                            void *ctx)
{
    s_bp_cb = cb;
    s_bp_cb_ctx = ctx;
    (void)dev;
}

bp_device_state_t bp_device_get_state(const bp_device_t *dev)
{
    return dev ? dev->state : BP_STATE_DISCONNECTED;
}

#else /* !TARGET_PRO_PLUS */

/* 空实现，避免未定义引用 */
bool bp_device_init(bp_device_t *d)      { (void)d; return false; }
void bp_device_scan_start(bp_device_t *d) { (void)d; }
void bp_device_scan_stop(bp_device_t *d)  { (void)d; }
bool bp_device_connect(bp_device_t *d, const uint8_t*a, uint32_t t)
{ (void)d; (void)a; (void)t; return false; }
void bp_device_disconnect(bp_device_t *d) { (void)d; }
bool bp_device_take_measurement(bp_device_t *d, bp_measurement_t *r, uint32_t t)
{ (void)d; (void)r; (void)t; return false; }
void bp_device_set_callback(bp_device_t *d, bp_result_cb_t cb, void *ctx)
{ (void)d; (void)cb; (void)ctx; }
bp_device_state_t bp_device_get_state(const bp_device_t *d)
{ (void)d; return BP_STATE_DISCONNECTED; }

#endif /* TARGET_PRO_PLUS */
