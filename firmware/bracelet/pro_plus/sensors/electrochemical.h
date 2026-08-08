/*
 * Eregen (颐贞) - Pro+ 试纸检测模块 · 电化学检测驱动头文件
 * 负责控制 GD32 ADC 完成电化学试纸的电流-电压采样，
 * 并调用校准系数换算出生物标志物浓度。
 *
 * 硬件接口：
 *   - ADC1 (PA0)   : 电流采样（跨阻放大器输出）
 *   - ADC1 (PA1)   : 工作电极参考电压
 *   - GPIO PC4     : 试纸插入检测（上拉，插入接地）
 *   - GPIO PC5     : 电化学模块供电使能
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef PRO_PLUS_ELECTROCHEMICAL_H
#define PRO_PLUS_ELECTROCHEMICAL_H

#include <stdint.h>
#include <stdbool.h>
#include "strip_type.h"
#include "electrochemical_msg.h"

/* ----------------------------------------------------------------
 * 硬件引脚定义（Pro+ 专属）
 * ---------------------------------------------------------------- */

/* 试纸插入检测：PC4，高=无试纸，低=插入 */
#define ELEC_DETECT_PORT    GPIOC
#define ELEC_DETECT_PIN     GPIO_PIN_4

/* 电化学模块电源使能：PC5，高=上电 */
#define ELEC_POWER_PORT     GPIOC
#define ELEC_POWER_PIN      GPIO_PIN_5

/* ADC 通道 */
#define ELEC_ADC            ADC1
#define ELEC_ADC_IN_CUR     GPIO_ADC_IN_0   /* PA0 — 电流采样 */
#define ELEC_ADC_IN_REF     GPIO_ADC_IN_1   /* PA1 — 参考电压 */
#define ELEC_ADC_GPIO_PORT  GPIOA

/* 采样参数 */
#define ELEC_SAMPLE_COUNT   64U             /* 每次采集样本数 */
#define ELEC_AVG_OVERSAMPLE 8U              /* 过采样平均次数 */

/* 电流-电压转换系数（跨阻放大器增益） */
#define ELEC_TIA_GAIN_MV_PER_UA  10.0f     /* 10 mV/μA */

/* 参考电压（mV） */
#define ELEC_REF_VOLTAGE_MV      3300U

/* ----------------------------------------------------------------
 * 检测结果回调函数指针
 * ---------------------------------------------------------------- */
typedef void (*elec_result_cb_t)(const detection_result_t *result, void *ctx);

/* ----------------------------------------------------------------
 * 电化学检测器上下文
 * ---------------------------------------------------------------- */
typedef struct {
    bool            powered;              /* 模块是否已上电 */
    strip_type_t    current_type;         /* 当前插入试纸类型 */
    strip_state_t   state;                /* 当前状态机状态 */
    uint32_t        measure_start_ms;     /* 测量起始时间 */
    float           calibration[STRIP_CALIB_POINTS];  /* 校准系数数组 */
    uint8_t         calibr_point_count;   /* 已校准点数 */
    elec_result_cb_t result_cb;           /* 结果回调 */
    void           *result_cb_ctx;        /* 回调上下文 */
} electrochemical_dev_t;

/* ----------------------------------------------------------------
 * 公共 API
 * ---------------------------------------------------------------- */

/**
 * 初始化电化学检测模块（GPIO + ADC）。
 * @param dev 设备上下文指针。
 * @return true 表示初始化成功。
 */
bool elec_detect_init(electrochemical_dev_t *dev);

/**
 * 检测试纸是否插入。
 * @param dev 设备上下文。
 * @return true 表示试纸已插入。
 */
bool elec_detect_strip_inserted(const electrochemical_dev_t *dev);

/**
 * 读取试纸类型（通过 ID 引脚电压分压判断）。
 * 简化实现：返回 STRIP_TYPE_NONE，实际硬件通过分压识别。
 * @param dev 设备上下文。
 * @return 试纸类型枚举值。
 */
strip_type_t elec_detect_strip_type(electrochemical_dev_t *dev);

/**
 * 上电电化学模块。
 * @param dev 设备上下文。
 */
void elec_power_on(electrochemical_dev_t *dev);

/**
 * 下电电化学模块（省电）。
 * @param dev 设备上下文。
 */
void elec_power_off(electrochemical_dev_t *dev);

/**
 * 对指定类型试纸执行一次完整检测。
 * 包括基线采样 → 电解反应 → 峰值读取 → 浓度换算。
 * @param dev     设备上下文。
 * @param type    试纸类型。
 * @param out     输出检测结果。
 * @param timeout_ms 最大等待时间（ms）。
 * @return true 表示检测完成且结果有效。
 */
bool elec_run_measurement(electrochemical_dev_t *dev,
                          strip_type_t type,
                          detection_result_t *out,
                          uint32_t timeout_ms);

/**
 * 使用已知浓度标准液完成单点校准，更新校准系数。
 * @param dev     设备上下文。
 * @param type    试纸类型。
 * @param known_value 标准液已知浓度。
 * @param unit    单位字符串。
 */
void elec_calibrate_single(electrochemical_dev_t *dev,
                           strip_type_t type,
                           float known_value,
                           const char *unit);

/**
 * 注册检测结果回调。
 * @param dev     设备上下文。
 * @param cb      回调函数。
 * @param ctx     回调上下文指针。
 */
void elec_set_result_callback(electrochemical_dev_t *dev,
                              elec_result_cb_t cb,
                              void *ctx);

/**
 * 将检测结果编码为 MQTT JSON 消息。
 * @param result  检测结果。
 * @param out     输出缓冲区。
 * @param out_len 缓冲区大小。
 * @return 写入字节数，负数表示错误。
 */
int elec_result_to_json(const detection_result_t *result,
                        char *out, uint16_t out_len);

#endif /* PRO_PLUS_ELECTROCHEMICAL_H */
