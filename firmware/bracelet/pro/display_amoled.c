/*
 * Eregen (颐贞) - AMOLED Display Driver Implementation
 * ST7701S 240x296 panel driver over SPI with rich UI primitives.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "display_amoled.h"
#include "board_pro.h"
#include "../common/log.h"
#include <string.h>
#include <math.h>

/* Frame buffer: 240x296 @ 16bpp = 142,080 bytes */
/* For SRAM efficiency on Cortex-M4, use a double-buffered approach
 * with a smaller scratch buffer for partial updates. */
static uint16_t s_fb_line[AMOLED_WIDTH]; /* Single-line scratch buffer */

/* Full frame buffer (in external PSRAM or large internal SRAM) */
#if defined(__EMBEDDED__)
static uint16_t s_frame_buffer[AMOLED_WIDTH * AMOLED_HEIGHT];
#else
/* Host build fallback: allocate dynamically */
static uint16_t *s_frame_buffer = NULL;
#endif

/* ----------------------------------------------------------------
 * SPI helpers
 * ---------------------------------------------------------------- */

static void amoled_spi_send_cmd(uint8_t cmd)
{
    /* DC = LOW for command */
    gpio_bit_reset(BOARD_PRO_DISPLAY_DC_PORT, BOARD_PRO_DISPLAY_DC_PIN);
    /* CS = LOW */
    gpio_bit_reset(BOARD_PRO_DISPLAY_CS_PORT, BOARD_PRO_DISPLAY_CS_PIN);

    /* Write byte via SPI */
    while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
        /* Wait */
    }
    spi_data_transmit(SPI1, cmd);

    while (spi_flag_get(SPI1, SPI_FLAG_RBNE) == RESET) {
        /* Wait */
    }
    (void)spi_data_receive(SPI1);

    /* CS = HIGH */
    gpio_bit_set(BOARD_PRO_DISPLAY_CS_PORT, BOARD_PRO_DISPLAY_CS_PIN);
}

static void amoled_spi_send_data(const uint8_t *data, uint16_t len)
{
    if (!data || len == 0) {
        return;
    }

    /* DC = HIGH for data */
    gpio_bit_set(BOARD_PRO_DISPLAY_DC_PORT, BOARD_PRO_DISPLAY_DC_PIN);
    gpio_bit_reset(BOARD_PRO_DISPLAY_CS_PORT, BOARD_PRO_DISPLAY_CS_PIN);

    for (uint16_t i = 0; i < len; i++) {
        while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
            /* Wait */
        }
        spi_data_transmit(SPI1, data[i]);

        while (spi_flag_get(SPI1, SPI_FLAG_RBNE) == RESET) {
            /* Wait */
        }
        (void)spi_data_receive(SPI1);
    }

    gpio_bit_set(BOARD_PRO_DISPLAY_CS_PORT, BOARD_PRO_DISPLAY_CS_PIN);
}

/* Re-read the CS pin reset properly */
static void amoled_cs_high(void)
{
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
    amoled_cs_high();
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
    for (volatile uint32_t d = 0; d < 1200000U; d++) (void)d; /* 10ms */

    /* Send ST7701S initialization commands */
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
    for (volatile uint32_t d = 0; d < 1200000U; d++) (void)d; /* 120ms */

    /* Pixel Format: 16-bit RGB565 */
    amoled_spi_send_cmd(AMOLED_CMD_COLMOD);
    uint8_t colmod = 0x55; /* 16 bits/pixel */
    amoled_spi_send_data_simple(&colmod, 1);

    /* Memory Access Control: RGB order, normal orientation */
    amoled_spi_send_cmd(AMOLED_CMD_MADCTL);
    uint8_t madctl = AMOLED_MADCTL_RGB;
    amoled_spi_send_data_simple(&madctl, 1);

    /* Display On */
    amoled_spi_send_cmd(AMOLED_CMD_DISPON);
    for (volatile uint32_t d = 0; d < 120000U; d++) (void)d; /* 10ms */

    /* Clear frame buffer */
    amoled_display_clear(AMOLED_BLACK);

    log_info("AMOLED: ST7701S initialized (%ux%u, RGB565)",
             AMOLED_WIDTH, AMOLED_HEIGHT);
    return true;
}

/* ----------------------------------------------------------------
 * Frame buffer management
 * ---------------------------------------------------------------- */

void amoled_display_clear(uint16_t color)
{
    memset(s_fb_line, color, sizeof(s_fb_line));

    for (uint16_t y = 0; y < AMOLED_HEIGHT; y++) {
        /* Set column address */
        amoled_spi_send_cmd(AMOLED_CMD_CASET);
        uint8_t caset[4] = { 0, 0, (uint8_t)(AMOLED_WIDTH - 1), 0 };
        amoled_spi_send_data(caset, 4);

        /* Set row address */
        amoled_spi_send_cmd(AMOLED_CMD_RASET);
        uint8_t raset[4] = { 0, (uint8_t)y, 0, (uint8_t)(y + 1) };
        amoled_spi_send_data(raset, 4);

        /* Write pixel data */
        amoled_spi_send_cmd(AMOLED_CMD_RAMWR);
        amoled_spi_send_data((const uint8_t *)s_fb_line, AMOLED_WIDTH * 2);
    }
}

void amoled_display_update(void)
{
    /* Push full frame buffer to display */
    for (uint16_t y = 0; y < AMOLED_HEIGHT; y++) {
        amoled_spi_send_cmd(AMOLED_CMD_CASET);
        uint8_t caset[4] = { 0, 0, (uint8_t)(AMOLED_WIDTH - 1), 0 };
        amoled_spi_send_data(caset, 4);

        amoled_spi_send_cmd(AMOLED_CMD_RASET);
        uint8_t raset[4] = { 0, (uint8_t)y, 0, (uint8_t)(y + 1) };
        amoled_spi_send_data(raset, 4);

        amoled_spi_send_cmd(AMOLED_CMD_RAMWR);

        /* Pull line from frame buffer */
        uint16_t *line = &s_frame_buffer[y * AMOLED_WIDTH];
        amoled_spi_send_data((const uint8_t *)line, AMOLED_WIDTH * 2);
    }
}

/* ----------------------------------------------------------------
 * Drawing primitives
 * ---------------------------------------------------------------- */

void amoled_draw_rect_filled(uint16_t x0, uint16_t y0,
                             uint16_t x1, uint16_t y1, uint16_t color)
{
    if (x0 >= AMOLED_WIDTH || y0 >= AMOLED_HEIGHT) return;
    if (x1 >= AMOLED_WIDTH) x1 = AMOLED_WIDTH - 1;
    if (y1 >= AMOLED_HEIGHT) y1 = AMOLED_HEIGHT - 1;
    if (x1 < x0 || y1 < y0) return;

    for (uint16_t y = y0; y <= y1; y++) {
        for (uint16_t x = x0; x <= x1; x++) {
            s_frame_buffer[y * AMOLED_WIDTH + x] = color;
        }
    }
}

void amoled_draw_rect(uint16_t x0, uint16_t y0,
                      uint16_t w, uint16_t h, uint16_t color)
{
    amoled_draw_rect_filled(x0, y0, x0 + w - 1, y0, color);
    amoled_draw_rect_filled(x0, y0 + h - 1, x0 + w - 1, y0 + h - 1, color);
    amoled_draw_rect_filled(x0, y0, x0, y0 + h - 1, color);
    amoled_draw_rect_filled(x0 + w - 1, y0, x0 + w - 1, y0 + h - 1, color);
}

void amoled_draw_circle(uint16_t cx, uint16_t cy, uint16_t r, uint16_t color)
{
    /* Midpoint circle algorithm (filled) */
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

        /* Draw horizontal lines through each circle point */
        for (int16_t dy = -y; dy <= y; dy++) {
            amoled_draw_rect_filled(cx - x, (int16_t)cy + dy,
                                    cx + x, (int16_t)cy + dy, color);
        }
        for (int16_t dy = -x; dy <= x; dy++) {
            amoled_draw_rect_filled(cx - y, (int16_t)cy + dy,
                                    cx + y, (int16_t)cy + dy, color);
        }
    }
}

void amoled_draw_circle_outline(uint16_t cx, uint16_t cy,
                                uint16_t r, uint16_t color)
{
    /* Bresenham circle for outline only */
    int16_t x = (int16_t)r;
    int16_t y = 0;
    int16_t d = 3 - 2 * x;

    while (y <= x) {
        amoled_draw_rect_filled(cx + x, cy + y, cx + x, cy + y, color);
        amoled_draw_rect_filled(cx - x, cy + y, cx - x, cy + y, color);
        amoled_draw_rect_filled(cx + x, cy - y, cx + x, cy - y, color);
        amoled_draw_rect_filled(cx - x, cy - y, cx - x, cy - y, color);
        amoled_draw_rect_filled(cx + y, cy + x, cx + y, cy + x, color);
        amoled_draw_rect_filled(cx - y, cy + x, cx - y, cy + x, color);
        amoled_draw_rect_filled(cx + y, cy - x, cx + y, cy - x, color);
        amoled_draw_rect_filled(cx - y, cy - x, cx - y, cy - x, color);

        if (d > 0) {
            x--;
            d += 4 * (x - y) + 10;
        } else {
            d += 4 * y + 6;
        }
        y++;
    }
}

void amoled_draw_arc(uint16_t cx, uint16_t cy, uint16_t r,
                     uint16_t start_angle, uint16_t end_angle,
                     uint8_t thickness, uint16_t color)
{
    if (start_angle > end_angle) {
        /* Draw in two passes: start->360 and 0->end */
        amoled_draw_arc(cx, cy, r, start_angle, 360U, thickness, color);
        amoled_draw_arc(cx, cy, r, 0, end_angle, thickness, color);
        return;
    }

    for (uint16_t angle = start_angle; angle < end_angle; angle++) {
        float rad = angle * 3.14159265f / 180.0f;
        int16_t x_outer = (int16_t)(cx + r * cosf(rad));
        int16_t y_outer = (int16_t)(cy - r * sinf(rad));
        int16_t x_inner = (int16_t)(cx + (r - thickness) * cosf(rad));
        int16_t y_inner = (int16_t)(cy - (r - thickness) * sinf(rad));

        if (x_inner >= 0 && x_inner < AMOLED_WIDTH &&
            y_inner >= 0 && y_inner < AMOLED_HEIGHT) {
            amoled_draw_rect_filled((uint16_t)x_inner, (uint16_t)y_inner,
                                    (uint16_t)x_outer, (uint16_t)y_outer, color);
        }
    }
}

/* ----------------------------------------------------------------
 * Gradient helpers
 * ---------------------------------------------------------------- */

static uint16_t rgb565_lerp(uint16_t a, uint16_t b, float t)
{
    /* Extract R(5), G(6), B(5) components, lerp each, recombine */
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
    for (uint16_t x = x0; x <= x1; x++) {
        float t = (float)(x - x0) / (float)width;
        uint16_t color = rgb565_lerp(color_left, color_right, t);
        for (uint16_t y = y0; y <= y1; y++) {
            s_frame_buffer[y * AMOLED_WIDTH + x] = color;
        }
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
    for (uint16_t y = y0; y <= y1; y++) {
        float t = (float)(y - y0) / (float)height;
        uint16_t color = rgb565_lerp(color_top, color_bottom, t);
        for (uint16_t x = x0; x <= x1; x++) {
            s_frame_buffer[y * AMOLED_WIDTH + x] = color;
        }
    }
}

/* ----------------------------------------------------------------
 * Text rendering
 * ---------------------------------------------------------------- */

void amoled_draw_char(uint16_t x, uint16_t y, char ch,
                      uint16_t color, uint16_t bg_color)
{
    if (ch < 0x20 || ch > 0x7E) return; /* Skip non-printable */

    int idx = (ch - 0x20) * 5;
    uint16_t w = 5;
    uint16_t h = 7;

    if (x + w > AMOLED_WIDTH || y + h > AMOLED_HEIGHT) return;

    for (uint16_t row = 0; row < h; row++) {
        uint8_t pattern = amoled_font_5x7[idx + row];
        for (uint16_t col = 0; col < w; col++) {
            uint16_t c = (pattern & (1U << (w - 1 - col))) ? color : bg_color;
            s_frame_buffer[(y + row) * AMOLED_WIDTH + (x + col)] = c;
        }
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
            cursor_x += 6; /* 5px char + 1px gap */
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
    amoled_draw_rect_filled(x0, y, x1, y, color);
}

void amoled_draw_line_v(uint16_t x, uint16_t y0, uint16_t y1, uint16_t color)
{
    if (x >= AMOLED_WIDTH) return;
    if (y1 < y0) { uint16_t tmp = y0; y0 = y1; y1 = tmp; }
    if (y1 >= AMOLED_HEIGHT) y1 = AMOLED_HEIGHT - 1;
    amoled_draw_rect_filled(x, y0, x, y1, color);
}

void amoled_draw_line(uint16_t x0, uint16_t y0,
                      uint16_t x1, uint16_t y1, uint16_t color)
{
    /* Bresenham's line algorithm */
    int16_t dx = (int16_t)x1 - (int16_t)x0;
    int16_t dy = (int16_t)y1 - (int16_t)y0;
    bool steep = abs(dy) > abs(dx);

    int16_t sx = dx > 0 ? 1 : -1;
    int16_t sy = dy > 0 ? 1 : -1;

    int16_t xx = steep ? dy : dx;
    int16_t xy = steep ? dx : dy;

    if (xx < 0) {
        xx = -xx;
        xy = -xy;
    }

    int16_t y = 0;
    int16_t error = xy / 2;

    for (int16_t i = 0; i <= xx; i++) {
        int16_t px = steep ? y0 + y * sy : x0 + i * sx;
        int16_t py = steep ? x0 + i * sx : y0 + y * sy;

        if (px >= 0 && px < AMOLED_WIDTH && py >= 0 && py < AMOLED_HEIGHT) {
            s_frame_buffer[py * AMOLED_WIDTH + px] = color;
        }

        error -= xy;
        if (error < 0) {
            y += sy;
            error += xx;
        }
    }
}

/* ----------------------------------------------------------------
 * Scroll and queries
 * ---------------------------------------------------------------- */

void amoled_set_scroll_area(uint16_t top, uint16_t bot)
{
    /* TFA: Top Fixed Area */
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
    /* SSD: Scrolling Display Area command */
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
