/*
 * Eregen (颐贞) - CRC16 Test Harness
 * Host-compiled test driver for crc16 module.
 * Uses an independent bit-by-bit CRC implementation as reference.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <string.h>
#include "../common/crc16.h"

static int passed = 0;
static int failed = 0;

static void check(const char *label, uint16_t actual, uint16_t expected)
{
    if (actual == expected) {
        printf("  PASS: %s\n", label);
        passed++;
    } else {
        printf("  FAIL: %s (got 0x%04X, expected 0x%04X)\n", label, actual, expected);
        failed++;
    }
}

/* Independent reference implementation for verification */
static uint16_t crc16_ref(const uint8_t *data, uint16_t len)
{
    if (!data || len == 0) return 0;
    uint16_t crc = 0xFFFF;
    for (uint16_t i = 0; i < len; i++) {
        crc ^= ((uint16_t)data[i] << 8);
        for (uint8_t j = 0; j < 8; j++) {
            if (crc & 0x8000) {
                crc = (crc << 1) ^ 0x1021;
            } else {
                crc <<= 1;
            }
        }
    }
    return crc;
}

int main(void)
{
    printf("CRC16-CCITT tests:\n");

    /* Empty input returns 0 */
    check("empty input", crc16_calc(NULL, 0), 0);

    /* Known test vector: "123456789" -> 0x29B1 (canonical CRC16-CCITT) */
    const uint8_t vec[] = "123456789";
    check("standard vector", crc16_calc(vec, sizeof(vec) - 1), 0x29B1);

    /* Cross-check against independent reference implementation */
    uint8_t test_data[] = "Eregen platform v1.0";
    check("cross-check: platform string",
          crc16_calc(test_data, strlen((char*)test_data)),
          crc16_ref(test_data, strlen((char*)test_data)));

    uint8_t single_byte = 0x42;
    check("cross-check: single 0x42",
          crc16_calc(&single_byte, 1),
          crc16_ref(&single_byte, 1));

    uint8_t zeros[256];
    memset(zeros, 0, sizeof(zeros));
    check("cross-check: 256 zero bytes",
          crc16_calc(zeros, 256),
          crc16_ref(zeros, 256));

    uint8_t ones[100];
    memset(ones, 0xFF, sizeof(ones));
    check("cross-check: 100 0xFF bytes",
          crc16_calc(ones, 100),
          crc16_ref(ones, 100));

    /* Device ID format */
    const char dev_id[] = "BR-1234";
    check("cross-check: device ID",
          crc16_calc((const uint8_t *)dev_id, strlen(dev_id)),
          crc16_ref((const uint8_t *)dev_id, strlen(dev_id)));

    /* Determinism */
    check("deterministic", crc16_calc(vec, 5), crc16_calc(vec, 5));

    /* Different inputs produce different CRCs */
    uint8_t a = 0x00, b = 0x01;
    check("different bytes produce different CRCs",
          crc16_calc(&a, 1) != crc16_calc(&b, 1), 1);

    /* Non-trivial: repeated block differs from single block */
    uint8_t block[10];
    memset(block, 0xAA, sizeof(block));
    uint8_t block2[20];
    memcpy(block2, block, sizeof(block));
    memcpy(block2 + 10, block, sizeof(block));
    check("repeated block has different CRC",
          crc16_calc(block, 10) != crc16_calc(block2, 20), 1);

    printf("\nResults: %d passed, %d failed\n", passed, failed);
    return failed > 0 ? 1 : 0;
}
