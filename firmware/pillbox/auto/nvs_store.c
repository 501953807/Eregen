/*
 * Eregen (颐贞) - NVS Storage Implementation
 * File-based mock for TEST_MODE, native ESP32-C3 NVS for production.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "nvs_store.h"

#include <string.h>
#include <stdio.h>

#ifdef TEST_MODE
#define ESP_LOGI(...) printf(__VA_ARGS__)
#define ESP_LOGW(...) fprintf(stderr, __VA_ARGS__)
#define ESP_LOGE(...) fprintf(stderr, __VA_ARGS__)
#else
#include "esp_log.h"
#include "nvs_flash.h"
#endif

static const char *TAG = "nvs";

/* Mock NVS file path (in TEST_MODE) */
#define MOCK_NVS_FILE  "/tmp/eregen_nvs_rules.dat"

#ifdef TEST_MODE
/* Mock state — whether the mock file exists */
static bool s_nvs_initialized = false;
#endif

void nvs_init(void)
{
#ifdef TEST_MODE
    /* In test mode, initialize is a no-op — files are created on demand */
    s_nvs_initialized = true;
#else
    /* Initialize NVS flash if not already done */
    esp_err_t ret = nvs_flash_init();
    if (ret == ESP_ERR_NVS_NO_FREE_PAGES ||
        ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        nvs_flash_erase();
        ret = nvs_flash_init();
    }
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS flash init failed: %d", ret);
    }
#endif
}

bool nvs_save_rules(const med_rule_t *rules, uint8_t count)
{
    if (!rules || count == 0) return false;

#ifdef TEST_MODE
    FILE *f = fopen(MOCK_NVS_FILE, "wb");
    if (!f) {
        ESP_LOGE(TAG, "Failed to open mock NVS file for writing");
        return false;
    }

    /* Write count first */
    fwrite(&count, sizeof(uint8_t), 1, f);

    /* Write rules blob */
    fwrite(rules, sizeof(med_rule_t), count, f);
    fclose(f);

    ESP_LOGI(TAG, "Saved %d rules to mock NVS (%s)", count, MOCK_NVS_FILE);
    return true;
#else
    nvs_handle_t handle;
    esp_err_t ret = nvs_open("pillbox", NVS_READWRITE, &handle);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS open failed: %d", ret);
        return false;
    }

    ret = nvs_set_u8(handle, "rule_count", count);
    if (ret != ESP_OK) {
        nvs_close(handle);
        return false;
    }

    ret = nvs_set_blob(handle, "med_rules", rules,
                       count * sizeof(med_rule_t));
    if (ret != ESP_OK) {
        nvs_close(handle);
        return false;
    }

    ret = nvs_commit(handle);
    nvs_close(handle);

    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS commit failed: %d", ret);
        return false;
    }

    ESP_LOGI(TAG, "Saved %d rules to NVS", count);
    return true;
#endif
}

bool nvs_load_rules(med_rule_t *rules, uint8_t *count)
{
    if (!rules || !count) return false;

    memset(rules, 0, MED_RULE_PARSER_MAX_RULES * sizeof(med_rule_t));

#ifdef TEST_MODE
    FILE *f = fopen(MOCK_NVS_FILE, "rb");
    if (!f) {
        ESP_LOGW(TAG, "Mock NVS file not found: %s", MOCK_NVS_FILE);
        return false;
    }

    /* Read count */
    size_t read_count = fread(count, sizeof(uint8_t), 1, f);
    if (read_count != 1 || *count == 0 || *count > MED_RULE_PARSER_MAX_RULES) {
        fclose(f);
        return false;
    }

    /* Read rules */
    size_t items_read = fread(rules, sizeof(med_rule_t), *count, f);
    fclose(f);

    if ((uint8_t)items_read != *count) {
        return false;
    }

    ESP_LOGI(TAG, "Loaded %d rules from mock NVS", *count);
    return true;
#else
    nvs_handle_t handle;
    esp_err_t ret = nvs_open("pillbox", NVS_READONLY, &handle);
    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "NVS open failed: %d", ret);
        return false;
    }

    uint8_t loaded_count = 0;
    ret = nvs_get_u8(handle, "rule_count", &loaded_count);
    if (ret != ESP_OK || loaded_count == 0 ||
        loaded_count > MED_RULE_PARSER_MAX_RULES) {
        nvs_close(handle);
        return false;
    }

    size_t needed = loaded_count * sizeof(med_rule_t);
    ret = nvs_get_blob(handle, "med_rules", rules, &needed);
    nvs_close(handle);

    if (ret != ESP_OK) {
        return false;
    }

    *count = loaded_count;
    ESP_LOGI(TAG, "Loaded %d rules from NVS", *count);
    return true;
#endif
}
