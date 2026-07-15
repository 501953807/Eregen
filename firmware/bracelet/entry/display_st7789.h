/*
 * Eregen (颐贞) - Display Driver Header
 * 1.14" IPS LCD ST7789 via SPI, 135×240 resolution
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef DISPLAY_ST7789_H
#define DISPLAY_ST7789_H

#include <stdint.h>
#include <stdbool.h>

/* Display resolution */
#define DISPLAY_WIDTH          135U
#define DISPLAY_HEIGHT         240U

/* ST7789 command codes */
#define ST7789_NOP             0x00U
#define ST7789_SWRESET         0x01U
#define ST7789_RDDID           0x04U
#define ST7789_RDDST           0x04U
#define ST7789_RDDPM           0x0AU
#define ST7789_RDD_MADCTL      0x0BU
#define ST7789_RDD_COLMOD      0x0CU
#define ST7789_RAMRD           0x0EU
#define ST7789_PTLON           0x0FU
#define ST7789_NORON           0x13U
#define ST7789_INVOFF          0x20U
#define ST7789_INVON           0x21U
#define ST7789_DISPOFF         0x28U
#define ST7789_DISPON          0x29U
#define ST7789_CASET           0x2AU
#define ST7789_RASET           0x2BU
#define ST7789_RAMWR           0x2CU
#define ST7789_RAMWR_BUF       0x3CU
#define ST7789_PTLAR           0x30U
#define ST7789_VSCRDEF         0x33U
#define ST7789_TEOFF           0x34U
#define ST7789_TEON            0x35U
#define ST7789_MADCTL          0x36U
#define ST7789_COLMOD          0x3AU
#define ST7789_WRMEMCONT       0x3CU
#define ST7789_STEPPER         0x44U
#define ST7789_DTIN            0x56U
#define ST7789_PWCTR1          0xC0U
#define ST7789_PWCTR2          0xC1U
#define ST7789_VMCTR1          0xC5U
#define ST7789_VMCTR2          0xC7U
#define ST7789_GAMSET          0x26U
#define ST7789_PGAMCTRL        0xE0U
#define ST7789_NGAMCTRL        0xE1U

/* Color definitions (RGB565) */
#define DISPLAY_COLOR_BLACK       0x0000U
#define DISPLAY_COLOR_WHITE       0xFFFFU
#define DISPLAY_COLOR_RED         0xF800U
#define DISPLAY_COLOR_GREEN       0x07E0U
#define DISPLAY_COLOR_BLUE        0x001FU
#define DISPLAY_COLOR_YELLOW      0xFFE0U
#define DISPLAY_COLOR_CYAN        0x07FFU
#define DISPLAY_COLOR_MAGENTA     0xF81FU

/* Built-in 5×7 pixel font bitmap data */
extern const uint8_t display_font_5x7[];

/*
 * Initialize the ST7789 display over SPI.
 * Sends initialization sequence and powers on the display.
 * @return true on success.
 */
bool display_init(void);

/*
 * Clear the entire display to a solid color.
 * @param color RGB565 color value
 */
void display_clear(uint16_t color);

/*
 * Draw a single character at the given position.
 * @param x Column (0 to DISPLAY_WIDTH-1)
 * @param y Row (0 to DISPLAY_HEIGHT-1)
 * @param ch ASCII character to draw
 * @param color RGB565 foreground color
 * @param bg_color RGB565 background color (use DISPLAY_COLOR_BLACK to skip)
 */
void display_draw_char(uint16_t x, uint16_t y, char ch,
                       uint16_t color, uint16_t bg_color);

/*
 * Draw a null-terminated string at the given position.
 * @param x Starting column
 * @param y Starting row
 * @param str Null-terminated string
 * @param color RGB565 foreground color
 * @param bg_color RGB565 background color
 */
void display_draw_string(uint16_t x, uint16_t y, const char *str,
                         uint16_t color, uint16_t bg_color);

/*
 * Draw a filled circle.
 * @param cx Center X coordinate
 * @param cy Center Y coordinate
 * @param r Radius in pixels
 * @param color RGB565 fill color
 */
void display_draw_circle(uint16_t cx, uint16_t cy, uint8_t r,
                         uint16_t color);

/*
 * Draw a hollow circle (outline only).
 * @param cx Center X coordinate
 * @param cy Center Y coordinate
 * @param r Radius in pixels
 * @param color RGB565 outline color
 */
void display_draw_circle_outline(uint16_t cx, uint16_t cy, uint8_t r,
                                 uint16_t color);

/*
 * Draw a filled rectangle.
 * @param x0 Top-left X
 * @param y0 Top-left Y
 * @param x1 Bottom-right X
 * @param y1 Bottom-right Y
 * @param color RGB565 fill color
 */
void display_draw_rect_filled(uint16_t x0, uint16_t y0,
                              uint16_t x1, uint16_t y1,
                              uint16_t color);

/*
 * Push display buffer to screen.
 * Call after drawing operations to update the screen.
 */
void display_update(void);

/*
 * Set the scroll area for partial screen updates.
 * @param top Fixed rows at top
 * @param bot Fixed rows at bottom
 */
void display_set_scroll_area(uint16_t top, uint16_t bot);

#endif /* DISPLAY_ST7789_H */
