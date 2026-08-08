/*
 * Eregen (颐贞) - Pro+ 试纸检测模块 · BLE 血压计设备连接头文件
 * Pro+ 通过 BLE 连接外接血压计模块，获取收缩压/舒张压/脉搏数据。
 *
 * 注：BLE 协议栈由 nRF5x SDK 或 GD32 内置 BLE 提供，此处仅定义
 * 应用层数据结构和接口，不与底层协议栈耦合。
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef PRO_PLUS_BP_DEVICE_H
#define PRO_PLUS_BP_DEVICE_H

#include <stdint.h>
#include <stdbool.h>
#include "FreeRTOS.h"

/* ----------------------------------------------------------------
 * BLE 服务 UUID（Eregen 血压计外设自定义服务）
 * ---------------------------------------------------------------- */
#define BP_SERVICE_UUID \
    {0x71,0x3E,0x00,0x11,0x00,0x00,0x10,0x00, \
     0x80,0x00,0x00,0x80,0x5F,0x9B,0x34,0xFB}

/* 特征 UUID：血压测量值 */
#define BP_MEAS_CHAR_UUID \
    {0x71,0x3E,0x00,0x21,0x00,0x00,0x10,0x00, \
     0x80,0x00,0x00,0x80,0x5F,0x9B,0x34,0xFB}

/* ----------------------------------------------------------------
 * 连接状态
 * ---------------------------------------------------------------- */
typedef enum {
    BP_STATE_DISCONNECTED = 0U,
    BP_STATE_SCANNING,
    BP_STATE_CONNECTING,
    BP_STATE_CONNECTED,
    BP_STATE_MEASURING,
    BP_STATE_ERROR,
} bp_device_state_t;

/* ----------------------------------------------------------------
 * 血压测量结果
 * ---------------------------------------------------------------- */
typedef struct {
    uint16_t systolic;      /* 收缩压 mmHg */
    uint16_t diastolic;     /* 舒张压 mmHg */
    uint16_t pulse_bpm;     /* 脉搏 次/分 */
    uint32_t timestamp_s;   /* 测量时间戳 */
    bool     valid;         /* 结果是否有效 */
    uint8_t  quality;       /* 0-100 */
} bp_measurement_t;

/* ----------------------------------------------------------------
 * 设备上下文
 * ---------------------------------------------------------------- */
typedef struct {
    bp_device_state_t state;
    uint8_t           peer_addr[6];   /* BLE 远端地址 */
    bool              bonded;         /* 是否已配对 */
    bp_measurement_t  last_result;    /* 最近一次测量结果 */
    TaskHandle_t      measure_task;   /* 测量任务句柄 */
} bp_device_t;

/* ----------------------------------------------------------------
 * 回调函数
 * ---------------------------------------------------------------- */
typedef void (*bp_result_cb_t)(const bp_measurement_t *m, void *ctx);

/* ----------------------------------------------------------------
 * 公共 API
 * ---------------------------------------------------------------- */

/**
 * 初始化 BLE 血压计设备模块。
 * @param dev 设备上下文。
 * @return true 表示初始化成功。
 */
bool bp_device_init(bp_device_t *dev);

/**
 * 开始扫描附近的血压计外设。
 * @param dev 设备上下文。
 */
void bp_device_scan_start(bp_device_t *dev);

/**
 * 停止扫描。
 * @param dev 设备上下文。
 */
void bp_device_scan_stop(bp_device_t *dev);

/**
 * 连接到指定 MAC 地址的血压计。
 * @param dev     设备上下文。
 * @param addr    6 字节 BLE 地址。
 * @param timeout_ms 连接超时。
 * @return true 表示连接成功。
 */
bool bp_device_connect(bp_device_t *dev,
                       const uint8_t addr[6],
                       uint32_t timeout_ms);

/**
 * 断开与血压计的连接。
 * @param dev 设备上下文。
 */
void bp_device_disconnect(bp_device_t *dev);

/**
 * 发起一次血压测量（触发外设开始测量）。
 * @param dev     设备上下文。
 * @param result  输出测量结果。
 * @param timeout_ms 等待超时。
 * @return true 表示获得有效结果。
 */
bool bp_device_take_measurement(bp_device_t *dev,
                                bp_measurement_t *result,
                                uint32_t timeout_ms);

/**
 * 注册测量结果回调。
 * @param dev  设备上下文。
 * @param cb   回调函数。
 * @param ctx  回调上下文。
 */
void bp_device_set_callback(bp_device_t *dev,
                            bp_result_cb_t cb,
                            void *ctx);

/**
 * 获取当前连接状态。
 * @param dev 设备上下文。
 * @return 当前状态枚举值。
 */
bp_device_state_t bp_device_get_state(const bp_device_t *dev);

#endif /* PRO_PLUS_BP_DEVICE_H */
