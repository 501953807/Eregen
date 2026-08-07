/*
 * Eregen (颐贞) - Boot Switch Module Header
 * Dual-bank partition switching logic for safe OTA firmware upgrades.
 * Bank A = current running firmware, Bank B = pending update.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef OTA_BOOT_SWITCH_H
#define OTA_BOOT_SWITCH_H

#include <stdint.h>
#include <stdbool.h>

/** Maximum number of failed upgrade attempts before reverting to original bank */
#define OTA_MAX_ROLLBACK_ATTEMPTS 3U

/** Magic value indicating a successful boot from this bank */
#define OTA_BANK_MAGIC_SUCCESS   0x4552474EU  /* "EREG" */

/** Boot flags */
typedef enum {
    OTA_BOOT_FLAG_OK       = (1U << 0),  /* Previous boot was successful */
    OTA_BOOT_FLAG_ROLLBACK = (1U << 1),  /* Must rollback to other bank */
    OTA_BOOT_FLAG_FORCE    = (1U << 2)   /* Force switch to other bank */
} ota_boot_flags_t;

/**
 * Current boot bank status.
 */
typedef struct {
    uint8_t  current_bank;      /* 0 = Bank A (production), 1 = Bank B (update) */
    uint8_t  pending_bank;      /* Bank that needs confirmation */
    uint8_t  fail_count;        /* Number of consecutive failed boots */
    uint32_t flags;             /* Boot flags (OTA_BOOT_FLAG_*) */
    bool     valid;             /* True if boot info is valid */
} ota_boot_info_t;

/**
 * Initialize the boot switch subsystem.
 * Reads boot info from non-volatile storage (flash sector).
 */
void ota_boot_init(void);

/**
 * Get the current boot information.
 * @param info Pointer to boot info structure (caller-allocated).
 */
void ota_boot_get_info(ota_boot_info_t *info);

/**
 * Mark the current bank as successfully booted.
 * Clears the pending flag so the other bank won't be rolled back.
 */
void ota_boot_mark_success(void);

/**
 * Request a switch to the other bank on next reboot.
 * Sets the pending bank and increments fail count if already in pending state.
 */
void ota_boot_request_switch(void);

/**
 * Execute the bank switch.
 * Writes bootloader jump vector to flash so the next reboot loads the other bank.
 * @return true if switch requested successfully.
 */
bool ota_boot_execute_switch(void);

/**
 * Check if rollback is needed.
 * @return true if we should boot from the other bank due to failures.
 */
bool ota_boot_needs_rollback(void);

/**
 * Reset the boot failure counter.
 * Called after a successful boot to prevent unnecessary rollbacks.
 */
void ota_boot_reset_fail_count(void);

#endif /* OTA_BOOT_SWITCH_H */
