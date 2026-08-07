/*
 * Eregen Medical Wristband - NFC Server Header
 * NFC A protocol (106 kbps) for nurse terminal verification
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef NFC_SERVER_H
#define NFC_SERVER_H

#include "esp_err.h"
#include <stdint.h>
#include <stdbool.h>

#define NFC_MAX_DATA_LEN  256
#define NFC_TIMEOUT_MS    5000

typedef enum {
    NFC_EVENT_TAG_READ = 0,
    NFC_EVENT_TAG_WRITE,
    NFC_EVENT_TAG_LOST,
} nfc_event_t;

typedef struct {
    uint8_t uid[7];
    uint8_t uid_len;
    uint8_t atqa[2];
    uint8_t sak;
    uint32_t tag_type;
} nfc_tag_info_t;

typedef void (*nfc_callback_t)(nfc_event_t event, const nfc_tag_info_t *tag_info, void *user_data);

esp_err_t nfc_server_init(nfc_callback_t callback, void *user_data);
esp_err_t nfc_server_start(void);
void nfc_server_stop(void);
esp_err_t nfc_server_read_tag(nfc_tag_info_t *tag_info);
esp_err_t nfc_server_write_tag(const nfc_tag_info_t *tag_info, const uint8_t *data, size_t len);
void nfc_server_set_callback(nfc_callback_t callback, void *user_data);

#endif /* NFC_SERVER_H */
