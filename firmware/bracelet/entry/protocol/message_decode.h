/*
 * Eregen (颐贞) - Message Decode Header
 * Decodes wire format (JSON + CRC16) into eregen_msg_t.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef MESSAGE_DECODE_H
#define MESSAGE_DECODE_H

#include "message_encode.h"
#include <stdbool.h>

/**
 * Decode a message from wire format into eregen_msg_t.
 * Performs CRC16 verification before parsing.
 * @param in   Input buffer containing encoded message.
 * @param in_len Length of input buffer.
 * @param out  Pointer to message structure to fill.
 * @return true if decode and CRC check succeed, false otherwise.
 */
bool message_decode(const uint8_t *in, uint16_t in_len, eregen_msg_t *out);

#endif /* MESSAGE_DECODE_H */
