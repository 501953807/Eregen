/*
 * Eregen Medical Wristband - OLED Display Header
 * SSD1306 I2C 128x64 display
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef DISPLAY_OLED_H
#define DISPLAY_OLED_H

#include "esp_err.h"
#include <stdint.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "patient_store.h"

#define OLED_I2C_ADDRESS    0x3C
#define OLED_WIDTH          128
#define OLED_HEIGHT         64

typedef enum {
    OLED_ALERT_GREEN = 0,
    OLED_ALERT_RED,
    OLED_ALERT_YELLOW,
} oled_alert_color_t;

typedef enum {
    FONT_6X8 = 0,
    FONT_8X16,
} oled_font_t;

esp_err_t display_oled_init(void);
esp_err_t display_oled_show_patient(const patient_info_t *info);
esp_err_t display_oled_show_alert(const char *alert_text, oled_alert_color_t color);
esp_err_t display_oled_show_battery(int percent);
void display_oled_deinit(void);

#endif /* DISPLAY_OLED_H */
