/*
 * Eregen (颐贞) - CRC16-CCITT Utility Header
 * Standard CRC16-CCITT (poly=0x1021, init=0xFFFF)
 * Host-testable via #ifdef TEST_MODE guard.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef CRC16_H
#define CRC16_H

#include <stdint.h>

/**
 * Calculate CRC16-CCITT over a byte buffer.
 * @param data Pointer to input data.
 * @param len  Number of bytes to process.
 * @return CRC16 checksum.
 */
uint16_t crc16_calc(const uint8_t *data, uint16_t len);

#endif /* CRC16_H */
