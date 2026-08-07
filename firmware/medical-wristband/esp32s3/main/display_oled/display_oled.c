/*
 * Eregen Medical Wristband - OLED Display Module
 * SSD1306 I2C display for patient info and alerts
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "display_oled.h"
#include <string.h>
#include <stdio.h>
#include "esp_log.h"
#include "driver/i2c.h"
#include "oled_ssd1306.h"

static const char *TAG = "display_oled";
static oled_handle_t s_oled = NULL;

esp_err_t display_oled_init(void) {
    oled_config_t config = {
        .i2c_port = I2C_NUM_0,
        .i2c_addr = OLED_I2C_ADDRESS,
        .width = 128,
        .height = 64,
    };

    esp_err_t err = oled_init(&config, &s_oled);
    if (err == ESP_OK) {
        ESP_LOGI(TAG, "OLED display initialized");
    }
    return err;
}

esp_err_t display_oled_show_patient(const patient_info_t *info) {
    if (!s_oled || !info) return ESP_ERR_INVALID_ARG;

    oled_clear(s_oled);

    // Header
    oled_set_font(s_oled, FONT_6X8);
    oled_print(s_oled, "EREGEN MEDICAL", 0, 0);
    oled_print(s_oled, "Wristband", 0, 10);

    // Patient info
    char line[32];
    snprintf(line, sizeof(line), "Patient: %.16s", info->name);
    oled_print(s_oled, line, 0, 22);

    snprintf(line, sizeof(line), "Room: %s", info->bed_number);
    oled_print(s_oled, line, 0, 32);

    snprintf(line, sizeof(line), "Dept: %s", info->department);
    oled_print(s_oled, line, 0, 42);

    // Alert tags
    if (strlen(info->alert_tags) > 0) {
        oled_print(s_oled, "ALERTS:", 0, 52);
        oled_print(s_oled, info->alert_tags, 0, 60);
    }

    return oled_refresh(s_oled);
}

esp_err_t display_oled_show_alert(const char *alert_text, oled_alert_color_t color) {
    if (!s_oled || !alert_text) return ESP_ERR_INVALID_ARG;

    // Flash alert message
    for (int i = 0; i < 3; i++) {
        oled_clear(s_oled);
        oled_set_font(s_oled, FONT_8X16);
        oled_set_color(s_oled, color == OLED_ALERT_RED ? WHITE : BLACK);
        oled_print(s_oled, alert_text, 8, 24);
        oled_refresh(s_oled);
        vTaskDelay(pdMS_TO_TICKS(200));

        oled_clear(s_oled);
        oled_refresh(s_oled);
        vTaskDelay(pdMS_TO_TICKS(200));
    }

    return ESP_OK;
}

esp_err_t display_oled_show_battery(int percent) {
    if (!s_oled) return ESP_ERR_INVALID_ARG;

    char buf[16];
    snprintf(buf, sizeof(buf), "BAT: %d%%", percent);
    oled_print(s_oled, buf, 96, 56);
    return oled_refresh(s_oled);
}

void display_oled_deinit(void) {
    if (s_oled) {
        oled_deinit(s_oled);
        s_oled = NULL;
    }
}
