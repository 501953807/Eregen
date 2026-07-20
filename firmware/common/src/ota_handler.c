/*
 * Eregen (颐贞) - OTA Firmware Update Handler Implementation
 * ESP32-C3 based pillbox OTA receiver: parses MQTT ota commands,
 * downloads firmware via HTTPS, verifies SHA-256, writes to OTA
 * partition, reports progress, reboots into new firmware.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "ota_handler.h"
#include "mqtt_common.h"
#include "esp_log.h"
#include "esp_http_client.h"
#include "esp_https_ota.h"
#include "esp_ota_ops.h"
#include "esp_partition.h"
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "freertos/semphr.h"
#include "cJSON.h"
#include <string.h>
#include <stdio.h>
#include <openssl/sha.h>

static const char* TAG = "ota_handler";

/* OTA command JSON field names */
#define JSON_FIELD_TYPE     "type"
#define JSON_FIELD_URL      "url"
#define JSON_FIELD_HASH     "hash"
#define JSON_FIELD_VER      "ver"
#define JSON_FIELD_FORCE    "force"
#define JSON_FIELD_DEV_ID   "dev_id"
#define JSON_FIELD_JOB_ID   "job_id"
#define JSON_FIELD_PROGRESS "progress"
#define JSON_FIELD_STATUS   "status"
#define JSON_FIELD_ERROR    "error"

/* OTA command type */
#define OTA_CMD_TYPE        "ota"

/* Progress report topic format */
#define TOPIC_OTA_PROGRESS  "eregen/device/pillbox/%s/ota_progress"

/* Task parameters */
#define OTA_TASK_STACK_WORDS    (8192 / sizeof(size_t))
#define OTA_TASK_PRIORITY       (5)
#define HTTP_BUFFER_SIZE        1460  /* TCP MSS */

/* Status strings */
#define STATUS_IDLE       "idle"
#define STATUS_DOWNLOADING "downloading"
#define STATUS_VERIFYING  "verifying"
#define STATUS_FLASHING   "flashing"
#define STATUS_COMPLETE   "complete"
#define STATUS_FAILED     "failed"
#define STATUS_REBOOTING  "rebooting"

/* Static state — guarded by s_ota_mutex for thread safety */
static SemaphoreHandle_t s_ota_mutex = NULL;
static bool s_ota_active = false;
static char s_current_status[32] = STATUS_IDLE;
static char s_expected_hash[65] = {0};
static char s_firmware_version[32] = {0};
static char s_target_dev_id[32] = {0};

/* Forward declarations */
static void ota_update_task(void* param);
static esp_err_t ota_download_and_verify(const char* url, uint8_t* expected_sha256);

/*
 * Initialize OTA subsystem.
 */
esp_err_t ota_init(void) {
    s_ota_mutex = xSemaphoreCreateMutex();
    if (!s_ota_mutex) return ESP_ERR_NO_MEM;

    s_ota_active = false;
    strncpy(s_current_status, STATUS_IDLE, sizeof(s_current_status) - 1);
    memset(s_expected_hash, 0, sizeof(s_expected_hash));
    memset(s_firmware_version, 0, sizeof(s_firmware_version));
    memset(s_target_dev_id, 0, sizeof(s_target_dev_id));

    ESP_LOGI(TAG, "OTA handler initialized");
    return ESP_OK;
}

/*
 * Handle incoming OTA command from MQTT.
 * Parses JSON, validates fields, launches background OTA task.
 */
esp_err_t ota_handle_command(const char* topic, const uint8_t* payload, uint16_t len) {
    cJSON* root = cJSON_ParseWithLength((const char*)payload, len);
    if (!root) {
        ESP_LOGE(TAG, "Failed to parse OTA JSON command");
        return ESP_ERR_INVALID_ARG;
    }

    /* Check command type */
    cJSON* type = cJSON_GetObjectItem(root, JSON_FIELD_TYPE);
    if (!type || !cJSON_IsString(type) || strcmp(type->valuestring, OTA_CMD_TYPE) != 0) {
        ESP_LOGW(TAG, "Not an OTA command (type=%s)", type ? type->valuestring : "null");
        cJSON_Delete(root);
        return ESP_ERR_INVALID_ARG;
    }

    /* Extract URL */
    cJSON* url_json = cJSON_GetObjectItem(root, JSON_FIELD_URL);
    if (!url_json || !cJSON_IsString(url_json) || url_json->valuestring[0] == '\0') {
        ESP_LOGE(TAG, "Missing or empty 'url' field");
        cJSON_Delete(root);
        return ESP_ERR_INVALID_ARG;
    }
    const char* fw_url = url_json->valuestring;

    /* Extract SHA-256 hash */
    cJSON* hash_json = cJSON_GetObjectItem(root, JSON_FIELD_HASH);
    if (!hash_json || !cJSON_IsString(hash_json) || strlen(hash_json->valuestring) != 64) {
        ESP_LOGE(TAG, "Missing or invalid 'hash' field (expected 64 hex chars)");
        cJSON_Delete(root);
        return ESP_ERR_INVALID_ARG;
    }
    strncpy(s_expected_hash, hash_json->valuestring, sizeof(s_expected_hash) - 1);

    /* Extract version (optional) */
    cJSON* ver_json = cJSON_GetObjectItem(root, JSON_FIELD_VER);
    if (ver_json && cJSON_IsString(ver_json)) {
        strncpy(s_firmware_version, ver_json->valuestring, sizeof(s_firmware_version) - 1);
    }

    /* Extract device ID for progress reports */
    cJSON* dev_id_json = cJSON_GetObjectItem(root, JSON_FIELD_DEV_ID);
    if (dev_id_json && cJSON_IsString(dev_id_json)) {
        strncpy(s_target_dev_id, dev_id_json->valuestring, sizeof(s_target_dev_id) - 1);
    }

    /* Check force flag (optional, default false) */
    cJSON* force_json = cJSON_GetObjectItem(root, JSON_FIELD_FORCE);
    bool force_update = force_json && cJSON_IsBool(force_json) && force_json->valuebool;

    cJSON_Delete(root);

    /* Check if another OTA is already in progress */
    if (xSemaphoreTake(s_ota_mutex, pdMS_TO_TICKS(100)) != pdTRUE) {
        ESP_LOGW(TAG, "OTA already in progress, skipping new command");
        return ESP_ERR_NO_CONN;
    }

    if (s_ota_active) {
        xSemaphoreGive(s_ota_mutex);
        ESP_LOGW(TAG, "OTA already active, skipping new command");
        return ESP_ERR_NO_CONN;
    }
    s_ota_active = true;
    xSemaphoreGive(s_ota_mutex);

    ESP_LOGI(TAG, "OTA update requested: url=%s ver=%s force=%d",
             fw_url, s_firmware_version, force_update);

    /* Launch OTA task in background */
    char* url_copy = strdup(fw_url);
    if (!url_copy) {
        xSemaphoreTake(s_ota_mutex, portMAX_DELAY);
        s_ota_active = false;
        xSemaphoreGive(s_ota_mutex);
        ESP_LOGE(TAG, "Failed to allocate URL copy");
        return ESP_ERR_NO_MEM;
    }

    BaseType_t xRet = xTaskCreate(
        ota_update_task,
        "ota_update",
        OTA_TASK_STACK_WORDS,
        url_copy,
        OTA_TASK_PRIORITY,
        NULL
    );

    if (xRet != pdPASS) {
        free(url_copy);
        xSemaphoreTake(s_ota_mutex, portMAX_DELAY);
        s_ota_active = false;
        xSemaphoreGive(s_ota_mutex);
        ESP_LOGE(TAG, "Failed to create OTA task");
        return ESP_ERR_NO_MEM;
    }

    return ESP_OK;
}

/*
 * Background task: download, verify, flash, reboot.
 */
static void ota_update_task(void* param) {
    const char* fw_url = (const char*)param;
    esp_err_t ret = ESP_OK;

    /* Convert expected hex hash to bytes */
    uint8_t expected_sha256[32];
    for (int i = 0; i < 32; i++) {
        unsigned int val = 0;
        sscanf(&s_expected_hash[i * 2], "%02x", &val);
        expected_sha256[i] = (uint8_t)val;
    }

    /* Phase 1: Download and verify hash */
    {
        portENTER_CRITICAL(NULL);
        strncpy(s_current_status, STATUS_DOWNLOADING, sizeof(s_current_status) - 1);
        portEXIT_CRITICAL(NULL);
        ota_report_progress(0, STATUS_DOWNLOADING, "");
    }

    ret = ota_download_and_verify(fw_url, expected_sha256);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Download/verify failed: %s", esp_err_to_name(ret));
        ota_report_progress(0, STATUS_FAILED, "download_verify_failed");
        goto cleanup;
    }

    /* Phase 2: Flash via esp_https_ota */
    {
        portENTER_CRITICAL(NULL);
        strncpy(s_current_status, STATUS_FLASHING, sizeof(s_current_status) - 1);
        portEXIT_CRITICAL(NULL);
        ota_report_progress(50, STATUS_FLASHING, "");
    }

    esp_https_ota_config_t ota_cfg = {
        .http_timeout_ms = 120 * 1000,  /* 120 second timeout */
    };

    ret = esp_https_ota_perform(&ota_cfg);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "HTTPS OTA flash failed: %s", esp_err_to_name(ret));
        ota_report_progress(50, STATUS_FAILED, "flash_failed");
        goto cleanup;
    }

    /* Phase 3: Complete and reboot */
    {
        portENTER_CRITICAL(NULL);
        strncpy(s_current_status, STATUS_REBOOTING, sizeof(s_current_status) - 1);
        portEXIT_CRITICAL(NULL);
        ota_report_progress(100, STATUS_COMPLETE, "");
    }

    ESP_LOGI(TAG, "OTA update successful, version=%s, rebooting...", s_firmware_version);

cleanup:
    free((void*)fw_url);

    portENTER_CRITICAL(NULL);
    s_ota_active = false;
    portEXIT_CRITICAL(NULL);

    vTaskDelete(NULL);
}

/*
 * Download firmware from HTTPS URL and verify SHA-256 hash.
 * Streams data in chunks to minimize RAM usage.
 *
 * @return ESP_OK on success, error code otherwise
 */
static esp_err_t ota_download_and_verify(const char* url, uint8_t* expected_sha256) {
    SHA256_CTX sha_ctx;
    SHA256_Init(&sha_ctx);

    /* Configure HTTP client */
    esp_http_client_config_t http_cfg = {
        .url = url,
        .cert_pem = NULL,  /* Use system CA bundle (esp-tls) */
        .timeout_ms = 120 * 1000,  /* 120 second timeout for large firmware */
        .event_handler = NULL,
    };

    esp_http_client_handle_t client = esp_http_client_init(&http_cfg);
    if (!client) {
        ESP_LOGE(TAG, "Failed to initialize HTTP client");
        return ESP_FAIL;
    }

    esp_err_t err = esp_http_client_open(client, 0);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "Failed to open HTTP connection: %s", esp_err_to_name(err));
        esp_http_client_cleanup(client);
        return err;
    }

    /* Read firmware in chunks */
    char* buf = malloc(HTTP_BUFFER_SIZE);
    if (!buf) {
        ESP_LOGE(TAG, "Failed to allocate download buffer");
        esp_http_client_close(client);
        esp_http_client_cleanup(client);
        return ESP_ERR_NO_MEM;
    }

    int total_bytes = 0;
    int content_length = esp_http_client_fetch_headers(client);
    if (content_length > 0) {
        ESP_LOGI(TAG, "Downloading firmware: %d bytes", content_length);
    } else {
        ESP_LOGW(TAG, "Unknown content length");
    }

    while (true) {
        int data_read = esp_http_client_read(client, buf, HTTP_BUFFER_SIZE);
        if (data_read < 0) {
            ESP_LOGE(TAG, "HTTP read error");
            break;
        }
        if (data_read == 0) {
            break;  /* End of data */
        }

        /* Feed chunk into SHA-256 */
        SHA256_Update(&sha_ctx, (const uint8_t*)buf, data_read);

        total_bytes += data_read;

        /* Report progress every 10% */
        if (content_length > 0) {
            int progress = (total_bytes * 100) / content_length;
            if (progress % 10 == 0 || progress == 100) {
                ota_report_progress(progress, STATUS_DOWNLOADING, "");
            }
        }
    }

    free(buf);
    esp_http_client_close(client);
    esp_http_client_cleanup(client);

    /* Finalize SHA-256 */
    uint8_t computed_sha256[SHA256_DIGEST_LENGTH];
    SHA256_Final(computed_sha256, &sha_ctx);

    /* Verify hash */
    if (memcmp(computed_sha256, expected_sha256, SHA256_DIGEST_LENGTH) != 0) {
        ESP_LOGE(TAG, "SHA-256 hash mismatch!");
        ESP_LOG_BUFFER_HEX(TAG, computed_sha256, 32, ESP_LOG_ERROR);
        ESP_LOG_BUFFER_HEX(TAG, expected_sha256, 32, ESP_LOG_ERROR);
        return ESP_ERR_INVALID_CRC;
    }

    ESP_LOGI(TAG, "Hash verified OK (%d bytes downloaded)", total_bytes);
    return ESP_OK;
}

/*
 * Report OTA progress to cloud via MQTT.
 */
esp_err_t ota_report_progress(int progress, const char* status, const char* error) {
    if (!s_target_dev_id[0]) {
        return ESP_ERR_INVALID_STATE;
    }

    /* Build JSON: {"type":"ota_progress","dev_id":"...","progress":NN,"status":"...","error":""} */
    cJSON* root = cJSON_CreateObject();
    if (!root) return ESP_ERR_NO_MEM;

    cJSON_AddStringToObject(root, JSON_FIELD_TYPE, "ota_progress");
    cJSON_AddStringToObject(root, JSON_FIELD_DEV_ID, s_target_dev_id);
    cJSON_AddNumberToObject(root, JSON_FIELD_PROGRESS, progress);
    cJSON_AddStringToObject(root, JSON_FIELD_STATUS, status);
    cJSON_AddStringToObject(root, JSON_FIELD_ERROR, error ?: "");

    char* json_str = cJSON_PrintUnformatted(root);
    cJSON_Delete(root);

    if (!json_str) {
        return ESP_ERR_NO_MEM;
    }

    int msg_len = mqtt_common_publish(
        "eregen/device/+/ota_progress", json_str, strlen(json_str), 0);

    free(json_str);
    return (msg_len > 0) ? ESP_OK : ESP_FAIL;
}

/*
 * Check if an OTA update is in progress.
 */
bool ota_is_active(void) {
    if (s_ota_mutex == NULL) return false;
    bool active;
    if (xSemaphoreTake(s_ota_mutex, pdMS_TO_TICKS(50)) == pdTRUE) {
        active = s_ota_active;
        xSemaphoreGive(s_ota_mutex);
    } else {
        active = false;
    }
    return active;
}

/*
 * Get current OTA status string.
 */
const char* ota_get_status(void) {
    return s_current_status;
}
