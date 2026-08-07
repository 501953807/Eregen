/*
 * Eregen (颐贞) - OTA Update Orchestration Header
 * State machine for complete firmware upgrade flow:
 *   download → verify → write to flash bank B → switch → reboot
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef OTA_UPDATE_H
#define OTA_UPDATE_H

#include <stdint.h>
#include <stdbool.h>

/** OTA update states */
typedef enum {
    OTA_STATE_IDLE = 0,       /* No update in progress */
    OTA_STATE_DOWNLOADING,    /* Downloading firmware from server */
    OTA_STATE_VERIFYING,      /* Verifying SHA256 hash */
    OTA_STATE_WRITING,        /* Writing to flash bank B */
    OTA_STATE_SWITCHING,      /* Requesting bank switch */
    OTA_STATE_COMPLETE,       /* Update finished successfully */
    OTA_STATE_ERROR           /* Update failed */
} ota_state_t;

/** Result codes for OTA operations */
typedef enum {
    OTA_OK = 0,
    OTA_ERR_CONNECT,          /* Network connection failed */
    OTA_ERR_DOWNLOAD,         /* Download incomplete/corrupt */
    OTA_ERR_VERIFY,           /* SHA256 mismatch */
    OTA_ERR_FLASH,            /* Flash write/erase failed */
    OTA_ERR_SWITCH,           /* Bank switch failed */
    OTA_ERR_TIMEOUT,          /* Operation timed out */
    OTA_ERR_ABORTED           /* Update aborted by caller */
} ota_result_t;

/** Callback invoked during OTA progress updates */
typedef void (*ota_progress_cb_t)(ota_state_t state, uint32_t progress, uint32_t total);

/**
 * Initialize the OTA update subsystem.
 * Must be called once before any OTA operation.
 */
void ota_update_init(void);

/**
 * Start an OTA firmware update.
 * Downloads firmware from the given URL, verifies it, and writes to bank B.
 * @param url       Full HTTP URL of the .bin firmware file.
 * @param sig       Expected firmware signature (SHA256 + metadata).
 * @param cb        Progress callback (may be NULL).
 * @return OTA_OK on success (update proceeds asynchronously), or error code.
 */
ota_result_t ota_update_start(const char *url, const void *sig, ota_progress_cb_t cb);

/**
 * Check current OTA update state.
 * @return Current state.
 */
ota_state_t ota_update_get_state(void);

/**
 * Get the last OTA update result.
 * @return Result code, or OTA_OK if no error.
 */
ota_result_t ota_update_get_result(void);

/**
 * Abort an in-progress OTA update.
 */
void ota_update_abort(void);

#endif /* OTA_UPDATE_H */
