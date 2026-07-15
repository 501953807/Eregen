/*
 * Eregen (颐贞) - OLED Status Display Module
 * Smart pillbox tier — SSD1306 0.96" I2C display for medication status
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef OLED_STATUS_H
#define OLED_STATUS_H

#include "esp_err.h"
#include <stdint.h>
#include <stdbool.h>

/* I2C bus configuration */
#define OLED_I2C_PORT           I2C_NUM_0
#define OLED_SDA_GPIO           GPIO_NUM_5
#define OLED_SCL_GPIO           GPIO_NUM_6
#define OLED_ADDR               0x3C

/* Display dimensions */
#define OLED_WIDTH              128
#define OLED_HEIGHT             64

/**
 * Initialize OLED display via I2C.
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t oled_init(void);

/**
 * Clear the entire display buffer.
 */
void oled_clear(void);

/**
 * Draw the status bar at the top of the screen.
 * Shows device ID prefix, battery level, and WiFi indicator.
 *
 * @param battery_percent Current battery percentage (0-100)
 * @param wifi_connected True if WiFi is connected
 */
void oled_draw_status_bar(uint8_t battery_percent, bool wifi_connected);

/**
 * Draw the medication list showing compartment statuses.
 * Up to 8 compartments displayed as rows.
 *
 * @param compartments Array of compartment status strings
 * @param count Number of compartments
 */
void oled_draw_medication_list(const char (*compartments)[16], uint8_t count);

/**
 * Draw the next upcoming reminder time.
 *
 * @param next_time_hours Next reminder hour (0-23)
 * @param next_time_minutes Next reminder minutes (0-59)
 */
void oled_draw_next_reminder(uint8_t next_time_hours, uint8_t next_time_minutes);

/**
 * Refresh the physical display from the internal buffer.
 */
void oled_refresh(void);

#endif /* OLED_STATUS_H */
