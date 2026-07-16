/*
 * Eregen (颐贞) - AMOLED Display Driver Header
 * ST7701S-compatible 240x296 AMOLED panel driver via SPI.
 * Supports rich UI primitives: circles, arcs, text, gradients.
 *
 * This replaces the entry-level ST7789 driver with a higher-resolution
 * AMOLED panel suitable for the Pro tier metal housing form factor.
 *
 * Key differences from entry:
 *   - 240x296 resolution (vs 135x240)
 *   - RGB565 over 4-wire SPI (command + data)
 *   - Hardware scroll and partial update support
 *   - Gradient fill for ring UI elements
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef DISPLAY_AMOLED_H
#define DISPLAY_AMOLED_H

#include <stdint.h>
#include <stdbool.h>

/* ----------------------------------------------------------------
 * Display specifications
 * ---------------------------------------------------------------- */
#define AMOLED_WIDTH             240U
#define AMOLED_HEIGHT            296U
#define AMOLED_PIXEL_FORMAT      16U  /* RGB565 */

/* SPI command codes */
#define AMOLED_CMD_NOP           0x00U
#define AMOLED_CMD_SWRESET       0x01U
#define AMOLED_CMD_RDDID         0x04U
#define AMOLED_CMD_RDDST         0x0AU
#define AMOLED_CMD_RPMCTR        0xB0U
#define AMOLED_CMD_RPMSEL        0xB3U
#define AMOLED_CMD_SLPOUT        0x11U
#define AMOLED_CMD_SLPIN         0x10U
#define AMOLED_CMD_PTLON         0x12U
#define AMOLED_CMD_NORON         0x13U
#define AMOLED_CMD_INVOFF        0x20U
#define AMOLED_CMD_INVON         0x21U
#define AMOLED_CMD_DISPOFF       0x28U
#define AMOLED_CMD_DISPON        0x29U
#define AMOLED_CMD_CASET         0x2AU
#define AMOLED_CMD_RASET         0x2BU
#define AMOLED_CMD_RAMWR         0x2CU
#define AMOLED_CMD_RAMRD         0x2EU
#define AMOLED_CMD_PTLAR         0x30U
#define AMOLED_CMD_VSCRDEF       0x33U
#define AMOLED_CMD_MADCTL        0x36U
#define AMOLED_CMD_COLMOD        0x3AU
#define AMOLED_CMD_DFCMD         0xB6U
#define AMOLED_CMD_WRMEMCONT     0x3CU

/* Color definitions (RGB565) */
#define AMOLED_BLACK       0x0000U
#define AMOLED_WHITE       0xFFFFU
#define AMOLED_RED         0xF800U
#define AMOLED_GREEN       0x07E0U
#define AMOLED_BLUE        0x001FU
#define AMOLED_YELLOW      0xFFE0U
#define AMOLED_CYAN        0x07FFU
#define AMOLED_MAGENTA     0xF81FU
#define AMOLED_ORANGE      0xFA60U
#define AMOLED_DARK_GRAY   0x7BEFU
#define AMOLED_LIGHT_GRAY  0xE7CDU
#define AMOLED_TEAL        0x041FU
#define AMOLED_PURPLE      0x780FU

/* Display orientation (MADCTL bits) */
#define AMOLED_MADCTL_MY     0x80U
#define AMOLED_MADCTL_MX     0x40U
#define AMOLED_MADCTL_MV     0x20U
#define AMOLED_MADCTL_ML     0x10U
#define AMOLED_MADCTL_RGB    0x00U
#define AMOLED_MADCTL_BGR     0x08U

/* Built-in font (5x7 pixel bitmap) */
extern const uint8_t amoled_font_5x7[];

/* ----------------------------------------------------------------
 * Drawing primitives
 * ---------------------------------------------------------------- */

/**
 * Initialize the AMOLED display over SPI.
 * Sends the ST7701S initialization sequence, powers on panel.
 * @return true on success.
 */
bool amoled_display_init(void);

/**
 * Clear the entire display to a solid color.
 * @param color RGB565 color value.
 */
void amoled_display_clear(uint16_t color);

/**
 * Push the internal frame buffer to the display.
 * Must be called after drawing operations to update screen.
 */
void amoled_display_update(void);

/**
 * Draw a filled rectangle.
 * @param x0 Top-left X coordinate.
 * @param y0 Top-left Y coordinate.
 * @param x1 Bottom-right X coordinate.
 * @param y1 Bottom-right Y coordinate.
 * @param color RGB565 fill color.
 */
void amoled_draw_rect_filled(uint16_t x0, uint16_t y0,
                             uint16_t x1, uint16_t y1, uint16_t color);

/**
 * Draw a hollow rectangle outline.
 * @param x0 Top-left X.
 * @param y0 Top-left Y.
 * @param w Width in pixels.
 * @param h Height in pixels.
 * @param color RGB565 outline color.
 */
void amoled_draw_rect(uint16_t x0, uint16_t y0,
                      uint16_t w, uint16_t h, uint16_t color);

/**
 * Draw a filled circle.
 * @param cx Center X.
 * @param cy Center Y.
 * @param r Radius in pixels.
 * @param color RGB565 fill color.
 */
void amoled_draw_circle(uint16_t cx, uint16_t cy, uint16_t r, uint16_t color);

/**
 * Draw a circle outline (hollow).
 * @param cx Center X.
 * @param cy Center Y.
 * @param r Radius in pixels.
 * @param color RGB565 outline color.
 */
void amoled_draw_circle_outline(uint16_t cx, uint16_t cy,
                                uint16_t r, uint16_t color);

/**
 * Draw an arc (partial circle segment).
 * @param cx Center X.
 * @param cy Center Y.
 * @param r Outer radius in pixels.
 * @param start_angle Start angle in degrees (0 = top, clockwise).
 * @param end_angle End angle in degrees (exclusive).
 * @param thickness Arc line thickness in pixels.
 * @param color RGB565 arc color.
 */
void amoled_draw_arc(uint16_t cx, uint16_t cy, uint16_t r,
                     uint16_t start_angle, uint16_t end_angle,
                     uint8_t thickness, uint16_t color);

/**
 * Draw a horizontal gradient-filled rectangle.
 * Interpolates between two colors across the width.
 * @param x0 Left X.
 * @param y0 Top Y.
 * @param x1 Right X.
 * @param y1 Bottom Y.
 * @param color_left RGB565 left edge color.
 * @param color_right RGB565 right edge color.
 */
void amoled_draw_gradient_h(uint16_t x0, uint16_t y0,
                            uint16_t x1, uint16_t y1,
                            uint16_t color_left, uint16_t color_right);

/**
 * Draw a vertical gradient-filled rectangle.
 * @param x0 Left X.
 * @param y0 Top Y.
 * @param x1 Right X.
 * @param y1 Bottom Y.
 * @param color_top RGB565 top edge color.
 * @param color_bottom RGB565 bottom edge color.
 */
void amoled_draw_gradient_v(uint16_t x0, uint16_t y0,
                            uint16_t x1, uint16_t y1,
                            uint16_t color_top, uint16_t color_bottom);

/**
 * Draw a single character using the built-in 5x7 font.
 * @param x Column coordinate.
 * @param y Row coordinate.
 * @param ch ASCII character.
 * @param color RGB565 foreground color.
 * @param bg_color RGB565 background color (use AMOLED_BLACK to skip).
 */
void amoled_draw_char(uint16_t x, uint16_t y, char ch,
                      uint16_t color, uint16_t bg_color);

/**
 * Draw a null-terminated string.
 * @param x Starting column.
 * @param y Starting row.
 * @param str Null-terminated string.
 * @param color RGB565 text color.
 * @param bg_color RGB565 background color.
 */
void amoled_draw_string(uint16_t x, uint16_t y, const char *str,
                        uint16_t color, uint16_t bg_color);

/**
 * Draw a horizontal line.
 * @param x0 Start X.
 * @param y Y coordinate.
 * @param x1 End X.
 * @param color RGB565 line color.
 */
void amoled_draw_line_h(uint16_t x0, uint16_t y, uint16_t x1, uint16_t color);

/**
 * Draw a vertical line.
 * @param x X coordinate.
 * @param y0 Start Y.
 * @param y1 End Y.
 * @param color RGB565 line color.
 */
void amoled_draw_line_v(uint16_t x, uint16_t y0, uint16_t y1, uint16_t color);

/**
 * Draw a diagonal line (Bresenham's algorithm).
 * @param x0 Start X.
 * @param y0 Start Y.
 * @param x1 End X.
 * @param y1 End Y.
 * @param color RGB565 line color.
 */
void amoled_draw_line(uint16_t x0, uint16_t y0,
                      uint16_t x1, uint16_t y1, uint16_t color);

/**
 * Set the scroll region for partial updates.
 * @param top Fixed rows at top (not scrolled).
 * @param bot Fixed rows at bottom (not scrolled).
 */
void amoled_set_scroll_area(uint16_t top, uint16_t bot);

/**
 * Scroll the display content vertically by N rows.
 * @param rows Number of rows to scroll (positive = down).
 */
void amoled_scroll_vertical(uint8_t rows);

/**
 * Get display width.
 * @return Width in pixels.
 */
uint16_t amoled_get_width(void);

/**
 * Get display height.
 * @return Height in pixels.
 */
uint16_t amoled_get_height(void);

#endif /* DISPLAY_AMOLED_H */
