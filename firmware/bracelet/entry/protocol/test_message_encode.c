/*
 * Eregen (颐贞) - Message Encode/Decode Test Harness
 * Host-compiled test driver for message_encode and message_decode modules.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <string.h>
#include <stdbool.h>
#include "../protocol/message_encode.h"
#include "../protocol/message_decode.h"

static int passed = 0;
static int failed = 0;

static void check(bool cond, const char *label)
{
    if (cond) {
        printf("  PASS: %s\n", label);
        passed++;
    } else {
        printf("  FAIL: %s\n", label);
        failed++;
    }
}

int main(void)
{
    printf("Message encode/decode tests:\n");

    /* Test 1: Encode a heartbeat message */
    eregen_msg_t msg;
    memset(&msg, 0, sizeof(msg));
    msg.type = MSG_HEARTBEAT;
    strncpy(msg.dev_id, "BR-1234", sizeof(msg.dev_id));
    msg.timestamp = 1720000000ULL;
    msg.payload_len = 0;

    uint8_t encoded[MAX_MSG_LEN];
    int enc_result = message_encode(&msg, encoded, sizeof(encoded));
    check(enc_result > 0, "encode heartbeat succeeds");

    /* Test 2: Decode back */
    eregen_msg_t decoded;
    bool dec_result = message_decode(encoded, (uint16_t)enc_result, &decoded);
    check(dec_result, "decode heartbeat succeeds");
    check(decoded.type == MSG_HEARTBEAT, "decoded type matches");
    check(strcmp(decoded.dev_id, "BR-1234") == 0, "decoded dev_id matches");
    check(decoded.timestamp == 1720000000ULL, "decoded timestamp matches");

    /* Test 3: Encode health message with payload */
    memset(&msg, 0, sizeof(msg));
    msg.type = MSG_HEALTH;
    strncpy(msg.dev_id, "BR-5678", sizeof(msg.dev_id));
    msg.timestamp = 1720000100ULL;
    msg.payload[0] = 72;   /* HR = 72 */
    msg.payload[1] = 98;   /* SpO2 = 98 */
    msg.payload_len = 2;

    enc_result = message_encode(&msg, encoded, sizeof(encoded));
    check(enc_result > 0, "encode health with payload succeeds");

    dec_result = message_decode(encoded, (uint16_t)enc_result, &decoded);
    check(dec_result, "decode health with payload succeeds");
    check(decoded.type == MSG_HEALTH, "decoded health type matches");
    check(decoded.payload_len == 2, "decoded payload length matches");
    check(decoded.payload[0] == 72, "decoded HR matches");
    check(decoded.payload[1] == 98, "decoded SpO2 matches");

    /* Test 4: Invalid inputs */
    check(message_encode(NULL, encoded, sizeof(encoded)) < 0, "NULL msg rejected");
    check(message_encode(&msg, NULL, sizeof(encoded)) < 0, "NULL out rejected");
    check(message_encode(&msg, encoded, 0) < 0, "zero-length out rejected");

    /* Test 5: Bad message type */
    msg.type = (msg_type_t)99;
    check(message_encode(&msg, encoded, sizeof(encoded)) < 0, "invalid type rejected");

    /* Test 6: Corrupted data fails CRC check */
    encoded[5] ^= 0xFF;  /* Flip bits in JSON data */
    check(!message_decode(encoded, (uint16_t)enc_result, &decoded), "corrupted data fails CRC");

    /* Test 7: Buffer too small */
    uint8_t small_buf[4];
    check(message_encode(&msg, small_buf, 4) < 0, "small output buffer rejected");

    /* Test 8: Too-large payload rejected */
    memset(&msg, 0, sizeof(msg));
    msg.type = MSG_LOCATION;
    strncpy(msg.dev_id, "BR-9999", sizeof(msg.dev_id));
    msg.payload_len = MAX_PAYLOAD_LEN + 1;
    check(message_encode(&msg, encoded, sizeof(encoded)) < 0, "oversized payload rejected");

    /* Test 9: Valid device ID format accepted */
    memset(&msg, 0, sizeof(msg));
    msg.type = MSG_SOS;
    strncpy(msg.dev_id, "BR-9999", sizeof(msg.dev_id));
    check(message_encode(&msg, encoded, sizeof(encoded)) > 0, "valid dev_id accepted");

    /* Test 10: Decode null/short inputs */
    check(!message_decode(NULL, 10, &decoded), "NULL input rejected");
    check(!message_decode(encoded, 1, &decoded), "too-short input rejected");

    printf("\nResults: %d passed, %d failed\n", passed, failed);
    return failed > 0 ? 1 : 0;
}
