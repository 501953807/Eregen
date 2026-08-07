/*
 * Eregen Medical Wristband - Security Module Header
 * AES-128-CBC encryption for data protection
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef SECURITY_H
#define SECURITY_H

#include "esp_err.h"
#include <stdint.h>
#include <stddef.h>

#define AES_BLOCK_SIZE 16

typedef struct {
    const uint8_t *key;
    size_t key_len;
    uint8_t iv[AES_BLOCK_SIZE];
} security_ctx_t;

esp_err_t security_init(security_ctx_t *ctx);
esp_err_t security_encrypt(const uint8_t *plaintext, size_t plaintext_len,
                           uint8_t *ciphertext, size_t *ciphertext_len);
esp_err_t security_decrypt(const uint8_t *ciphertext, size_t ciphertext_len,
                           uint8_t *plaintext, size_t *plaintext_len);
uint32_t security_compute_checksum(const uint8_t *data, size_t len);

#endif /* SECURITY_H */
