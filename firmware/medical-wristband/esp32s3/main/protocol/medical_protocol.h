/*
 * Eregen Medical Wristband - Medical Protocol Header
 * Message formats for medical wristband communication
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef MEDICAL_PROTOCOL_H
#define MEDICAL_PROTOCOL_H

#include <stdint.h>
#include <stdbool.h>

// Message types
#define MSG_TYPE_HEARTBEAT    0x01
#define MSG_TYPE_LOCATION     0x02
#define MSG_TYPE_SOS          0x03
#define MSG_TYPE_ALERT        0x04
#define MSG_TYPE_STATUS       0x05
#define MSG_TYPE_PATIENT_BIND 0x10
#define MSG_TYPE_PATIENT_UNBIND 0x11

// Alert types
#define ALERT_TYPE_FAL        0x01
#define ALERT_TYPE_SOS        0x02
#define ALERT_TYPE_BATTERY_LOW 0x03
#define ALERT_TYPE_UNBOUND    0x04

#pragma pack(push, 1)

typedef struct {
    uint8_t msg_type;
    uint8_t dev_id[8];
    uint16_t sequence;
    uint32_t timestamp;
} medical_msg_header_t;

typedef struct {
    medical_msg_header_t header;
    uint8_t battery_percent;
    uint32_t uptime;
} medical_heartbeat_msg_t;

typedef struct {
    medical_msg_header_t header;
    float latitude;
    float longitude;
    uint8_t accuracy;
    uint8_t battery_percent;
} medical_location_msg_t;

typedef struct {
    medical_msg_header_t header;
    float latitude;
    float longitude;
    uint8_t alert_type;
    uint8_t severity;
} medical_alert_msg_t;

typedef struct {
    medical_msg_header_t header;
    uint8_t patient_id[32];
    uint8_t admission_no[16];
    uint8_t name[16];
} medical_patient_bind_msg_t;

#pragma pack(pop)

#endif /* MEDICAL_PROTOCOL_H */
