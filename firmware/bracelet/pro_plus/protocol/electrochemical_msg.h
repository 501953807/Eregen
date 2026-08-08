/*
 * Eregen (颐贞) - Pro+ 试纸检测模块 · MQTT 电化学消息结构
 * 定义试纸检测结果上报给云端的 MQTT 消息格式。
 *
 * 与 device-cloud 协议 spec 中 type=strip_result 对应。
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef PRO_PLUS_ELECTROCHEMICAL_MSG_H
#define PRO_PLUS_ELECTROCHEMICAL_MSG_H

#include <stdint.h>
#include <stdbool.h>
#include "strip_type.h"
#include "../entry/protocol/message_encode.h"

/* ----------------------------------------------------------------
 * MQTT 消息类型值（与 cloud protocol 对齐）
 * ---------------------------------------------------------------- */
#define MSG_TYPE_STRIP_RESULT   8    /* 试纸检测结果 */

/* ----------------------------------------------------------------
 * 试纸检测结果 MQTT 消息
 * 对应 JSON:
 *   {"type":"strip_result","dev_id":"BR-XXXX",
 *    "strip_type":"glucose","value":5.6,"unit":"mmol/L",
 *    "quality":85,"ts":1720000000}
 * ---------------------------------------------------------------- */
typedef struct {
    msg_type_t    msg_type;       /* MSG_TYPE_STRIP_RESULT = 8 */
    char          dev_id[17];     /* "BR-XXXX" */
    strip_type_t  strip_type;     /* 试纸类型 */
    bool          valid;          /* 结果是否有效 */

    /* 检测结果值（按类型语义不同，由上层填充对应字段） */
    float         value;          /* 换算后浓度值 */
    char          unit[12];       /* 单位字符串，如 "mmol/L" */

    uint8_t       quality;        /* 0-100 */
    float         temperature_c;  /* 检测环境温度 */
    uint32_t      timestamp_s;    /* UTC 时间戳 */
} electrochemical_msg_t;

/* ----------------------------------------------------------------
 * 消息大小限制
 * ---------------------------------------------------------------- */
#define ELEC_MSG_MAX_JSON_LEN  256U   /* JSON 编码上限 */

#endif /* PRO_PLUS_ELECTROCHEMICAL_MSG_H */
