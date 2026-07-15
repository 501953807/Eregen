/*
 * Eregen (颐贞) - Message Encode Implementation
 * Wire format: JSON payload with CRC16-CCITT appended as last 2 bytes.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "message_encode.h"
#include "../common/crc16.h"
#include <stdio.h>
#include <string.h>

#ifdef TEST_MODE
int message_encode_test(void);
int main(void) { return message_encode_test(); }
#endif

int message_encode(const eregen_msg_t *msg, uint8_t *out, uint16_t out_len)
{
    if (!msg || !out || out_len == 0) {
        return -1;
    }

    /* Validate message fields */
    if (msg->type < MSG_HEARTBEAT || msg->type > MSG_MED_STATUS) {
        return -2;
    }
    if (msg->payload_len > MAX_PAYLOAD_LEN) {
        return -3;
    }
    if (strlen(msg->dev_id) >= sizeof(msg->dev_id)) {
        return -4;
    }

    /* Build JSON string: {"type":N,"dev_id":"...","ts":N,"data":"..."} */
    char json_buf[256];
    int json_len;

    if (msg->payload_len > 0 && msg->payload_len < 128) {
        /* Small payload: embed as JSON string value */
        json_len = snprintf(json_buf, sizeof(json_buf),
            "{\"type\":%u,\"dev_id\":\"%s\",\"ts\":%lu,\"data\":\"",
            (unsigned)msg->type,
            msg->dev_id,
            (unsigned long)msg->timestamp);

        /* Append payload bytes as hex */
        for (uint16_t i = 0; i < msg->payload_len; i++) {
            int remaining = (int)sizeof(json_buf) - json_len - 3;
            if (remaining <= 0) break;
            json_len += snprintf(json_buf + json_len, (size_t)remaining,
                                 "%02X", msg->payload[i]);
        }
        json_len += snprintf(json_buf + json_len, (size_t)(sizeof(json_buf) - json_len),
                             "\"}");
    } else {
        /* Large or empty payload: use payload_len as integer */
        json_len = snprintf(json_buf, sizeof(json_buf),
            "{\"type\":%u,\"dev_id\":\"%s\",\"ts\":%lu,\"payload_len\":%u}",
            (unsigned)msg->type,
            msg->dev_id,
            (unsigned long)msg->timestamp,
            (unsigned)msg->payload_len);
    }

    if (json_len <= 0 || json_len >= (int)sizeof(json_buf)) {
        return -5;  /* Buffer overflow */
    }

    /* Calculate total needed: json + 2-byte CRC16 */
    uint16_t total_needed = (uint16_t)(json_len + 2);
    if (total_needed > out_len) {
        return -6;  /* Output buffer too small */
    }

    /* Copy JSON to output */
    memcpy(out, json_buf, (size_t)json_len);

    /* Calculate CRC16 over the JSON data */
    uint16_t crc = crc16_calc((const uint8_t *)json_buf, (uint16_t)json_len);

    /* Append CRC16 as big-endian */
    out[json_len]     = (uint8_t)((crc >> 8) & 0xFF);
    out[json_len + 1] = (uint8_t)(crc & 0xFF);

    return (int)total_needed;
}
