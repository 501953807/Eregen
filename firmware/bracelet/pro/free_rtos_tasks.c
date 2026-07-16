/*
 * Eregen (颐贞) - Pro Tier FreeRTOS Task Implementation
 * ECG, AMOLED, and GNSS task implementations for the Pro bracelet.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "free_rtos_tasks.h"
#include "ecg_driver.h"
#include "display_amoled.h"
#include "gps_gnss.h"
#include "board_pro.h"
#include "../common/log.h"
#include "../protocol/message_encode.h"
#include <string.h>

/* Queue handles */
static QueueHandle_t s_ecg_health_queue = NULL;
static QueueHandle_t s_arrhythmia_queue = NULL;

/* Device state handles */
static ecg_device_t s_ecg_dev;
static TaskHandle_t s_ecg_task_handle = NULL;
static TaskHandle_t s_amoled_task_handle = NULL;
static TaskHandle_t s_gnss_task_handle = NULL;

/* ----------------------------------------------------------------
 * Initialization
 * ---------------------------------------------------------------- */

bool pro_tasks_init(void)
{
    s_ecg_health_queue = xQueueCreate(QUEUE_ECG_SIZE, sizeof(pro_health_data_t));
    s_arrhythmia_queue = xQueueCreate(QUEUE_ARRITHMIA_SIZE, sizeof(pro_arrhythmia_alert_t));

    if (!s_ecg_health_queue || !s_arrhythmia_queue) {
        log_error("ProTasks: Failed to create queues");
        return false;
    }

    log_info("ProTasks: Queues created (ECG=%u, Arrithmia=%u)",
             QUEUE_ECG_SIZE, QUEUE_ARRITHMIA_SIZE);
    return true;
}

bool pro_tasks_send_ecg_health(const pro_health_data_t *data, uint32_t timeout_ms)
{
    if (!data || !s_ecg_health_queue) {
        return false;
    }

    BaseType_t ret = xQueueSend(s_ecg_health_queue, data,
                                pdMS_TO_TICKS(timeout_ms));
    return (ret == pdPASS);
}

bool pro_tasks_send_arrhythmia_alert(const pro_arrhythmia_alert_t *alert,
                                     uint32_t timeout_ms)
{
    if (!alert || !s_arrhythmia_queue) {
        return false;
    }

    BaseType_t ret = xQueueSendToBack(s_arrhythmia_queue, alert,
                                       pdMS_TO_TICKS(timeout_ms));
    return (ret == pdPASS);
}

/* ----------------------------------------------------------------
 * ECG Task - 200Hz sampling with arrhythmia detection
 * ---------------------------------------------------------------- */

static void vECGTask(void *pvParameters)
{
    (void)pvParameters;

    /* Initialize ECG hardware */
    if (!ecg_init(&s_ecg_dev)) {
        log_error("ProECG: ECG chip init failed");
        /* Blink amber LED to indicate error */
        for (;;) {
            gpio_bit_toggle(BOARD_PRO_LED_AMBER_PORT, BOARD_PRO_LED_AMBER_PIN);
            vTaskDelay(pdMS_TO_TICKS(200));
        }
    }

    log_info("ProECG: ECG initialized, starting acquisition...");

    /* Start continuous ECG measurement */
    if (!ecg_start_measure(&s_ecg_dev)) {
        log_error("ProECG: Failed to start measurement");
        for (;;) {
            gpio_bit_toggle(BOARD_PRO_LED_AMBER_PORT, BOARD_PRO_LED_AMBER_PIN);
            vTaskDelay(pdMS_TO_TICKS(500));
        }
    }

    /* Timing variables for 200Hz sampling */
    uint32_t sample_period_ms = 1000U / ecg_get_sample_rate(&s_ecg_dev); /* 5ms */
    uint32_t ecg_tick = 0;
    uint32_t health_interval = 1000U; /* Send health data every 1 second */
    uint32_t arrhythmia_interval = 30000U; /* Check arrhythmia every 30 seconds */
    uint32_t last_health_send = 0;
    uint32_t last_arrhythmia_check = 0;

    /* Batch buffer for R-peak detection */
    ecg_sample_t sample_batch[20];

    for (;;) {
        uint32_t now = xTaskGetTickCount() * (1000U / configTICK_RATE_HZ);

        /* Read ECG samples at 200Hz (every 5ms) */
        if (++ecg_tick >= sample_period_ms) {
            ecg_tick = 0;

            /* Read batch for R-peak detection */
            uint8_t n = ecg_read_batch(&s_ecg_dev, sample_batch, 20);

            /* Run arrhythmia detection periodically */
            if ((now - last_arrhythmia_check) >= arrhythmia_interval) {
                last_arrhythmia_check = now;

                ecg_arrhythmia_result_t result = ecg_detect_arrhythmia(&s_ecg_dev);
                ecg_lead_off_check(&s_ecg_dev);

                if (result.afib_detected &&
                    result.alert_counter >= ECG_AFRIB_ALERT_COUNT) {
                    /* Send critical arrhythmia alert */
                    pro_arrhythmia_alert_t alert;
                    alert.afib_detected = true;
                    alert.rr_stddev_ms = result.rr_stddev_ms;
                    alert.timestamp = now;
                    alert.severity = 3; /* Critical */

                    if (pro_tasks_send_arrhythmia_alert(&alert, pdMS_TO_TICKS(100))) {
                        log_error("ProECG: CRITICAL - AFib alert sent!");
                        /* Flash amber LED rapidly */
                        for (uint8_t i = 0; i < 8; i++) {
                            gpio_bit_toggle(BOARD_PRO_LED_AMBER_PORT,
                                            BOARD_PRO_LED_AMBER_PIN);
                            vTaskDelay(pdMS_TO_TICKS(50));
                        }
                    }
                }
            }

            /* Compute heart rate from R-R interval */
            pro_health_data_t health_msg;
            memset(&health_msg, 0, sizeof(health_msg));

            if (s_ecg_dev.last_rpeak_ms > 0) {
                float rr_ms = (float)(now - s_ecg_dev.last_rpeak_ms);
                if (rr_ms > 200.0f && rr_ms < 2000.0f) { /* Valid HR range: 30-300 bpm */
                    health_msg.ecg_rr_interval = rr_ms;
                    health_msg.hr = (uint16_t)(60000.0f / rr_ms);
                    health_msg.ecg_valid = true;
                }
            }

            health_msg.ecg_peak_uv = s_ecg_dev.last_sample.raw_ecg_uv;

            /* Send health data periodically */
            if ((now - last_health_send) >= health_interval) {
                last_health_send = now;
                if (health_msg.ecg_valid) {
                    pro_tasks_send_ecg_health(&health_msg, pdMS_TO_TICKS(100));
                    log_info("ProECG: HR=%u bpm, RR=%.1f ms, ECG=%d uV",
                             health_msg.hr, health_msg.ecg_rr_interval,
                             health_msg.ecg_peak_uv);
                }
            }
        }

        vTaskDelay(pdMS_TO_TICKS(1)); /* Tick at 1ms granularity */
    }
}

/* ----------------------------------------------------------------
 * AMOLED Display Task - UI rendering
 * ---------------------------------------------------------------- */

static void vAmoledTask(void *pvParameters)
{
    (void)pvParameters;

    /* Initialize AMOLED display */
    if (!amoled_display_init()) {
        log_error("ProAMOLED: Display init failed");
        for (;;) {
            gpio_bit_toggle(BOARD_PRO_LED_GREEN_PORT, BOARD_PRO_LED_GREEN_PIN);
            vTaskDelay(pdMS_TO_TICKS(500));
        }
    }

    log_info("ProAMOLED: Display initialized (%ux%u)",
             amoled_get_width(), amoled_get_height());

    /* Show startup screen */
    amoled_display_clear(AMOLED_BLACK);

    /* Draw brand header with gradient */
    amoled_draw_gradient_v(0, 0, AMOLED_WIDTH - 1, 40,
                           AMOLED_TEAL, AMOLED_BLACK);
    amoled_draw_string(60, 8, "Eregen", AMOLED_WHITE, AMOLED_TEAL);
    amoled_draw_string(30, 22, "颐人", AMOLED_LIGHT_GRAY, AMOLED_TEAL);
    amoled_draw_line_h(0, 40, AMOLED_WIDTH - 1, AMOLED_DARK_GRAY);

    /* Draw status bar area */
    amoled_draw_rect_filled(0, 41, AMOLED_WIDTH - 1, 55, AMOLED_DARK_GRAY);

    /* Main area: health rings (heart rate, steps, SpO2) */
    uint16_t center_x = AMOLED_WIDTH / 2;
    uint16_t center_y = 180;
    uint16_t ring_r = 80;

    /* Background circle */
    amoled_draw_circle(center_x, center_y, ring_r + 4, AMOLED_DARK_GRAY);
    amoled_draw_circle(center_x, center_y, ring_r, AMOLED_BLACK);

    /* Heart rate ring (red arc) */
    amoled_draw_arc(center_x, center_y, ring_r, 0, 180, 6, AMOLED_RED);

    /* Steps ring (green arc) */
    amoled_draw_arc(center_x, center_y, ring_r - 12, 0, 270, 4, AMOLED_GREEN);

    /* SpO2 ring (blue arc) */
    amoled_draw_arc(center_x, center_y, ring_r - 24, 0, 120, 4, AMOLED_BLUE);

    /* Center text placeholder */
    amoled_draw_string(center_x - 20, center_y - 10, "HR", AMOLED_WHITE, AMOLED_BLACK);
    amoled_draw_string(center_x - 10, center_y + 5, "--", AMOLED_WHITE, AMOLED_BLACK);

    amoled_display_update();

    /* Main loop: update UI periodically */
    uint32_t ui_tick = 0;
    for (;;) {
        vTaskDelay(pdMS_TO_TICKS(2000)); /* Update every 2 seconds */

        ui_tick++;

        /* Redraw battery indicator in top-right */
        amoled_draw_rect_filled(AMOLED_WIDTH - 24, 42, AMOLED_WIDTH - 2, 54, AMOLED_BLACK);
        char batt_str[8];
        snprintf(batt_str, sizeof(batt_str), "%us", (unsigned)ui_tick);
        amoled_draw_string(AMOLED_WIDTH - 40, 44, batt_str, AMOLED_GREEN, AMOLED_BLACK);

        /* Redraw GNSS status */
        if (gnss_has_valid_fix()) {
            amoled_draw_string(2, 44, "GPS", AMOLED_GREEN, AMOLED_BLACK);
        } else {
            amoled_draw_string(2, 44, "GPS..", AMOLED_YELLOW, AMOLED_BLACK);
        }

        /* Update health rings with live data */
        pro_health_data_t health_msg;
        /* In production, dequeue from health queue. For now, redraw static rings. */
        amoled_draw_arc(center_x, center_y, ring_r, 0,
                        (uint16_t)(180 + (ui_tick * 30) % 180), 6, AMOLED_RED);

        amoled_display_update();
    }
}

/* ----------------------------------------------------------------
 * GNSS Task - Multi-constellation positioning
 * ---------------------------------------------------------------- */

static void vGNSSTask(void *pvParameters)
{
    (void)pvParameters;

    /* Enable GNSS module */
    gpio_init(BOARD_PRO_GNSS_EN_PORT, GPIO_MODE_OUT_PP,
              GPIO_OSPEED_50MHZ, BOARD_PRO_GNSS_EN_PIN);
    gpio_bit_set(BOARD_PRO_GNSS_EN_PORT, BOARD_PRO_GNSS_EN_PIN);
    vTaskDelay(pdMS_TO_TICKS(500)); /* Warm-up time */

    /* Initialize GNSS parser */
    gnss_init();
    log_info("ProGNSS: Multi-constellation parser initialized");

    for (;;) {
        /* Read characters from GNSS UART and feed to parser */
        while (usart_flag_get(BOARD_PRO_GNSS_UART, USART_FLAG_RBNE) != RESET) {
            char c = (char)usart_data_receive(BOARD_PRO_GNSS_UART);
            gnss_parse_char(c);
        }

        /* Check for valid fix */
        if (gnss_has_valid_fix()) {
            gnss_fix_t fix = gnss_get_fix();

            /* Log fix details */
            log_info("ProGNSS: lat=%.6f, lon=%.6f, sats=%u, hdop=%.2f, acc=%um%s",
                     fix.latitude, fix.longitude, fix.satellites, fix.hdop,
                     fix.accuracy_m, fix.multi_const ? " [MULTI]" : "");

            /* Check for precision fix (cm-level) */
            if (gnss_has_precision_fix()) {
                log_info("ProGNSS: PRECISION FIX: hdop=%.2f, multi-const=true", fix.hdop);
            }
        }

        /* Periodically log SV visibility */
        static uint32_t sv_tick = 0;
        if (++sv_tick >= 600) { /* Every ~60 seconds */
            sv_tick = 0;
            gnss_sv_t visible[GNSS_MAX_SVS];
            uint8_t count = gnss_get_visible_svs(visible, GNSS_MAX_SVS);

            uint8_t gps_count = 0, glo_count = 0, gal_count = 0;
            for (uint8_t i = 0; i < count; i++) {
                switch (visible[i].constell) {
                    case GNSS_CONST_GPS:     gps_count++; break;
                    case GNSS_CONST_GLONASS: glo_count++; break;
                    case GNSS_CONST_GALILEO: gal_count++; break;
                    default: break;
                }
            }
            log_info("ProGNSS: Visible SVs=%u (GPS=%u, GLONASS=%u, Galileo=%u)",
                     count, gps_count, glo_count, gal_count);
        }

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/* ----------------------------------------------------------------
 * Task creation
 * ---------------------------------------------------------------- */

bool pro_tasks_create(void)
{
    /* Create ECG task */
    BaseType_t ret = xTaskCreate(vECGTask,
                                  "ProECG",
                                  TASK_PRO_ECG_STACK,
                                  NULL,
                                  TASK_PRO_ECG_PRIORITY,
                                  &s_ecg_task_handle);
    if (ret != pdPASS) {
        log_error("ProTasks: Failed to create ECG task");
        return false;
    }

    /* Create AMOLED display task */
    ret = xTaskCreate(vAmoledTask,
                      "ProAMOLED",
                      TASK_PRO_AMOLED_STACK,
                      NULL,
                      TASK_PRO_AMOLED_PRIORITY,
                      &s_amoled_task_handle);
    if (ret != pdPASS) {
        log_error("ProTasks: Failed to create AMOLED task");
        return false;
    }

    /* Create GNSS task */
    ret = xTaskCreate(vGNSSTask,
                      "ProGNSS",
                      TASK_PRO_GNSS_STACK,
                      NULL,
                      TASK_PRO_GNSS_PRIORITY,
                      &s_gnss_task_handle);
    if (ret != pdPASS) {
        log_error("ProTasks: Failed to create GNSS task");
        return false;
    }

    log_info("ProTasks: All Pro tasks created successfully");
    return true;
}
