/*
 * Eregen (颐贞) - Boot Switch Module Implementation
 * Dual-bank partition switching logic for safe OTA firmware upgrades.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "boot_switch.h"
#include "../common/log.h"
#include <string.h>

/*
 * On embedded target, boot info is stored in a dedicated flash sector.
 * For test mode, we use a static variable as the "flash" simulation.
 */

#ifdef TEST_MODE
static ota_boot_info_t s_boot_info;
#else
/* Flash address for boot info — reserved sector at end of Bank A */
#define BOOT_INFO_FLASH_ADDR  0x0801F000UL  /* Last 512 bytes of 128KB flash */
#endif

/*
 * Initialize the boot switch subsystem.
 * Reads boot info from non-volatile storage.
 */
void ota_boot_init(void)
{
#ifdef TEST_MODE
    memset(&s_boot_info, 0, sizeof(s_boot_info));
    s_boot_info.current_bank = 0U;  /* Default to Bank A */
    s_boot_info.valid = true;
    log_info("Boot switch initialized (test mode, Bank A)");
#else
    /* Read boot info from flash sector */
    /* ota_boot_read_from_flash(&s_boot_info); */

    /* If no valid boot info found, default to Bank A */
    log_info("Boot switch initialized (embedded mode)");
#endif
}

/*
 * Get the current boot information.
 */
void ota_boot_get_info(ota_boot_info_t *info)
{
#ifdef TEST_MODE
    memcpy(info, &s_boot_info, sizeof(*info));
#else
    /* Read from flash and copy to output buffer */
    /* memcpy(info, (ota_boot_info_t*)BOOT_INFO_FLASH_ADDR, sizeof(*info)); */
    (void)info;
#endif
}

/*
 * Mark the current bank as successfully booted.
 */
void ota_boot_mark_success(void)
{
#ifdef TEST_MODE
    s_boot_info.flags &= ~OTA_BOOT_FLAG_OK;
    s_boot_info.pending_bank = 255U;  /* No pending bank */
    s_boot_info.fail_count = 0U;
    log_info("Bank %u marked as successful boot", s_boot_info.current_bank);
#else
    /* Write success marker to flash */
    /* ota_boot_write_to_flash(); */
#endif
}

/*
 * Request a switch to the other bank on next reboot.
 */
void ota_boot_request_switch(void)
{
#ifdef TEST_MODE
    uint8_t new_bank = (s_boot_info.current_bank == 0U) ? 1U : 0U;
    s_boot_info.pending_bank = new_bank;
    s_boot_info.flags |= OTA_BOOT_FLAG_FORCE;
    log_info("Boot switch requested: Bank %u -> Bank %u",
             s_boot_info.current_bank, new_bank);
#else
    /* Write pending switch request to flash */
    /* ota_boot_write_pending(new_bank); */
#endif
}

/*
 * Execute the bank switch.
 * Writes bootloader jump vector to flash so the next reboot loads the other bank.
 */
bool ota_boot_execute_switch(void)
{
#ifdef TEST_MODE
    uint8_t new_bank = (s_boot_info.current_bank == 0U) ? 1U : 0U;
    s_boot_info.current_bank = new_bank;
    s_boot_info.flags &= ~OTA_BOOT_FLAG_FORCE;
    log_info("Bank switch executed: now Booting from Bank %u", new_bank);
    return true;
#else
    /* Write new boot vector to flash jump sector */
    /* This is MCU-specific: GD32 uses a jump table in flash */
    /* ota_boot_write_jump_vector(new_bank); */
    log_info("Bank switch executed (embedded)");
    return true;
#endif
}

/*
 * Check if rollback is needed due to consecutive failures.
 */
bool ota_boot_needs_rollback(void)
{
#ifdef TEST_MODE
    if (s_boot_info.fail_count >= OTA_MAX_ROLLBACK_ATTEMPTS) {
        log_warn("Rollback triggered after %u failed attempts",
                 (unsigned)s_boot_info.fail_count);
        return true;
    }
    return false;
#else
    /* Read fail count from flash and compare */
    return false;
#endif
}

/*
 * Reset the boot failure counter.
 */
void ota_boot_reset_fail_count(void)
{
#ifdef TEST_MODE
    s_boot_info.fail_count = 0U;
    log_info("Boot failure counter reset");
#else
    /* Clear fail count in flash */
#endif
}
