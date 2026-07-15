/*
 * Eregen (颐贞) - OLED Status Display Module Implementation
 * Smart pillbox tier — SSD1306 0.96" I2C display
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "oled_status.h"

#include <string.h>

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include "driver/i2c.h"
#include "esp_log.h"

/* SSD1306 commands */
#define SSD_SET_CONTRAST      0x81
#define SSD_DISPLAY_OFF       0xAE
#define SSD_DISPLAY_ON        0xAF
#define SSD_SET_MUX_RATIO     0xA8
#define SSD_SET_DISPLAY_OFFSET 0xD3
#define SSD_SET_START_LINE    0x40
#define SSD_SET_SEG_REMAP       0xA1
#define SSD_COM_SCAN_DEC      0xC0
#define SSD_SET_COMP_PINS     0xDA
#define SSD_SET_DISPLAY_CLK   0xD5
#define SSD_SET_PRECHARGE     0xD9
#define SSD_SET_VCOM_DESEL    0xDB
#define SSD_CHARGE_PUMP       0x8D
#define SSD_MEMORY_MODE       0x20
#define SSD_COLUMN_ADDR       0x21
#define SSD_PAGE_ADDR         0x22

/* I2C addresses and timing */
#define I2C_TIMEOUT_MS        1000

/* Log tag */
static const char *TAG = "oled";

/* Internal buffer: 128x64 / 8 = 1024 bytes (one page per byte) */
static uint8_t s_buffer[OLED_WIDTH * OLED_HEIGHT / 8];

/**
 * Write a single byte to the OLED via I2C.
 */
static esp_err_t oled_write_byte(uint8_t data)
{
    i2c_cmd_handle_t cmd = i2c_cmd_link_create();
    i2c_master_start(cmd);
    i2c_master_write_byte(cmd, (OLED_ADDR << 1) | I2C_MASTER_WRITE, true);
    i2c_master_write_byte(cmd, 0x00, true);  /* Control byte: command mode */
    i2c_master_write_byte(cmd, data, true);
    i2c_master_stop(cmd);
    esp_err_t ret = i2c_master_cmd_begin(OLED_I2C_PORT, cmd, pdMS_TO_TICKS(I2C_TIMEOUT_MS));
    i2c_cmd_link_delete(cmd);
    return ret;
}

/**
 * Write multiple bytes to the OLED via I2C.
 */
static esp_err_t oled_write_bytes(const uint8_t *data, size_t len)
{
    i2c_cmd_handle_t cmd = i2c_cmd_link_create();
    i2c_master_start(cmd);
    i2c_master_write_byte(cmd, (OLED_ADDR << 1) | I2C_MASTER_WRITE, true);
    i2c_master_write_byte(cmd, 0x40, true);  /* Control byte: data mode */
    i2c_master_write(cmd, data, len, true);
    i2c_master_stop(cmd);
    esp_err_t ret = i2c_master_cmd_begin(OLED_I2C_PORT, cmd, pdMS_TO_TICKS(I2C_TIMEOUT_MS));
    i2c_cmd_link_delete(cmd);
    return ret;
}

esp_err_t oled_init(void)
{
    i2c_config_t conf = {
        .mode = I2C_MODE_MASTER,
        .sda_io_num = OLED_SDA_GPIO,
        .scl_io_num = OLED_SCL_GPIO,
        .sda_pullup_en = GPIO_PULLUP_ENABLE,
        .scl_pullup_en = GPIO_PULLUP_ENABLE,
        .master.clock_speed = 100000,
    };

    esp_err_t ret = i2c_param_config(OLED_I2C_PORT, &conf);
    if (ret != ESP_OK)
        return ret;

    ret = i2c_driver_install(OLED_I2C_PORT, conf.mode, 0, 0, 0);
    if (ret != ESP_OK)
        return ret;

    /* Reset buffer */
    memset(s_buffer, 0, sizeof(s_buffer));

    /* SSD1306 initialization sequence */
    const uint8_t init_seq[] = {
        SSD_DISPLAY_OFF,
        SSD_SET_DISPLAY_CLK,  0x80,
        SSD_SET_MUX_RATIO,    0x3F,  /* 64 */
        SSD_SET_DISPLAY_OFFSET, 0x00,
        SSD_SET_START_LINE,
        SSD_SET_SEG_REMAP,
        SSD_COM_SCAN_DEC,
        SSD_SET_COMP_PINS,    0x12,
        SSD_SET_CONTRAST,     0xCF,
        SSD_SET_PRECHARGE,    0xF1,
        SSD_SET_VCOM_DESEL,   0x30,
        SSD_CHARGE_PUMP,      0x14,
        SSD_MEMORY_MODE,      0x00,  /* Horizontal addressing */
        SSD_COLUMN_ADDR,      0, 0, 127,
        SSD_PAGE_ADDR,        0, 7,
        SSD_DISPLAY_ON,
    };

    for (size_t i = 0; i < sizeof(init_seq); i++) {
        oled_write_byte(init_seq[i]);
    }

    ESP_LOGI(TAG, "OLED initialized (SSD1306, 128x64)");
    return ESP_OK;
}

void oled_clear(void)
{
    memset(s_buffer, 0, sizeof(s_buffer));
}

void oled_draw_status_bar(uint8_t battery_percent, bool wifi_connected)
{
    /* Row 0: status bar (page 0) */
    /* Battery indicator: 4 segments on the right side */
    uint8_t segments = battery_percent / 25;
    for (uint8_t i = 0; i < 4; i++) {
        if (i < segments) {
            s_buffer[0] |= (1 << i);
            s_buffer[1] |= (1 << i);
        }
    }

    /* WiFi icon at far right */
    if (wifi_connected) {
        s_buffer[0] |= 0x80;
        s_buffer[1] |= 0xC0;
        s_buffer[2] |= 0xE0;
    }

    /* Draw separator line at row 1 */
    for (int x = 0; x < OLED_WIDTH; x++) {
        s_buffer[x] |= 0x02;
    }
}

void oled_draw_medication_list(const char (*compartments)[16], uint8_t count)
{
    /* Display compartments starting at page 2, ~2 rows per compartment */
    for (uint8_t i = 0; i < count && i < 6; i++) {
        int page = 2 + i * 2;
        if (page + 1 >= 8)
            break;

        /* Draw text using a simple monospace font (5x7 chars) */
        const char *text = compartments[i];
        int col = 2 + (i % 2) * 64;
        int pg = page + (i / 2);

        for (int c = 0; text[c] != '\0' && c < 12; c++) {
            uint8_t ch = text[c];
            if (ch >= ' ' && ch <= '~') {
                ch -= ' ';  /* ASCII offset */
                /* Simple 5-wide font lookup would go here.
                 * For now, draw placeholder blocks. */
                for (int row = 0; row < 7 && (pg + row) < 8; row++) {
                    s_buffer[col + c] |= (1 << row);
                }
            }
        }
    }
}

void oled_draw_next_reminder(uint8_t next_time_hours, uint8_t next_time_minutes)
{
    /* Draw time string "HH:MM" at page 6-7 */
    char time_str[6];
    snprintf(time_str, sizeof(time_str), "%02d:%02d", next_time_hours, next_time_minutes);

    /* Place at bottom of screen */
    int col = (OLED_WIDTH / 2) - 2;
    int pg = 6;

    for (int c = 0; c < 5 && pg < 8; c++) {
        uint8_t ch = time_str[c];
        if (ch >= ' ' && ch <= '~') {
            ch -= ' ';
            for (int row = 0; row < 7; row++) {
                if ((col + c) < OLED_WIDTH) {
                    s_buffer[(col + c) + (pg * OLED_WIDTH)] |= (1 << row);
                }
            }
        }
    }
}

void oled_refresh(void)
{
    /* Send buffer pages to display */
    for (uint8_t page = 0; page < 8; page++) {
        /* Set page address */
        oled_write_byte(SSD_PAGE_ADDR);
        oled_write_byte(page);
        oled_write_byte(7);

        /* Send 128 bytes for this page */
        uint8_t page_data[OLED_WIDTH];
        for (int x = 0; x < OLED_WIDTH; x++) {
            page_data[x] = s_buffer[x * 8 + page];
        }
        oled_write_bytes(page_data, OLED_WIDTH);
    }
}
