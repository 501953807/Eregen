/*
 * Eregen (颐贞) - Pro+ 试纸检测模块 · 电化学检测驱动实现
 * 完成 ADC 采样、信号处理、浓度换算和 MQTT 消息编码。
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "electrochemical.h"

#ifdef TARGET_PRO_PLUS

#include <stdio.h>
#include <string.h>
#include "gd32e230_gpio.h"
#include "gd32e230_adc.h"
#include "gd32e230_rcu.h"
#include "../common/log.h"

/* ----------------------------------------------------------------
 * 内部辅助函数：ADC 采样（带过采样平均）
 * ---------------------------------------------------------------- */

static float adc_read_averaged(ADC_TypeDef *adc, uint32_t channel, uint8_t samples)
{
    uint32_t sum = 0;
    for (uint8_t i = 0; i < samples; i++) {
        adc_regular_channel_set(adc, 0, channel);
        adc_software_trigger_enable(adc, ADC_REGULAR_CHANNEL);
        /* 等待转换完成 */
        while (adc_flag_get(adc, ADC_FLAG_EOC) == RESET) {}
        sum += adc_regular_data_get(adc);
    }
    /* 12-bit ADC: 0-4095, 对应 0-3300mV */
    return ((float)sum / (float)samples) * (3300.0f / 4095.0f);
}

/* ----------------------------------------------------------------
 * 内部辅助函数：读取电流（通过跨阻放大器电压）
 * ---------------------------------------------------------------- */

static float elec_read_current_uA(const electrochemical_dev_t *dev)
{
    (void)dev;
    /* 简化实现：模拟返回一个随机范围内的电流值
     * 实际硬件需在 init 后接入真实 ADC */
    float voltage_mv = adc_read_averaged(ADC1, ELEC_ADC_IN_CUR, ELEC_AVG_OVERSAMPLE);
    return voltage_mv / ELEC_TIA_GAIN_MV_PER_UA;
}

/* ----------------------------------------------------------------
 * 公共 API 实现
 * ---------------------------------------------------------------- */

bool elec_detect_init(electrochemical_dev_t *dev)
{
    if (!dev) return false;

    memset(dev, 0, sizeof(*dev));
    dev->state = STRIP_STATE_IDLE;
    dev->current_type = STRIP_TYPE_NONE;

    /* 使能 GPIO 时钟 */
    rcu_periph_clock_enable(RCU_GPIOC);
    rcu_periph_clock_enable(RCU_GPIOA);
    rcu_periph_clock_enable(RCU_ADC1);

    /* PC4: 试纸检测输入（上拉） */
    gpio_init(ELEC_DETECT_PORT, GPIO_MODE_IPU,
              GPIO_OSPEED_50MHZ, ELEC_DETECT_PIN);

    /* PC5: 电化学模块电源使能（推挽输出） */
    gpio_init(ELEC_POWER_PORT, GPIO_MODE_OUT_PP,
              GPIO_OSPEED_50MHZ, ELEC_POWER_PIN);
    gpio_bit_reset(ELEC_POWER_PORT, ELEC_POWER_PIN);  /* 默认关闭 */

    /* PA0/PA1: ADC 模拟输入 */
    gpio_init(ELEC_ADC_GPIO_PORT, GPIO_MODE_AIN,
              GPIO_OSPEED_50MHZ, GPIO_PIN_0 | GPIO_PIN_1);

    /* ADC 初始化 */
    adc_deinit(ADC1);
    adc_mode_config(ADC_MODE_FREE);
    adc_special_function_config(ADC1, ADC_CONTINUOUS_MODE, ENABLE);
    adc_resolution_config(ADC1, ADC_RESOLUTION_12B);
    adc_external_trigger_source_config(ADC1, ADC_REGULAR_CHANNEL,
                                       ADC_EXTTRIG_REGULAR_NONE);
    adc_external_trigger_config(ADC1, ADC_REGULAR_CHANNEL, ENABLE);
    adc_data_alignment_config(ADC1, ADC_DATAALIGN_RIGHT);
    adc_channel_length_config(ADC1, ADC_REGULAR_CHANNEL, 1);

    adc_regular_channel_config(0, ELEC_ADC_IN_CUR, ADC_SAMPLETIME_55POINT5);

    adc_enable(ADC1);
    /* ADC 稳定延时 */
    for (volatile uint32_t i = 0; i < 200000U; i++) {}

    log_info("ElecDetect: Module initialized (PC4=det, PC5=pwr)");
    return true;
}

bool elec_detect_strip_inserted(const electrochemical_dev_t *dev)
{
    if (!dev) return false;
    /* PC4 低电平 = 试纸插入（接地） */
    return (gpio_input_data_bit_get(ELEC_DETECT_PORT, ELEC_DETECT_PIN) == RESET);
}

strip_type_t elec_detect_strip_type(electrochemical_dev_t *dev)
{
    if (!dev) return STRIP_TYPE_NONE;
    /*
     * 简化实现：实际硬件通过 ID 引脚分压判断。
     * 这里返回 STRIP_TYPE_GLUCOSE 作为占位。
     * 正式固件应根据 ID 引脚 ADC 读数查表。
     */
    return STRIP_TYPE_GLUCOSE;
}

void elec_power_on(electrochemical_dev_t *dev)
{
    if (!dev || dev->powered) return;
    gpio_bit_set(ELEC_POWER_PORT, ELEC_POWER_PIN);
    dev->powered = true;
    /* 等待模块稳定 */
    vTaskDelay(pdMS_TO_TICKS(100));
    log_info("ElecDetect: Power ON");
}

void elec_power_off(electrochemical_dev_t *dev)
{
    if (!dev || !dev->powered) return;
    gpio_bit_reset(ELEC_POWER_PORT, ELEC_POWER_PIN);
    dev->powered = false;
    dev->state = STRIP_STATE_IDLE;
    log_info("ElecDetect: Power OFF");
}

bool elec_run_measurement(electrochemical_dev_t *dev,
                          strip_type_t type,
                          detection_result_t *out,
                          uint32_t timeout_ms)
{
    if (!dev || !out) return false;

    elec_power_on(dev);
    dev->state = STRIP_STATE_CALIBRATING;
    dev->measure_start_ms = xTaskGetTickCount() * (1000U / configTICK_RATE_HZ);

    /* 阶段1：基线采样 */
    float baseline = elec_read_current_uA(dev);
    dev->state = STRIP_STATE_MEASURING;

    /* 阶段2：等待峰值（模拟：固定延时） */
    uint32_t elapsed = 0;
    float peak_current = baseline;
    while (elapsed < timeout_ms) {
        float cur = elec_read_current_uA(dev);
        if (cur > peak_current) {
            peak_current = cur;
        }
        vTaskDelay(pdMS_TO_TICKS(STRIP_MEASURE_INTERVAL_MS));
        elapsed += STRIP_MEASURE_INTERVAL_MS;
        if (elapsed >= timeout_ms) break;
    }

    /* 阶段3：浓度换算（线性校准模型） */
    float value = 0.0f;
    const char *unit = "mmol/L";
    uint8_t quality = 85U;

    if (peak_current > baseline * 0.1f) {
        quality = (uint8_t)(80.0f + 20.0f * (peak_current / (baseline + 1.0f)));
        if (quality > 100U) quality = 100U;
    } else {
        quality = 20U;  /* 信号太弱，试纸可能失效 */
    }

    /* 简化换算：value = (peak - baseline) * slope */
    switch (type) {
        case STRIP_TYPE_GLUCOSE:
            value = (peak_current - baseline) * 0.5f;  /* 占位系数 */
            unit = "mmol/L";
            break;
        case STRIP_TYPE_URIC_ACID:
            value = (peak_current - baseline) * 2.0f;
            unit = "μmol/L";
            break;
        case STRIP_TYPE_LIPID:
            value = (peak_current - baseline) * 1.0f;
            unit = "mmol/L";
            break;
        case STRIP_TYPE_CHOLESTEROL:
            value = (peak_current - baseline) * 0.8f;
            unit = "mmol/L";
            break;
        default:
            quality = 0U;
            break;
    }

    /* 填充结果 */
    memset(out, 0, sizeof(*out));
    out->strip_type = type;
    out->state = STRIP_STATE_COMPLETE;
    out->valid = (quality >= 50U);
    out->quality_score = quality;
    out->baseline_current_uA = baseline;
    out->peak_current_uA = peak_current;
    out->timestamp_s = (uint32_t)xTaskGetTickCount() * (1000U / configTICK_RATE_HZ);

    /* 按类型填入对应字段 */
    switch (type) {
        case STRIP_TYPE_GLUCOSE:
            out->result.glucose_mmol_l = value;
            break;
        case STRIP_TYPE_URIC_ACID:
            out->result.uric_acid_umol_l = value;
            break;
        case STRIP_TYPE_LIPID:
            out->result.triglyceride_mmol_l = value;
            break;
        case STRIP_TYPE_CHOLESTEROL:
            out->result.cholesterol_mmol_l = value;
            break;
        default:
            break;
    }

    dev->state = STRIP_STATE_COMPLETE;
    elec_power_off(dev);

    log_info("ElecDetect: type=%d, value=%.2f %s, quality=%u",
             (int)type, value, unit, (unsigned)quality);

    /* 调用回调 */
    if (dev->result_cb && out->valid) {
        dev->result_cb(out, dev->result_cb_ctx);
    }

    return out->valid;
}

void elec_calibrate_single(electrochemical_dev_t *dev,
                           strip_type_t type,
                           float known_value,
                           const char *unit)
{
    if (!dev) return;
    (void)unit;

    /* 简化：存储已知值用于后续换算 */
    log_info("ElecDetect: Calibrate type=%d, known=%.4f", (int)type, known_value);
    /* 正式实现需读取标准液电流并拟合 slope/intercept */
}

void elec_set_result_callback(electrochemical_dev_t *dev,
                              elec_result_cb_t cb,
                              void *ctx)
{
    if (!dev) return;
    dev->result_cb = cb;
    dev->result_cb_ctx = ctx;
}

int elec_result_to_json(const detection_result_t *result,
                        char *out, uint16_t out_len)
{
    if (!result || !out || out_len == 0) return -1;

    const char *type_str = "unknown";
    const char *unit = "mmol/L";
    float value = 0.0f;

    switch (result->strip_type) {
        case STRIP_TYPE_GLUCOSE:
            type_str = "glucose";
            value = result->result.glucose_mmol_l;
            unit = "mmol/L";
            break;
        case STRIP_TYPE_URIC_ACID:
            type_str = "uric_acid";
            value = result->result.uric_acid_umol_l;
            unit = "μmol/L";
            break;
        case STRIP_TYPE_LIPID:
            type_str = "lipid";
            value = result->result.triglyceride_mmol_l;
            unit = "mmol/L";
            break;
        case STRIP_TYPE_CHOLESTEROL:
            type_str = "cholesterol";
            value = result->result.cholesterol_mmol_l;
            unit = "mmol/L";
            break;
        default:
            type_str = "unknown";
            break;
    }

    int n = snprintf(out, out_len,
                     "{\"type\":\"strip_result\","
                      "\"strip_type\":\"%s\","
                      "\"value\":%.2f,\"unit\":\"%s\","
                      "\"quality\":%u,\"valid\":%s,"
                      "\"ts\":%u}",
                     type_str, value, unit,
                     (unsigned)result->quality_score,
                     result->valid ? "true" : "false",
                     (unsigned)result->timestamp_s);

    return (n < 0 || (uint32_t)n >= out_len) ? -1 : n;
}

#else /* !TARGET_PRO_PLUS */

/* 当未启用 TARGET_PRO_PLUS 时，提供空实现避免链接错误 */
bool elec_detect_init(electrochemical_dev_t *dev)  { (void)dev; return false; }
bool elec_detect_strip_inserted(const electrochemical_dev_t *d) { (void)d; return false; }
strip_type_t elec_detect_strip_type(electrochemical_dev_t *d) { (void)d; return STRIP_TYPE_NONE; }
void elec_power_on(electrochemical_dev_t *d)  { (void)d; }
void elec_power_off(electrochemical_dev_t *d) { (void)d; }
bool elec_run_measurement(electrochemical_dev_t *d, strip_type_t t,
                          detection_result_t *o, uint32_t timeout)
{ (void)d; (void)t; (void)o; (void)timeout; return false; }
void elec_calibrate_single(electrochemical_dev_t *d, strip_type_t t,
                           float k, const char *u) { (void)d; (void)t; (void)k; (void)u; }
void elec_set_result_callback(electrochemical_dev_t *d,
                              elec_result_cb_t cb, void *ctx)
{ (void)d; (void)cb; (void)ctx; }
int elec_result_to_json(const detection_result_t *r, char *o, uint16_t l)
{ (void)r; (void)o; (void)l; return -1; }

#endif /* TARGET_PRO_PLUS */
