/*
 * Eregen (颐贞) - Pro+ 试纸检测模块 · 慢病管理任务调度头文件
 * 负责轮询试纸插入状态、触发测量、连接血压计，
 * 并将结果送入 FreeRTOS 消息队列供通信任务上报云端。
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef PRO_PLUS_CHRONIC_MANAGER_H
#define PRO_PLUS_CHRONIC_MANAGER_H

#include <stdint.h>
#include <stdbool.h>
#include "FreeRTOS.h"
#include "strip_type.h"
#include "electrochemical.h"
#include "bp_device.h"

/* ----------------------------------------------------------------
 * 任务优先级与栈大小
 * ---------------------------------------------------------------- */
#define TASK_CHRONIC_PRIORITY  (tskIDLE_PRIORITY + 4)
#define TASK_CHRONIC_STACK     (configMINIMAL_STACK_SIZE * 6)

/* ----------------------------------------------------------------
 * 检测轮询间隔（ms）
 * ---------------------------------------------------------------- */
#define CHRONIC_POLL_INTERVAL_MS   2000U    /* 每 2 秒检查一次试纸插入 */
#define CHRONIC_BP_POLL_INTERVAL_S 60U      /* 每 60 秒尝试血压测量 */

/* ----------------------------------------------------------------
 * 消息队列大小
 * ---------------------------------------------------------------- */
#define QUEUE_CHRONIC_SIZE    8U

/* ----------------------------------------------------------------
 * 慢病检测结果消息（发给通信任务）
 * ---------------------------------------------------------------- */
typedef struct {
    bool      has_elec_result;           /* 是否有试纸检测结果 */
    bool      has_bp_result;             /* 是否有血压测量结果 */
    detection_result_t elec_result;      /* 试纸检测结果 */
    bp_measurement_t   bp_result;        /* 血压测量结果 */
    uint32_t  timestamp_s;               /* 消息时间戳 */
} chronic_report_t;

/* ----------------------------------------------------------------
 * 慢病管理上下文
 * ---------------------------------------------------------------- */
typedef struct {
    electrochemical_dev_t elec_dev;      /* 电化学检测器 */
    bp_device_t           bp_dev;        /* BLE 血压计 */
    QueueHandle_t         report_queue;  /* 结果队列 */
    TaskHandle_t          task_handle;   /* 任务句柄 */
    uint32_t              measure_count; /* 累计测量次数 */
    uint32_t              error_count;   /* 累计错误次数 */
} chronic_ctx_t;

/* ----------------------------------------------------------------
 * 公共 API
 * ---------------------------------------------------------------- */

/**
 * 初始化慢病管理模块（队列 + 设备）。
 * 必须在 FreeRTOS 调度器启动前调用。
 * @return true 表示初始化成功。
 */
bool chronic_mgr_init(void);

/**
 * 创建慢病管理 FreeRTOS 任务。
 * @return true 表示任务创建成功。
 */
bool chronic_mgr_create_task(void);

/**
 * 发送慢病报告到通信队列。
 * @param report     报告数据。
 * @param timeout_ms 阻塞超时。
 * @return true 表示发送成功。
 */
bool chronic_mgr_send_report(const chronic_report_t *report,
                             uint32_t timeout_ms);

/**
 * 手动触发一次试纸检测（由用户按键或 APP 远程触发）。
 * @param type  试纸类型（由检测器自动识别时传 STRIP_TYPE_NONE）。
 * @return true 表示检测成功。
 */
bool chronic_mgr_trigger_test(strip_type_t type);

/**
 * 手动触发一次血压测量。
 * @return true 表示测量成功。
 */
bool chronic_mgr_trigger_bp_measure(void);

/**
 * 获取慢病管理上下文（供其他模块访问设备状态）。
 * @return 上下文指针，未初始化时返回 NULL。
 */
chronic_ctx_t *chronic_mgr_get_ctx(void);

#endif /* PRO_PLUS_CHRONIC_MANAGER_H */
