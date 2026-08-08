/*
 * Eregen (颐贞) - Pro+ 试纸检测模块 · 试纸类型与检测结果定义
 * 电化学试纸检测用于血糖、尿酸、血脂等慢病指标测量。
 *
 * 试纸类型枚举 + 检测结果结构体，供 sensors/electrochemical.* 使用。
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef PRO_PLUS_STRIP_TYPE_H
#define PRO_PLUS_STRIP_TYPE_H

#include <stdint.h>
#include <stdbool.h>

/* ----------------------------------------------------------------
 * 试纸类型枚举
 * 每种试纸对应不同生物标志物检测通道和校准参数
 * ---------------------------------------------------------------- */
typedef enum {
    STRIP_TYPE_NONE   = 0U,   /* 未插入试纸 */
    STRIP_TYPE_GLUCOSE,         /* 血糖检测（葡萄糖氧化酶法） */
    STRIP_TYPE_URIC_ACID,       /* 尿酸检测（尿酸酶法） */
    STRIP_TYPE_LIPID,           /* 血脂检测（甘油三酯法） */
    STRIP_TYPE_CHOLESTEROL,     /* 总胆固醇检测 */
    STRIP_TYPE_MAX,
} strip_type_t;

/* ----------------------------------------------------------------
 * 试纸状态枚举
 * ---------------------------------------------------------------- */
typedef enum {
    STRIP_STATE_IDLE       = 0U,   /* 待机，无试纸 */
    STRIP_STATE_INSERTED,          /* 试纸已插入 */
    STRIP_STATE_CALIBRATING,       /* 校准中 */
    STRIP_STATE_MEASURING,         /* 测量中 */
    STRIP_STATE_COMPLETE,          /* 测量完成 */
    STRIP_STATE_ERROR,             /* 检测错误 */
} strip_state_t;

/* ----------------------------------------------------------------
 * 单次检测原始数据
 * ---------------------------------------------------------------- */
typedef struct {
    int32_t  raw_voltage_uv;    /* 原始电压（微伏） */
    float    current_uA;        /* 电流（微安） */
    float    temperature_c;     /* 检测时温度（摄氏度） */
    uint32_t measurement_ms;    /* 测量耗时（毫秒） */
    uint8_t  quality_score;     /* 0-100，试纸接触质量 */
} strip_raw_sample_t;

/* ----------------------------------------------------------------
 * 检测结果（经校准换算后）
 * ---------------------------------------------------------------- */
typedef struct {
    strip_type_t  strip_type;   /* 试纸类型 */
    strip_state_t state;        /* 当前状态 */
    bool          valid;        /* 结果是否有效 */

    /* 分析结果：根据 strip_type 语义不同 */
    union {
        float glucose_mmol_l;    /* 血糖 mmol/L */
        float uric_acid_umol_l;  /* 尿酸 μmol/L */
        float triglyceride_mmol_l; /* 甘油三酯 mmol/L */
        float cholesterol_mmol_l;  /* 总胆固醇 mmol/L */
    } result;

    float    baseline_current_uA; /* 基线电流（校准用） */
    float    peak_current_uA;     /* 峰值电流 */
    uint32_t timestamp_s;         /* 检测时间戳（UTC秒） */
    uint8_t  quality_score;       /* 0-100 */
} detection_result_t;

/* ----------------------------------------------------------------
 * 常量：试纸检测参数
 * ---------------------------------------------------------------- */

/* 最大测量耗时（ms）— 超过则报错 */
#define STRIP_MAX_MEASURE_MS   30000U

/* 基线采样窗口（ms） */
#define STRIP_BASELINE_WINDOW_MS  2000U

/* 校准参数数量 */
#define STRIP_CALIB_POINTS        5U

/* 测量周期（ms） */
#define STRIP_MEASURE_INTERVAL_MS 100U

/* ----------------------------------------------------------------
 * 试纸校准系数
 * ---------------------------------------------------------------- */
typedef struct {
    float slope;                    /* 校准斜率 */
    float intercept;                /* 校准截距 */
    float temperature_coeff;        /* 温度补偿系数 (1/°C) */
    strip_type_t type;              /* 对应试纸类型 */
} strip_calibration_t;

#endif /* PRO_PLUS_STRIP_TYPE_H */
