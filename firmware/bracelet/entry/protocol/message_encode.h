/*
 * Eregen (颐贞) - Message Encode Header
 * Encodes eregen_msg_t into wire format (JSON + CRC16).
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef MESSAGE_ENCODE_H
#define MESSAGE_ENCODE_H

#include <stdint.h>
#include <stddef.h>

#define MAX_PAYLOAD_LEN 512U

/** Maximum encoded message length.
 *  1-byte type + 16-byte dev_id + 8-byte timestamp +
 *  4-byte payload_len + MAX_PAYLOAD_LEN + 2-byte CRC16
 */
#define MAX_MSG_LEN (1 + 16 + 8 + 4 + MAX_PAYLOAD_LEN + 2)

/** Message types matching the device-cloud protocol spec. */
typedef enum {
    MSG_HEARTBEAT = 1,
    MSG_LOCATION,
    MSG_HEALTH,
    MSG_SOS,
    MSG_FALL,
    MSG_MED_STATUS
} msg_type_t;

/**
 * Parsed message structure for encode/decode operations.
 */
typedef struct {
    msg_type_t type;         /* Message type */
    char dev_id[17];         /* "BR-XXXX" or "PX-XXXX" format */
    uint64_t timestamp;      /* UTC timestamp in seconds */
    uint8_t payload[MAX_PAYLOAD_LEN];  /* Raw payload bytes */
    uint16_t payload_len;    /* Length of payload data */
    uint16_t checksum;       /* CRC16 over everything before this field */
} eregen_msg_t;

/**
 * Encode a message into wire format (JSON string + CRC16 appended).
 * @param msg    Pointer to message to encode.
 * @param out    Output buffer for encoded bytes.
 * @param out_len Size of output buffer.
 * @return Number of bytes written, or negative on error.
 */
int message_encode(const eregen_msg_t *msg, uint8_t *out, uint16_t out_len);

#endif /* MESSAGE_ENCODE_H */
