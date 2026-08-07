/*
 * Eregen (颐贞) - Boot Switch Module Implementation
 * Dual-bank partition switching logic for safe OTA firmware upgrades.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "boot_switch.h"
#include "../common/log.h"
#include <string.h>

#ifndef TEST_MODE
/* CMSIS core headers for SCB->VTOR and NVIC_SystemReset */
#include "cmsis_core.h"
/* GD32E230 FMC (flash) peripheral — if available via standard library */
#ifdef HAVE_GD32E230_FMC
#include "gd32e230_fmc.h"
#endif
#endif

#ifdef TEST_MODE
static ota_boot_info_t s_boot_info;
#else
/* Boot info stored at end of Bank A flash (last 512 bytes) */
#define BOOT_INFO_FLASH_ADDR    0x0801F000UL

/* Page size for GD32E230: 1KB */
#define FLASH_PAGE_SIZE         1024U

/* Static boot info (stored in RAM, synced to flash on changes) */
static ota_boot_info_t s_boot_info;

/* Unlock FMC and read boot info from flash */
static void boot_flash_read(ota_boot_info_t *info)
{
    fmc_unlock();

    const ota_boot_info_t *flash_info = (const ota_boot_info_t*)BOOT_INFO_FLASH_ADDR;

    if (flash_info->valid && flash_info->flags == OTA_BANK_MAGIC_SUCCESS) {
        memcpy(info, flash_info, sizeof(*info));
        log_info("Boot switch loaded from flash: bank=%u, fail_count=%u",
                 info->current_bank, info->fail_count);
    } else {
        memset(info, 0, sizeof(*info));
        info->current_bank = 0U;
        info->pending_bank = 255U;
        info->valid = true;
        log_info("Boot switch: no valid flash data, defaulting to Bank A");
    }

    if (info->fail_count >= OTA_MAX_ROLLBACK_ATTEMPTS) {
        info->flags |= OTA_BOOT_FLAG_ROLLBACK;
        log_warn("Boot switch: rollback required after %u failures",
                 info->fail_count);
    }

    fmc_lock();
}

/* Write boot info to flash (erases page first) */
static void boot_flash_write(const ota_boot_info_t *info)
{
    fmc_unlock();

    /* Erase page containing boot info */
    fmc_page_erase(BOOT_INFO_FLASH_ADDR);

    /* Program word by word (32-bit writes) */
    const uint32_t *src = (const uint32_t*)info;
    uint32_t words = sizeof(ota_boot_info_t) / sizeof(uint32_t);

    for (uint32_t i = 0; i < words; i++) {
        fmc_word_program(BOOT_INFO_FLASH_ADDR + (i * sizeof(uint32_t)), src[i]);
    }

    fmc_lock();
    log_info("Boot info written to flash at 0x%08X", BOOT_INFO_FLASH_ADDR);
}
#endif

/*
 * Initialize the boot switch subsystem.
 */
void ota_boot_init(void)
{
#ifdef TEST_MODE
    memset(&s_boot_info, 0, sizeof(s_boot_info));
    s_boot_info.current_bank = 0U;
    s_boot_info.valid = true;
    log_info("Boot switch initialized (test mode, Bank A)");
#else
    boot_flash_read(&s_boot_info);
#endif
}

/*
 * Get the current boot information.
 */
void ota_boot_get_info(ota_boot_info_t *info)
{
    memcpy(info, &s_boot_info, sizeof(*info));
}

/*
 * Mark the current bank as successfully booted.
 */
void ota_boot_mark_success(void)
{
    s_boot_info.flags &= ~OTA_BOOT_FLAG_OK;
    s_boot_info.pending_bank = 255U;
    s_boot_info.fail_count = 0U;

#ifdef TEST_MODE
    log_info("Bank %u marked as successful boot", s_boot_info.current_bank);
#else
    boot_flash_write(&s_boot_info);
#endif
}

/*
 * Request a switch to the other bank on next reboot.
 */
void ota_boot_request_switch(void)
{
    uint8_t new_bank = (s_boot_info.current_bank == 0U) ? 1U : 0U;
    s_boot_info.pending_bank = new_bank;
    s_boot_info.flags |= OTA_BOOT_FLAG_FORCE;

#ifdef TEST_MODE
    log_info("Boot switch requested: Bank %u -> Bank %u",
             s_boot_info.current_bank, new_bank);
#else
    boot_flash_write(&s_boot_info);
#endif
}

/*
 * Execute the bank switch.
 */
bool ota_boot_execute_switch(void)
{
    uint8_t new_bank = (s_boot_info.current_bank == 0U) ? 1U : 0U;
    s_boot_info.current_bank = new_bank;
    s_boot_info.flags &= ~OTA_BOOT_FLAG_FORCE;

#ifdef TEST_MODE
    log_info("Bank switch executed: now booting from Bank %u", new_bank);
    return true;
#else
    /* Trigger system reset — bootloader will jump to new bank */
    log_info("Bank switch requested: resetting system...");

    /* Set the vector table offset to the other bank */
    /* Bank A: 0x08000000, Bank B: 0x08020000 (128KB apart) */
    SCB->VTOR = 0x08000000UL | ((uint32_t)new_bank << 17);

    /* Reset the system */
    NVIC_SystemReset();
    return false;  /* Never reached if reset succeeds */
#endif
}

/*
 * Check if rollback is needed due to consecutive failures.
 */
bool ota_boot_needs_rollback(void)
{
    if (s_boot_info.fail_count >= OTA_MAX_ROLLBACK_ATTEMPTS) {
        log_warn("Rollback triggered after %u failed attempts",
                 s_boot_info.fail_count);
        return true;
    }
    return false;
}

/*
 * Increment boot failure counter and persist.
 */
void ota_boot_record_failure(void)
{
    s_boot_info.fail_count++;

#ifdef TEST_MODE
    log_warn("Boot failure count: %u/%u",
             s_boot_info.fail_count, OTA_MAX_ROLLBACK_ATTEMPTS);
#else
    boot_flash_write(&s_boot_info);
#endif
}

/*
 * Reset the boot failure counter.
 */
void ota_boot_reset_fail_count(void)
{
    s_boot_info.fail_count = 0U;

#ifdef TEST_MODE
    log_info("Boot failure counter reset");
#else
    boot_flash_write(&s_boot_info);
#endif
}
