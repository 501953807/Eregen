/*
 * Eregen (颐贞) - Pro+ 试纸检测模块 · 慢病管理任务调度实现
 * 轮询试纸插入状态 → 触发检测 → 打包结果 → 发往通信队列。
 * 同时管理 BLE 血压计的周期性测量。
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "chronic_manager.h"

#ifdef TARGET_PRO_PLUS

#include <string.h>
#include "../common/log.h"
#include "../entry/free_rtos_tasks.h"

/* ----------------------------------------------------------------
 * 静态上下文
 * ---------------------------------------------------------------- */
static chronic_ctx_t s_ctx;
static bool          s_initialized = false;

/* ----------------------------------------------------------------
 * 试纸检测结果回调（由 elec_run_measurement 调用）
 * ---------------------------------------------------------------- */
static void s_elec_result_callback(const detection_result_t *result, void *ctx)
{
    (void)ctx;
    if (!result) return;

    log_info("ChronicMgr: Electrochemical result received, type=%d, valid=%d",
             (int)result->strip_type, (int)result->valid);
}

/* ----------------------------------------------------------------
 * 血压测量结果回调
 * ---------------------------------------------------------------- */
static void s_bp_result_callback(const bp_measurement_t *m, void *ctx)
{
    (void)ctx;
    if (!m) return;

    log_info("ChronicMgr: BP measurement: SYS=%u DIA=%u PR=%u",
             m->systolic, m->diastolic, m->pulse_bpm);
}

/* ----------------------------------------------------------------
 * 慢病管理 FreeRTOS 任务主函数
 * ---------------------------------------------------------------- */
static void vChronicManagerTask(void *pvParameters)
{
    (void)pvParameters;

    electrochemical_dev_t *elec = &s_ctx.elec_dev;
    bp_device_t           *bp   = &s_ctx.bp_dev;

    /* 初始化电化学模块 */
    if (!elec_detect_init(elec)) {
        log_error("ChronicMgr: Electrochemical init failed");
        /* 继续运行，血压功能仍可用 */
    }
    elec_set_result_callback(elec, s_elec_result_callback, &s_ctx);

    /* 初始化 BLE 血压计模块 */
    if (!bp_device_init(bp)) {
        log_error("ChronicMgr: BP device init failed");
    } else {
        bp_device_set_callback(bp, s_bp_result_callback, &s_ctx);
    }

    log_info("ChronicMgr: Task started");

    uint32_t  bp_tick_s = 0;
    uint32_t  poll_tick = 0;
    bool      strip_was_inserted = false;

    for (;;) {
        uint32_t now_ms = xTaskGetTickCount() * (1000U / configTICK_RATE_HZ);

        /* 轮询试纸插入状态 */
        if (++poll_tick >= (CHRONIC_POLL_INTERVAL_MS /
                            (1000U / configTICK_RATE_HZ))) {
            poll_tick = 0;

            bool inserted = elec_detect_strip_inserted(elec);

            if (inserted && !strip_was_inserted) {
                /* 试纸刚插入，自动检测类型并开始测量 */
                strip_type_t type = elec_detect_strip_type(elec);
                log_info("ChronicMgr: Strip inserted, type=%d", (int)type);

                detection_result_t result;
                if (elec_run_measurement(elec, type, &result,
                                         STRIP_MAX_MEASURE_MS)) {
                    s_ctx.measure_count++;

                    /* 打包并发送报告 */
                    chronic_report_t report;
                    memset(&report, 0, sizeof(report));
                    report.has_elec_result = true;
                    report.elec_result = result;
                    report.timestamp_s = (uint32_t)xTaskGetTickCount()
                                         * (1000U / configTICK_RATE_HZ);

                    if (chronic_mgr_send_report(&report, pdMS_TO_TICKS(500))) {
                        log_info("ChronicMgr: Report sent (measures=%u)",
                                 (unsigned)s_ctx.measure_count);
                    } else {
                        s_ctx.error_count++;
                        log_warn("ChronicMgr: Report send failed");
                    }
                } else {
                    s_ctx.error_count++;
                    log_error("ChronicMgr: Measurement failed");
                }
            }

            strip_was_inserted = inserted;
        }

        /* 定期血压测量 */
        if (++bp_tick_s >= CHRONIC_BP_POLL_INTERVAL_S) {
            bp_tick_s = 0;

            if (bp_device_get_state(bp) == BP_STATE_CONNECTED) {
                bp_measurement_t bp_result;
                if (bp_device_take_measurement(bp, &bp_result,
                                               pdMS_TO_TICKS(30000))) {
                    chronic_report_t report;
                    memset(&report, 0, sizeof(report));
                    report.has_bp_result = true;
                    report.bp_result = bp_result;
                    report.timestamp_s = (uint32_t)xTaskGetTickCount()
                                         * (1000U / configTICK_RATE_HZ);

                    chronic_mgr_send_report(&report, pdMS_TO_TICKS(500));
                }
            } else if (bp_device_get_state(bp) != BP_STATE_SCANNING) {
                /* 未连接时尝试重新连接（周期性地） */
                log_debug("ChronicMgr: BP not connected, will retry next cycle");
            }
        }

        vTaskDelay(pdMS_TO_TICKS(100));
    }
}

/* ----------------------------------------------------------------
 * 公共 API 实现
 * ---------------------------------------------------------------- */

bool chronic_mgr_init(void)
{
    if (s_initialized) return true;

    memset(&s_ctx, 0, sizeof(s_ctx));
    s_ctx.report_queue = xQueueCreate(QUEUE_CHRONIC_SIZE,
                                      sizeof(chronic_report_t));
    if (!s_ctx.report_queue) {
        log_error("ChronicMgr: Failed to create report queue");
        return false;
    }

    s_initialized = true;
    log_info("ChronicMgr: Module initialized");
    return true;
}

bool chronic_mgr_create_task(void)
{
    if (!s_initialized) {
        if (!chronic_mgr_init()) return false;
    }

    BaseType_t ret = xTaskCreate(vChronicManagerTask,
                                  "ChronicMgr",
                                  TASK_CHRONIC_STACK,
                                  NULL,
                                  TASK_CHRONIC_PRIORITY,
                                  &s_ctx.task_handle);
    if (ret != pdPASS) {
        log_error("ChronicMgr: Failed to create task");
        return false;
    }

    log_info("ChronicMgr: Task created (priority=%d)", (int)TASK_CHRONIC_PRIORITY);
    return true;
}

bool chronic_mgr_send_report(const chronic_report_t *report,
                             uint32_t timeout_ms)
{
    if (!report || !s_ctx.report_queue) return false;

    BaseType_t ret = xQueueSend(s_ctx.report_queue, report,
                                pdMS_TO_TICKS(timeout_ms));
    return (ret == pdPASS);
}

bool chronic_mgr_trigger_test(strip_type_t type)
{
    if (!s_initialized) return false;

    detection_result_t result;
    bool ok = elec_run_measurement(&s_ctx.elec_dev, type, &result,
                                    STRIP_MAX_MEASURE_MS);
    if (ok) {
        s_ctx.measure_count++;
        chronic_report_t report;
        memset(&report, 0, sizeof(report));
        report.has_elec_result = true;
        report.elec_result = result;
        report.timestamp_s = (uint32_t)xTaskGetTickCount()
                             * (1000U / configTICK_RATE_HZ);
        chronic_mgr_send_report(&report, pdMS_TO_TICKS(500));
    }
    return ok;
}

bool chronic_mgr_trigger_bp_measure(void)
{
    if (!s_initialized) return false;
    bp_measurement_t result;
    return bp_device_take_measurement(&s_ctx.bp_dev, &result,
                                       pdMS_TO_TICKS(30000));
}

chronic_ctx_t *chronic_mgr_get_ctx(void)
{
    return s_initialized ? &s_ctx : NULL;
}

#else /* !TARGET_PRO_PLUS */

/* 空实现 */
bool chronic_mgr_init(void)                          { return false; }
bool chronic_mgr_create_task(void)                   { return false; }
bool chronic_mgr_send_report(const chronic_report_t *r, uint32_t t)
{ (void)r; (void)t; return false; }
bool chronic_mgr_trigger_test(strip_type_t t)        { (void)t; return false; }
bool chronic_mgr_trigger_bp_measure(void)            { return false; }
chronic_ctx_t *chronic_mgr_get_ctx(void)             { return NULL; }

#endif /* TARGET_PRO_PLUS */
