/*
 * Eregen (颐贞) - Message Decode Implementation
 * Verifies CRC16 then parses JSON into eregen_msg_t struct.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "message_decode.h"
#include "../common/crc16.h"
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

#ifdef TEST_MODE
int message_decode_test(void);
int main(void) { return message_decode_test(); }
#endif

/* Helper: skip whitespace */
static const char *skip_ws(const char *p)
{
    while (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r') {
        p++;
    }
    return p;
}

/* Helper: find a JSON key and return pointer to its value (after colon) */
static const char *find_key(const char *json, const char *key)
{
    char search[64];
    int klen = (int)strlen(key);
    snprintf(search, sizeof(search), "\"%s\"", key);

    const char *pos = strstr(json, search);
    if (!pos) return NULL;

    pos += klen + 2;  /* Skip past "key" */
    pos = skip_ws(pos);
    if (*pos != ':') return NULL;
    pos++;
    return skip_ws(pos);
}

/* Helper: parse a JSON integer value (non-negative) */
static int parse_int_val(const char *val_str, uint64_t *out)
{
    if (!val_str || *val_str == '"') {
        return -1;
    }
    char *end;
    unsigned long long v = strtoull(val_str, &end, 10);
    if (end == val_str) return -1;
    *out = (uint64_t)v;
    return 0;
}

/* Helper: parse a JSON string value (null-terminated copy) */
static int parse_string_val(const char *val_str, char *out, size_t out_size)
{
    if (!val_str || *val_str != '"') return -1;
    val_str++;  /* Skip opening quote */

    size_t idx = 0;
    while (*val_str && *val_str != '"' && idx < out_size - 1) {
        if (*val_str == '\\' && *(val_str + 1)) {
            val_str++;
        }
        out[idx++] = *val_str++;
    }
    out[idx] = '\0';

    return (*val_str == '"') ? 0 : -1;
}

/* Helper: parse hex-encoded payload from JSON string value */
static int parse_hex_payload(const char *val_str, uint8_t *out, uint16_t max_len, uint16_t *out_len)
{
    if (!val_str || *val_str != '"') return -1;
    val_str++;  /* Skip opening quote */

    *out_len = 0;
    while (*val_str && *val_str != '"' && *out_len < max_len) {
        unsigned int byte_val;
        if (sscanf(val_str, "%02X", &byte_val) == 1) {
            out[(*out_len)++] = (uint8_t)byte_val;
            val_str += 2;
        } else {
            break;
        }
    }
    return (*val_str == '"') ? 0 : -1;
}

bool message_decode(const uint8_t *in, uint16_t in_len, eregen_msg_t *out)
{
    if (!in || in_len < 3 || !out) {
        return false;
    }

    /* Minimum: at least 1 byte JSON + 2 bytes CRC */

    /* Extract CRC16 from last 2 bytes */
    uint16_t received_crc = ((uint16_t)in[in_len - 2] << 8) | (uint16_t)in[in_len - 1];

    /* Verify CRC over the JSON portion */
    uint16_t computed_crc = crc16_calc(in, in_len - 2);
    if (received_crc != computed_crc) {
        return false;  /* CRC mismatch */
    }

    /* Copy JSON to temp buffer (null-terminated) */
    uint16_t json_len = in_len - 2;
    char json_buf[256];
    if (json_len >= sizeof(json_buf)) {
        return false;  /* Too large for our parser */
    }
    memcpy(json_buf, in, json_len);
    json_buf[json_len] = '\0';

    /* Parse JSON fields */
    memset(out, 0, sizeof(eregen_msg_t));

    /* Parse type */
    const char *type_val = find_key(json_buf, "type");
    if (type_val) {
        uint64_t t;
        if (parse_int_val(type_val, &t) == 0 && t >= MSG_HEARTBEAT && t <= MSG_MED_STATUS) {
            out->type = (msg_type_t)t;
        } else {
            return false;
        }
    } else {
        return false;
    }

    /* Parse dev_id */
    const char *dev_id_val = find_key(json_buf, "dev_id");
    if (dev_id_val) {
        if (parse_string_val(dev_id_val, out->dev_id, sizeof(out->dev_id)) != 0) {
            return false;
        }
    } else {
        return false;
    }

    /* Parse timestamp */
    const char *ts_val = find_key(json_buf, "ts");
    if (ts_val) {
        if (parse_int_val(ts_val, &out->timestamp) != 0) {
            return false;
        }
    } else {
        return false;
    }

    /* Parse payload */
    const char *data_val = find_key(json_buf, "data");
    const char *plen_val = find_key(json_buf, "payload_len");

    if (data_val) {
        /* Hex-encoded payload string */
        if (parse_hex_payload(data_val, out->payload, MAX_PAYLOAD_LEN, &out->payload_len) != 0) {
            return false;
        }
    } else if (plen_val) {
        /* Just a length indicator */
        uint64_t plen;
        if (parse_int_val(plen_val, &plen) == 0) {
            out->payload_len = (uint16_t)plen;
        }
    }

    return true;
}
