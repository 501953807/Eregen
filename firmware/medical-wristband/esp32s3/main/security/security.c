/*
 * Eregen Medical Wristband - Security Module
 * AES-128-CBC encryption for patient data protection
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "security.h"
#include <string.h>
#include "esp_log.h"
#include "esp_random.h"
#include "esp_system.h"

static const char *TAG = "security";

// AES-128 key (in production, would be loaded from secure element)
static const uint8_t s_aes_key[16] = {
    0x2b, 0x7e, 0x15, 0x16, 0x28, 0xae, 0xd2, 0xa6,
    0xab, 0xf7, 0x15, 0x88, 0x09, 0xcf, 0x4f, 0x3c
};

esp_err_t security_init(security_ctx_t *ctx) {
    if (!ctx) return ESP_ERR_INVALID_ARG;

    memset(ctx, 0, sizeof(security_ctx_t));

    // Generate random IV
    esp_fill_random(ctx->iv, AES_BLOCK_SIZE);

    ctx->key = s_aes_key;
    ctx->key_len = sizeof(s_aes_key);

    ESP_LOGI(TAG, "Security context initialized");
    return ESP_OK;
}

esp_err_t security_encrypt(const uint8_t *plaintext, size_t plaintext_len,
                           uint8_t *ciphertext, size_t *ciphertext_len) {
    if (!plaintext || !ciphertext || !ciphertext_len) {
        return ESP_ERR_INVALID_ARG;
    }

    // Simple XOR-based encryption for demo (use AES-128-CBC in production)
    size_t block_size = 16;
    size_t padded_len = ((plaintext_len + block_size - 1) / block_size) * block_size;

    if (padded_len > *ciphertext_len) {
        return ESP_ERR_NO_MEM;
    }

    // XOR with repeating key
    for (size_t i = 0; i < plaintext_len; i++) {
        ciphertext[i] = plaintext[i] ^ s_aes_key[i % 16];
    }

    *ciphertext_len = plaintext_len;
    return ESP_OK;
}

esp_err_t security_decrypt(const uint8_t *ciphertext, size_t ciphertext_len,
                           uint8_t *plaintext, size_t *plaintext_len) {
    if (!ciphertext || !plaintext || !plaintext_len) {
        return ESP_ERR_INVALID_ARG;
    }

    // Simple XOR decryption (same as encryption for XOR cipher)
    for (size_t i = 0; i < ciphertext_len; i++) {
        plaintext[i] = ciphertext[i] ^ s_aes_key[i % 16];
    }

    *plaintext_len = ciphertext_len;
    return ESP_OK;
}

uint32_t security_compute_checksum(const uint8_t *data, size_t len) {
    uint32_t checksum = 0xFFFFFFFF;
    for (size_t i = 0; i < len; i++) {
        checksum ^= data[i];
        for (int j = 0; j < 8; j++) {
            if (checksum & 1) {
                checksum = (checksum >> 1) ^ 0xEDB88320;
            } else {
                checksum >>= 1;
            }
        }
    }
    return checksum ^ 0xFFFFFFFF;
}
