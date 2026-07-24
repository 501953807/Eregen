/*
 * Eregen (颐贞) - AMOLED Display Driver Implementation
 * ST7701S 240x296 panel driver over SPI.
 *
 * Memory-constrained design: GD32E230C8T3 has only 20KB SRAM.
 * Full frame buffer (142KB) is impossible. This driver uses
 * immediate-mode rendering: each primitive pushes pixels
 * directly to the display via SPI. No persistent frame buffer.
 *
 * Total SRAM usage: ~480 bytes (single-row scratch buffer).
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "display_amoled.h"
#include "board_pro.h"
#include "../common/log.h"
#include <string.h>
#include <math.h>

/* Single-row scratch buffer: 240 * 2 = 480 bytes */
static uint16_t s_scratch[AMOLED_WIDTH];

/* ----------------------------------------------------------------
 * SPI helpers
 * ---------------------------------------------------------------- */

static void amoled_spi_send_cmd(uint8_t cmd)
{
    gpio_bit_reset(BOARD_PRO_DISPLAY_DC_PORT, BOARD_PRO_DISPLAY_DC_PIN);
    gpio_bit_reset(BOARD_PRO_DISPLAY_CS_PORT, BOARD_PRO_DISPLAY_CS_PIN);

    while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {}
    spi_data_transmit(SPI1, cmd);
    while (spi_flag_get(SPI1, SPI_FLAG_RBNE) == RESET) {}
    (void)spi_data_receive(SPI1);

    gpio_bit_set(BOARD_PRO_DISPLAY_CS_PORT, BOARD_PRO_DISPLAY_CS_PIN);
}

static void amoled_spi_send_data(const uint8_t *data, uint16_t len)
{
    if (!data || len == 0) return;

    gpio_bit_set(BOARD_PRO_DISPLAY_DC_PORT, BOARD_PRO_DISPLAY_DC_PIN);
    gpio_bit_reset(BOARD_PRO_DISPLAY_CS_PORT, BOARD_PRO_DISPLAY_CS_PIN);

    for (uint16_t i = 0; i < len; i++) {
        while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {}
        spi_data_transmit(SPI1, data[i]);
        while (spi_flag_get(SPI1, SPI_FLAG_RBNE) == RESET) {}
        (void)spi_data_receive(SPI1);
    }

    gpio_bit_set(BOARD_PRO_DISPLAY_CS_PORT, BOARD_PRO_DISPLAY_CS_PIN);
}

static void amoled_spi_send_data_simple(const uint8_t *data, uint16_t len)
{
    if (!data || len == 0) return;
    gpio_bit_set(BOARD_PRO_DISPLAY_DC_PORT, BOARD_PRO_DISPLAY_DC_PIN);
    gpio_bit_reset(BOARD_PRO_DISPLAY_CS_PORT, BOARD_PRO_DISPLAY_CS_PIN);
    for (uint16_t i = 0; i < len; i++) {
        while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {}
        spi_data_transmit(SPI1, data[i]);
        while (spi_flag_get(SPI1, SPI_FLAG_RBNE) == RESET) {}
        (void)spi_data_receive(SPI1);
    }
    gpio_bit_set(BOARD_PRO_DISPLAY_CS_PORT, BOARD_PRO_DISPLAY_CS_PIN);
}

/* Set address window and prepare for pixel data writes */
static void amoled_set_window(uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1)
{
    amoled_spi_send_cmd(AMOLED_CMD_CASET);
    uint8_t caset[4] = { (uint8_t)(x0 >> 8), (uint8_t)x0,
                         (uint8_t)(x1 >> 8), (uint8_t)x1 };
    amoled_spi_send_data(caset, 4);

    amoled_spi_send_cmd(AMOLED_CMD_RASET);
    uint8_t raset[4] = { (uint8_t)(y0 >> 8), (uint8_t)y0,
                         (uint8_t)(y1 >> 8), (uint8_t)y1 };
    amoled_spi_send_data(raset, 4);

    amoled_spi_send_cmd(AMOLED_CMD_RAMWR);
}

/* Push one row from scratch buffer to display */
static void amoled_push_row(uint16_t y)
{
    amoled_set_window(0, y, AMOLED_WIDTH - 1, y);
    amoled_spi_send_data((const uint8_t *)s_scratch, AMOLED_WIDTH * 2);
}

/* Fill horizontal segment [x0..x1] at row y with color */
static void amoled_fill_row_segment(uint16_t y, uint16_t x0, uint16_t x1,
                                    uint16_t color)
{
    if (y >= AMOLED_HEIGHT) return;
    if (x1 < x0) { uint16_t t = x0; x0 = x1; x1 = t; }
    if (x0 >= AMOLED_WIDTH) return;
    if (x1 >= AMOLED_WIDTH) x1 = AMOLED_WIDTH - 1;
    if (x1 < x0) return;

    amoled_set_window(x0, y, x1, y);
    uint16_t count = x1 - x0 + 1;
    for (uint16_t i = 0; i < count; i++) {
        uint8_t hi = (uint8_t)(color >> 8);
        uint8_t lo = (uint8_t)(color & 0xFF);
        uint8_t pair[2] = { hi, lo };
        amoled_spi_send_data(pair, 2);
    }
}

/* ----------------------------------------------------------------
 * Initialization sequence for ST7701S
 * ---------------------------------------------------------------- */

bool amoled_display_init(void)
{
    /* Reset display */
    gpio_bit_reset(BOARD_PRO_DISPLAY_RST_PORT, BOARD_PRO_DISPLAY_RST_PIN);
    for (volatile uint32_t d = 0; d < 100000U; d++) (void)d;
    gpio_bit_set(BOARD_PRO_DISPLAY_RST_PORT, BOARD_PRO_DISPLAY_RST_PIN);
    for (volatile uint32_t d = 0; d < 1200000U; d++) (void)d;

    /* PMR: Power Mode Register */
    uint8_t pmr_data[] = { 0x55, 0x90, 0x2B, 0x00, 0x0E };
    amoled_spi_send_cmd(AMOLED_CMD_RPMCTR);
    amoled_spi_send_data(pmr_data, sizeof(pmr_data));

    /* DFCN: Refresh Control */
    uint8_t dfcn_data[] = { 0x00, 0x10, 0x30, 0x10 };
    amoled_spi_send_cmd(AMOLED_CMD_DFCMD);
    amoled_spi_send_data(dfcn_data, sizeof(dfcn_data));

    /* Sleep Out */
    amoled_spi_send_cmd(AMOLED_CMD_SLPOUT);
    for (volatile uint32_t d = 0; d < 1200000U; d++) (void)d;

    /* Pixel Format: 16-bit RGB565 */
    amoled_spi_send_cmd(AMOLED_CMD_COLMOD);
    uint8_t colmod = 0x55;
    amoled_spi_send_data_simple(&colmod, 1);

    /* Memory Access Control */
    amoled_spi_send_cmd(AMOLED_CMD_MADCTL);
    uint8_t madctl = AMOLED_MADCTL_RGB;
    amoled_spi_send_data_simple(&madctl, 1);

    /* Display On */
    amoled_spi_send_cmd(AMOLED_CMD_DISPON);
    for (volatile uint32_t d = 0; d < 120000U; d++) (void)d;

    /* Clear to black */
    amoled_display_clear(AMOLED_BLACK);

    log_info("AMOLED: ST7701S initialized (%ux%u, RGB565)",
             AMOLED_WIDTH, AMOLED_HEIGHT);
    return true;
}

/* ----------------------------------------------------------------
 * Core: clear display
 * ---------------------------------------------------------------- */

void amoled_display_clear(uint16_t color)
{
    for (uint16_t i = 0; i < AMOLED_WIDTH; i++) {
        s_scratch[i] = color;
    }

    amoled_spi_send_cmd(AMOLED_CMD_CASET);
    uint8_t caset[4] = { 0, 0, (uint8_t)(AMOLED_WIDTH - 1), 0 };
    amoled_spi_send_data(caset, 4);

    for (uint16_t y = 0; y < AMOLED_HEIGHT; y++) {
        amoled_spi_send_cmd(AMOLED_CMD_RASET);
        uint8_t raset[4] = { 0, (uint8_t)y, 0, (uint8_t)(y + 1) };
        amoled_spi_send_data(raset, 4);

        amoled_spi_send_cmd(AMOLED_CMD_RAMWR);
        amoled_spi_send_data((const uint8_t *)s_scratch, AMOLED_WIDTH * 2);
    }
}

void amoled_display_update(void)
{
    /* Immediate-mode: no frame buffer, nothing to flush.
     * All drawing operations push directly to display. */
}

/* ----------------------------------------------------------------
 * Drawing primitives — all immediate mode
 * ---------------------------------------------------------------- */

void amoled_draw_rect_filled(uint16_t x0, uint16_t y0,
                             uint16_t x1, uint16_t y1, uint16_t color)
{
    if (x0 >= AMOLED_WIDTH || y0 >= AMOLED_HEIGHT) return;
    if (x1 >= AMOLED_WIDTH) x1 = AMOLED_WIDTH - 1;
    if (y1 >= AMOLED_HEIGHT) y1 = AMOLED_HEIGHT - 1;
    if (x1 < x0 || y1 < y0) return;

    for (uint16_t y = y0; y <= y1; y++) {
        amoled_fill_row_segment(y, x0, x1, color);
    }
}

void amoled_draw_rect(uint16_t x0, uint16_t y0,
                      uint16_t w, uint16_t h, uint16_t color)
{
    if (w == 0 || h == 0) return;
    uint16_t right = (x0 + w > AMOLED_WIDTH) ? AMOLED_WIDTH - 1 : x0 + w - 1;
    uint16_t bottom = (y0 + h > AMOLED_HEIGHT) ? AMOLED_HEIGHT - 1 : y0 + h - 1;
    if (right < x0 || bottom < y0) return;

    amoled_draw_rect_filled(x0, y0, right, y0, color);
    amoled_draw_rect_filled(x0, bottom, right, bottom, color);
    amoled_draw_rect_filled(x0, y0, x0, bottom, color);
    amoled_draw_rect_filled(right, y0, right, bottom, color);
}

void amoled_draw_circle(uint16_t cx, uint16_t cy, uint16_t r, uint16_t color)
{
    if (r == 0) return;
    int16_t x = (int16_t)r;
    int16_t y = 0;
    int16_t decision = 1 - x;

    while (y <= x) {
        if (decision > 0) {
            x--;
            decision += 2 * (1 - x);
        }
        y++;
        decision += 2 * y + 1;

        for (int16_t dy = -y; dy <= y; dy++) {
            uint16_t row = (uint16_t)((int16_t)cy + dy);
            if (row < AMOLED_HEIGHT) {
                int16_t lx = (int16_t)cx - x;
                int16_t rx = (int16_t)cx + x;
                if (rx >= 0 && lx < (int16_t)AMOLED_WIDTH) {
                    amoled_fill_row_segment(row,
                        (lx < 0) ? 0 : (uint16_t)lx,
                        (rx >= (int16_t)AMOLED_WIDTH) ? AMOLED_WIDTH - 1 : (uint16_t)rx,
                        color);
                }
            }
        }
        for (int16_t dy = -x; dy <= x; dy++) {
            uint16_t row = (uint16_t)((int16_t)cy + dy);
            if (row < AMOLED_HEIGHT) {
                int16_t lx = (int16_t)cx - y;
                int16_t rx = (int16_t)cx + y;
                if (rx >= 0 && lx < (int16_t)AMOLED_WIDTH) {
                    amoled_fill_row_segment(row,
                        (lx < 0) ? 0 : (uint16_t)lx,
                        (rx >= (int16_t)AMOLED_WIDTH) ? AMOLED_WIDTH - 1 : (uint16_t)rx,
                        color);
                }
            }
        }
    }
}

void amoled_draw_circle_outline(uint16_t cx, uint16_t cy,
                                uint16_t r, uint16_t color)
{
    if (r == 0) return;
    int16_t x = (int16_t)r;
    int16_t y = 0;
    int16_t d = 3 - 2 * x;

    while (y <= x) {
        /* 8 symmetric points as single-pixel segments */
        #define _PLOT_8(_px, _py) \
            do { \
                int16_t _x = (int16_t)(_px), _y = (int16_t)(_py); \
                if (_x >= 0 && _x < (int16_t)AMOLED_WIDTH && \
                    _y >= 0 && _y < (int16_t)AMOLED_HEIGHT) \
                    amoled_fill_row_segment((uint16_t)_y, (uint16_t)_x, (uint16_t)_x, color); \
            } while (0)

        _PLOT_8(cx + x, cy + y);
        _PLOT_8(cx - x, cy + y);
        _PLOT_8(cx + x, cy - y);
        _PLOT_8(cx - x, cy - y);
        _PLOT_8(cx + y, cy + x);
        _PLOT_8(cx - y, cy + x);
        _PLOT_8(cx + y, cy - x);
        _PLOT_8(cx - y, cy - x);

        if (d > 0) { x--; d += 4 * (x - y) + 10; }
        else       { d += 4 * y + 6; }
        y++;
    }
    #undef _PLOT_8
}

void amoled_draw_arc(uint16_t cx, uint16_t cy, uint16_t r,
                     uint16_t start_angle, uint16_t end_angle,
                     uint8_t thickness, uint16_t color)
{
    if (start_angle > end_angle) {
        amoled_draw_arc(cx, cy, r, start_angle, 360U, thickness, color);
        amoled_draw_arc(cx, cy, r, 0, end_angle, thickness, color);
        return;
    }

    for (uint16_t angle = start_angle; angle < end_angle; angle++) {
        float rad = angle * 3.14159265f / 180.0f;
        int16_t x_outer = (int16_t)(cx + r * cosf(rad));
        int16_t y_outer = (int16_t)(cy - r * sinf(rad));
        int16_t r_inner = (r > thickness) ? (int16_t)(r - thickness) : 0;
        int16_t x_inner = (int16_t)(cx + r_inner * cosf(rad));
        int16_t y_inner = (int16_t)(cy - r_inner * sinf(rad));

        if (x_inner < 0) x_inner = 0;
        if (x_outer >= (int16_t)AMOLED_WIDTH) x_outer = AMOLED_WIDTH - 1;
        if (y_inner < 0) y_inner = 0;
        if (y_outer >= (int16_t)AMOLED_HEIGHT) y_outer = AMOLED_HEIGHT - 1;

        if (x_inner <= x_outer && y_inner <= y_outer) {
            amoled_fill_row_segment((uint16_t)y_inner, (uint16_t)x_inner,
                                    (uint16_t)x_outer, color);
        }
    }
}

/* ----------------------------------------------------------------
 * Gradient helpers
 * ---------------------------------------------------------------- */

static uint16_t rgb565_lerp(uint16_t a, uint16_t b, float t)
{
    uint16_t ar = (a >> 11) & 0x1F;
    uint16_t ag = (a >> 5) & 0x3F;
    uint16_t ab = a & 0x1F;
    uint16_t br = (b >> 11) & 0x1F;
    uint16_t bg = (b >> 5) & 0x3F;
    uint16_t bb = b & 0x1F;

    uint16_t cr = (uint16_t)(ar + t * (br - ar) + 0.5f) & 0x1F;
    uint16_t cg = (uint16_t)(ag + t * (bg - ag) + 0.5f) & 0x3F;
    uint16_t cb = (uint16_t)(ab + t * (bb - ab) + 0.5f) & 0x1F;

    return (cr << 11) | (cg << 5) | cb;
}

void amoled_draw_gradient_h(uint16_t x0, uint16_t y0,
                            uint16_t x1, uint16_t y1,
                            uint16_t color_left, uint16_t color_right)
{
    if (x1 < x0) { uint16_t tmp = x0; x0 = x1; x1 = tmp; }
    if (y1 < y0) { uint16_t tmp = y0; y0 = y1; y1 = tmp; }
    if (x1 >= AMOLED_WIDTH) x1 = AMOLED_WIDTH - 1;
    if (y1 >= AMOLED_HEIGHT) y1 = AMOLED_HEIGHT - 1;

    uint16_t width = x1 - x0 + 1;
    if (width == 0) return;

    for (uint16_t y = y0; y <= y1; y++) {
        for (uint16_t x = x0; x <= x1; x++) {
            float t = (float)(x - x0) / (float)width;
            s_scratch[x] = rgb565_lerp(color_left, color_right, t);
        }
        amoled_push_row(y);
    }
}

void amoled_draw_gradient_v(uint16_t x0, uint16_t y0,
                            uint16_t x1, uint16_t y1,
                            uint16_t color_top, uint16_t color_bottom)
{
    if (x1 < x0) { uint16_t tmp = x0; x0 = x1; x1 = tmp; }
    if (y1 < y0) { uint16_t tmp = y0; y0 = y1; y1 = tmp; }
    if (x1 >= AMOLED_WIDTH) x1 = AMOLED_WIDTH - 1;
    if (y1 >= AMOLED_HEIGHT) y1 = AMOLED_HEIGHT - 1;

    uint16_t height = y1 - y0 + 1;
    uint16_t width = x1 - x0 + 1;
    if (height == 0 || width == 0) return;

    for (uint16_t y = y0; y <= y1; y++) {
        float t = (float)(y - y0) / (float)height;
        uint16_t color = rgb565_lerp(color_top, color_bottom, t);
        for (uint16_t i = 0; i < width; i++) {
            s_scratch[x0 + i] = color;
        }
        amoled_push_row(y);
    }
}

/* ----------------------------------------------------------------
 * Text rendering
 * ---------------------------------------------------------------- */

void amoled_draw_char(uint16_t x, uint16_t y, char ch,
                      uint16_t color, uint16_t bg_color)
{
    if (ch < 0x20 || ch > 0x7E) return;

    int idx = (ch - 0x20) * 5;
    const uint16_t w = 5;
    const uint16_t h = 7;

    if (x + w > AMOLED_WIDTH || y + h > AMOLED_HEIGHT) return;

    for (uint16_t row = 0; row < h; row++) {
        uint8_t pattern = amoled_font_5x7[idx + row];
        for (uint16_t col = 0; col < w; col++) {
            uint16_t c = (pattern & (1U << (w - 1 - col))) ? color : bg_color;
            s_scratch[x + col] = c;
        }
        amoled_push_row(y + row);
    }
}

void amoled_draw_string(uint16_t x, uint16_t y, const char *str,
                        uint16_t color, uint16_t bg_color)
{
    if (!str) return;

    uint16_t cursor_x = x;
    while (*str && cursor_x < AMOLED_WIDTH) {
        if (*str == '\n') {
            cursor_x = x;
            y += 8;
        } else {
            amoled_draw_char(cursor_x, y, *str, color, bg_color);
            cursor_x += 6;
        }
        str++;
    }
}

/* ----------------------------------------------------------------
 * Line drawing
 * ---------------------------------------------------------------- */

void amoled_draw_line_h(uint16_t x0, uint16_t y, uint16_t x1, uint16_t color)
{
    if (y >= AMOLED_HEIGHT) return;
    if (x1 < x0) { uint16_t tmp = x0; x0 = x1; x1 = tmp; }
    if (x1 >= AMOLED_WIDTH) x1 = AMOLED_WIDTH - 1;
    if (x1 < x0) return;
    amoled_fill_row_segment(y, x0, x1, color);
}

void amoled_draw_line_v(uint16_t x, uint16_t y0, uint16_t y1, uint16_t color)
{
    if (x >= AMOLED_WIDTH) return;
    if (y1 < y0) { uint16_t tmp = y0; y0 = y1; y1 = tmp; }
    if (y1 >= AMOLED_HEIGHT) y1 = AMOLED_HEIGHT - 1;
    if (y1 < y0) return;

    for (uint16_t y = y0; y <= y1; y++) {
        amoled_fill_row_segment(y, x, x, color);
    }
}

void amoled_draw_line(uint16_t x0, uint16_t y0,
                      uint16_t x1, uint16_t y1, uint16_t color)
{
    int16_t dx = (int16_t)x1 - (int16_t)x0;
    int16_t dy = (int16_t)y1 - (int16_t)y0;
    bool steep = abs(dy) > abs(dx);

    int16_t sx = dx > 0 ? 1 : -1;
    int16_t sy = dy > 0 ? 1 : -1;

    int16_t xx = steep ? dy : dx;
    int16_t xy = steep ? dx : dy;

    if (xx < 0) { xx = -xx; xy = -xy; }

    int16_t y = 0;
    int16_t error = xy / 2;

    for (int16_t i = 0; i <= xx; i++) {
        int16_t px = steep ? y0 + y * sy : x0 + i * sx;
        int16_t py = steep ? x0 + i * sx : y0 + y * sy;

        if (px >= 0 && px < (int16_t)AMOLED_WIDTH &&
            py >= 0 && py < (int16_t)AMOLED_HEIGHT) {
            amoled_fill_row_segment((uint16_t)py, (uint16_t)px, (uint16_t)px, color);
        }

        error -= xy;
        if (error < 0) { y += sy; error += xx; }
    }
}

/* ----------------------------------------------------------------
 * Scroll and queries
 * ---------------------------------------------------------------- */

void amoled_set_scroll_area(uint16_t top, uint16_t bot)
{
    amoled_spi_send_cmd(AMOLED_CMD_VSCRDEF);
    uint8_t scroll_cfg[6] = {
        (uint8_t)(top >> 8), (uint8_t)top,
        (uint8_t)((AMOLED_HEIGHT - top - bot) >> 8),
        (uint8_t)(AMOLED_HEIGHT - top - bot),
        (uint8_t)(bot >> 8), (uint8_t)bot
    };
    amoled_spi_send_data(scroll_cfg, 6);
}

void amoled_scroll_vertical(uint8_t rows)
{
    amoled_spi_send_cmd(0x37U);
    uint8_t cfg[3] = { 0, rows, 0 };
    amoled_spi_send_data(cfg, 3);
}

uint16_t amoled_get_width(void)
{
    return AMOLED_WIDTH;
}

uint16_t amoled_get_height(void)
{
    return AMOLED_HEIGHT;
}
