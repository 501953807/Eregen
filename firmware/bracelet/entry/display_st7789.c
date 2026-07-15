/*
 * Eregen (颐贞) - Display ST7789 Driver Implementation
 * 1.14" IPS LCD 135×240 via SPI
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "display_st7789.h"
#include "gd32e230_spi.h"
#include "gd32e230_gpio.h"
#include "gd32e230_rcu.h"
#include <string.h>

/* Built-in 5×7 pixel font bitmap (ASCII 32-126) */
const uint8_t display_font_5x7[] = {
    /* 0x20 ' ' (space) */
    0x00, 0x00, 0x00, 0x00, 0x00,
    /* 0x21 '!' */
    0x00, 0x00, 0x5F, 0x00, 0x00,
    /* 0x22 '"' */
    0x00, 0x07, 0x00, 0x07, 0x00,
    /* 0x23 '#' */
    0x14, 0x7F, 0x14, 0x7F, 0x14,
    /* 0x24 '$' */
    0x24, 0x2A, 0x7F, 0x2A, 0x12,
    /* 0x25 '%' */
    0x23, 0x13, 0x08, 0x64, 0x62,
    /* 0x26 '&' */
    0x36, 0x49, 0x55, 0x22, 0x50,
    /* 0x27 ''' */
    0x00, 0x05, 0x03, 0x00, 0x00,
    /* 0x28 '(' */
    0x00, 0x1C, 0x22, 0x41, 0x00,
    /* 0x29 ')' */
    0x00, 0x41, 0x22, 0x1C, 0x00,
    /* 0x2A '*' */
    0x14, 0x08, 0x3E, 0x08, 0x14,
    /* 0x2B '+' */
    0x08, 0x08, 0x3E, 0x08, 0x08,
    /* 0x2C ',' */
    0x00, 0x50, 0x30, 0x00, 0x00,
    /* 0x2D '-' */
    0x08, 0x08, 0x08, 0x08, 0x08,
    /* 0x2E '.' */
    0x00, 0x60, 0x60, 0x00, 0x00,
    /* 0x2F '/' */
    0x20, 0x10, 0x08, 0x04, 0x02,
    /* 0x30 '0' */
    0x3E, 0x51, 0x49, 0x45, 0x3E,
    /* 0x31 '1' */
    0x00, 0x42, 0x7F, 0x40, 0x00,
    /* 0x32 '2' */
    0x42, 0x61, 0x51, 0x49, 0x46,
    /* 0x33 '3' */
    0x21, 0x41, 0x45, 0x4B, 0x31,
    /* 0x34 '4' */
    0x18, 0x14, 0x12, 0x7F, 0x10,
    /* 0x35 '5' */
    0x27, 0x45, 0x45, 0x45, 0x39,
    /* 0x36 '6' */
    0x3C, 0x4A, 0x49, 0x49, 0x30,
    /* 0x37 '7' */
    0x01, 0x71, 0x09, 0x05, 0x03,
    /* 0x38 '8' */
    0x36, 0x49, 0x49, 0x49, 0x36,
    /* 0x39 '9' */
    0x06, 0x49, 0x49, 0x29, 0x1E,
    /* 0x3A ':' */
    0x00, 0x36, 0x36, 0x00, 0x00,
    /* 0x3B ';' */
    0x00, 0x56, 0x36, 0x00, 0x00,
    /* 0x3C '<' */
    0x08, 0x14, 0x22, 0x41, 0x00,
    /* 0x3D '=' */
    0x14, 0x14, 0x14, 0x14, 0x14,
    /* 0x3E '>' */
    0x00, 0x41, 0x22, 0x14, 0x08,
    /* 0x3F '?' */
    0x02, 0x01, 0x51, 0x09, 0x06,
    /* 0x40 '@' (skipped) */
    /* 0x41 'A' */
    0x32, 0x49, 0x79, 0x41, 0x3E,
    /* 0x42 'B' */
    0x7E, 0x09, 0x09, 0x09, 0x7E,
    /* 0x43 'C' */
    0x7E, 0x40, 0x40, 0x40, 0x40,
    /* 0x44 'D' */
    0x7E, 0x01, 0x01, 0x19, 0x7E,
    /* 0x45 'E' */
    0x7F, 0x49, 0x49, 0x49, 0x41,
    /* 0x46 'F' */
    0x7F, 0x09, 0x09, 0x09, 0x01,
    /* 0x47 'G' */
    0x7E, 0x01, 0x01, 0x19, 0x7E,
    /* 0x48 'H' */
    0x7F, 0x08, 0x08, 0x08, 0x7F,
    /* 0x49 'I' */
    0x00, 0x41, 0x7F, 0x41, 0x00,
    /* 0x4A 'J' */
    0x20, 0x40, 0x41, 0x3F, 0x01,
    /* 0x4B 'K' */
    0x7F, 0x08, 0x14, 0x22, 0x41,
    /* 0x4C 'L' */
    0x7F, 0x40, 0x40, 0x40, 0x40,
    /* 0x4D 'M' */
    0x7F, 0x02, 0x0C, 0x02, 0x7F,
    /* 0x4E 'N' */
    0x7F, 0x04, 0x08, 0x10, 0x7F,
    /* 0x4F 'O' */
    0x3E, 0x41, 0x41, 0x41, 0x3E,
    /* 0x50 'P' */
    0x7F, 0x09, 0x09, 0x09, 0x01,
    /* 0x51 'Q' */
    0x3E, 0x41, 0x51, 0x21, 0x5E,
    /* 0x52 'R' */
    0x7F, 0x09, 0x19, 0x29, 0x46,
    /* 0x53 'S' */
    0x46, 0x49, 0x49, 0x49, 0x31,
    /* 0x54 'T' */
    0x01, 0x01, 0x7F, 0x01, 0x01,
    /* 0x55 'U' */
    0x3F, 0x40, 0x40, 0x40, 0x3F,
    /* 0x56 'V' */
    0x1F, 0x20, 0x40, 0x20, 0x1F,
    /* 0x57 'W' */
    0x3F, 0x40, 0x38, 0x40, 0x3F,
    /* 0x58 'X' */
    0x63, 0x14, 0x08, 0x14, 0x63,
    /* 0x59 'Y' */
    0x07, 0x08, 0x70, 0x08, 0x07,
    /* 0x5A 'Z' */
    0x61, 0x51, 0x49, 0x45, 0x43,
    /* 0x5B '[' (skipped) */
    /* 0x5C '\' (skipped) */
    /* 0x5D ']' (skipped) */
    /* 0x5E '^' */
    0x20, 0x54, 0x54, 0x54, 0x78,
    /* 0x5F '_' */
    0x7F, 0x48, 0x48, 0x48, 0x48,
    /* 0x60 '`' (skipped) */
    /* 0x61 'a' */
    0x00, 0x62, 0x64, 0x38, 0x10,
    /* 0x62 'b' */
    0x78, 0x46, 0x41, 0x41, 0x3F,
    /* 0x63 'c' */
    0x00, 0x7C, 0x08, 0x04, 0x78,
    /* 0x64 'd' */
    0x7C, 0x08, 0x04, 0x78, 0x00,
    /* 0x65 'e' */
    0x00, 0x38, 0x54, 0x54, 0x54,
    /* 0x66 'f' */
    0x20, 0x74, 0x22, 0x01, 0x00,
    /* 0x67 'g' */
    0x00, 0x41, 0x7E, 0x42, 0x00,
    /* 0x68 'h' */
    0x7F, 0x08, 0x04, 0x04, 0x78,
    /* 0x69 'i' */
    0x00, 0x44, 0x7D, 0x40, 0x00,
    /* 0x6A 'j' */
    0x00, 0x40, 0x7C, 0x04, 0x00,
    /* 0x6B 'k' */
    0x7F, 0x08, 0x14, 0x22, 0x41,
    /* 0x6C 'l' */
    0x7F, 0x40, 0x40, 0x40, 0x40,
    /* 0x6D 'm' */
    0x00, 0x7C, 0x04, 0x18, 0x04,
    /* 0x6E 'n' */
    0x78, 0x04, 0x04, 0x78, 0x00,
    /* 0x6F 'o' */
    0x00, 0x38, 0x44, 0x44, 0x38,
    /* 0x70 'p' */
    0x7C, 0x14, 0x14, 0x04, 0x00,
    /* 0x71 'q' */
    0x00, 0x14, 0x14, 0x7C, 0x00,
    /* 0x72 'r' */
    0x00, 0x7C, 0x08, 0x04, 0x04,
    /* 0x73 's' */
    0x00, 0x5C, 0x30, 0x04, 0x00,
    /* 0x74 't' */
    0x08, 0x7E, 0x08, 0x00, 0x00,
    /* 0x75 'u' */
    0x00, 0x3C, 0x40, 0x40, 0x3C,
    /* 0x76 'v' */
    0x00, 0x1F, 0x20, 0x40, 0x20,
    /* 0x77 'w' */
    0x00, 0x3E, 0x40, 0x40, 0x3E,
    /* 0x78 'x' */
    0x00, 0x7C, 0x08, 0x04, 0x7C,
    /* 0x79 'y' */
    0x00, 0x7C, 0x08, 0x04, 0x78,
    /* 0x7A 'z' */
    0x00, 0x44, 0x7D, 0x40, 0x00,
};

/* Display buffer (framebuffer in RGB565) */
static uint16_t s_display_buf[DISPLAY_WIDTH * DISPLAY_HEIGHT];

/* Dirty rectangle tracking for partial updates */
static bool s_dirty = false;

/*
 * Send a command byte to ST7789 via SPI.
 */
static void display_send_cmd(uint8_t cmd)
{
    gpio_bit_reset(DISPLAY_CS_GPIO_PORT, DISPLAY_CS_GPIO_PIN);
    gpio_bit_reset(GPIOB, GPIO_PIN_15);  /* DC pin low for command */

    while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
        /* Wait */
    }
    spi_i2s_data_transmit(SPI1, cmd);
    while (spi_flag_get(SPI1, SPI_FLAG_BSY) != RESET) {
        /* Wait */
    }

    gpio_bit_set(GPIOB, GPIO_PIN_15);  /* DC pin high */
    gpio_bit_set(DISPLAY_CS_GPIO_PORT, DISPLAY_CS_GPIO_PIN);
}

/*
 * Send a data byte to ST7789 via SPI.
 */
static void display_send_data(uint8_t data)
{
    gpio_bit_reset(DISPLAY_CS_GPIO_PORT, DISPLAY_CS_GPIO_PIN);
    gpio_bit_set(GPIOB, GPIO_PIN_15);  /* DC pin high for data */

    while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
        /* Wait */
    }
    spi_i2s_data_transmit(SPI1, data);
    while (spi_flag_get(SPI1, SPI_FLAG_BSY) != RESET) {
        /* Wait */
    }

    gpio_bit_set(DISPLAY_CS_GPIO_PORT, DISPLAY_CS_GPIO_PIN);
}

/*
 * Send multiple data bytes to ST7789.
 */
static void display_send_data_multi(const uint8_t *data, uint16_t len)
{
    gpio_bit_reset(DISPLAY_CS_GPIO_PORT, DISPLAY_CS_GPIO_PIN);
    gpio_bit_set(GPIOB, GPIO_PIN_15);

    for (uint16_t i = 0; i < len; i++) {
        while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
            /* Wait */
        }
        spi_i2s_data_transmit(SPI1, data[i]);
    }
    while (spi_flag_get(SPI1, SPI_FLAG_BSY) != RESET) {
        /* Wait */
    }

    gpio_bit_set(DISPLAY_CS_GPIO_PORT, DISPLAY_CS_GPIO_PORT);
}

/*
 * Initialize the ST7789 display.
 * Sends the standard initialization sequence.
 */
bool display_init(void)
{
    rcu_periph_clock_enable(RCU_GPIOB);
    rcu_periph_clock_enable(RCU_SPI1);

    /* PB15 as GPIO output for DC pin */
    gpio_init(GPIOB, GPIO_MODE_OUT_PP, GPIO_OSPEED_50MHZ, GPIO_PIN_15);
    gpio_bit_set(GPIOB, GPIO_PIN_15);

    /* Hardware reset */
    gpio_bit_reset(GPIOB, GPIO_PIN_15);
    for (volatile int i = 0; i < 100000; i++);  /* ~1ms delay */
    gpio_bit_set(GPIOB, GPIO_PIN_15);
    for (volatile int i = 0; i < 100000; i++);  /* ~10ms delay */

    /* Software reset */
    display_send_cmd(ST7789_SWRESET);
    for (volatile int i = 0; i < 500000; i++);  /* 120ms delay */

    /* Sleep out */
    display_send_cmd(ST7789_NORON);
    for (volatile int i = 0; i < 500000; i++);  /* 120ms delay */

    /* Memory Access Control: RGB order, column/row order */
    display_send_cmd(ST7789_MADCTL);
    display_send_data(0x00U);  /* Row address order, normal col */

    /* Pixel format: 16-bit RGB565 */
    display_send_cmd(ST7789_COLMOD);
    display_send_data(0x55U);  /* 16 bits per pixel */

    /* Column address set */
    display_send_cmd(ST7789_CASET);
    display_send_data(0x00U);
    display_send_data(0x00U);
    display_send_data(0x00U);
    display_send_data(0x87U);  /* 135 pixels (0-134) */

    /* Row address set */
    display_send_cmd(ST7789_RASET);
    display_send_data(0x00U);
    display_send_data(0x00U);
    display_send_data(0x00U);
    display_send_data(0xF0U);  /* 240 pixels (0-239) */

    /* Display on */
    display_send_cmd(ST7789_DISPON);

    /* Clear display buffer */
    display_clear(DISPLAY_COLOR_BLACK);

    return true;
}

/*
 * Clear the entire display to a solid color.
 */
void display_clear(uint16_t color)
{
    uint32_t len = (uint32_t)DISPLAY_WIDTH * DISPLAY_HEIGHT;
    for (uint32_t i = 0; i < len; i++) {
        s_display_buf[i] = color;
    }
    s_dirty = true;
}

/*
 * Draw a single 5×7 character at the given position.
 */
void display_draw_char(uint16_t x, uint16_t y, char ch,
                       uint16_t color, uint16_t bg_color)
{
    if (ch < 0x20 || ch > 0x7E) {
        return;  /* Only printable ASCII */
    }

    uint8_t idx = (uint8_t)(ch - 0x20);
    const uint8_t *font = &display_font_5x7[idx * 5];

    for (uint8_t col = 0; col < 5; col++) {
        uint8_t pattern = font[col];
        for (uint8_t row = 0; row < 7; row++) {
            int16_t px = (int16_t)x + col;
            int16_t py = (int16_t)y + row;
            if (px < 0 || px >= DISPLAY_WIDTH || py < 0 || py >= DISPLAY_HEIGHT) {
                continue;
            }
            uint16_t pixel_color = (pattern & (1U << row)) ? color : bg_color;
            s_display_buf[(uint32_t)py * DISPLAY_WIDTH + (uint32_t)px] = pixel_color;
        }
    }
    s_dirty = true;
}

/*
 * Draw a null-terminated string at the given position.
 */
void display_draw_string(uint16_t x, uint16_t y, const char *str,
                         uint16_t color, uint16_t bg_color)
{
    if (!str) return;

    uint16_t cx = x;
    while (*str && cx < DISPLAY_WIDTH) {
        display_draw_char(cx, y, *str, color, bg_color);
        cx += 6;  /* 5 pixels wide + 1 pixel gap */
        str++;
    }
}

/*
 * Draw a filled circle using midpoint algorithm.
 */
void display_draw_circle(uint16_t cx, uint16_t cy, uint8_t r,
                         uint16_t color)
{
    int16_t x = (int16_t)r;
    int16_t y = 0;
    int16_t err = 0;

    while (x >= y) {
        /* Draw horizontal lines through the circle points */
        for (int16_t i = cx - (int16_t)x; i <= cx + x; i++) {
            if (i >= 0 && i < DISPLAY_WIDTH) {
                if (cy + y >= 0 && cy + y < DISPLAY_HEIGHT)
                    s_display_buf[(uint32_t)(cy + y) * DISPLAY_WIDTH + (uint32_t)i] = color;
                if (cy - y >= 0 && cy - y < DISPLAY_HEIGHT)
                    s_display_buf[(uint32_t)(cy - y) * DISPLAY_WIDTH + (uint32_t)i] = color;
            }
        }
        for (int16_t i = cx - (int16_t)y; i <= cx + y; i++) {
            if (i >= 0 && i < DISPLAY_WIDTH) {
                if (cy + x >= 0 && cy + x < DISPLAY_HEIGHT)
                    s_display_buf[(uint32_t)(cy + x) * DISPLAY_WIDTH + (uint32_t)i] = color;
                if (cy - x >= 0 && cy - x < DISPLAY_HEIGHT)
                    s_display_buf[(uint32_t)(cy - x) * DISPLAY_WIDTH + (uint32_t)i] = color;
            }
        }

        y++;
        err += 1 + 2 * y;
        if (2 * (err - x) + 1 > 0 && x > 0) {
            x--;
            err += 1 - 2 * x;
        }
    }
    s_dirty = true;
}

/*
 * Draw a hollow circle outline.
 */
void display_draw_circle_outline(uint16_t cx, uint16_t cy, uint8_t r,
                                 uint16_t color)
{
    int16_t x = (int16_t)r;
    int16_t y = 0;
    int16_t err = 0;

    while (x >= y) {
        /* Plot circle points */
        if (cx + x < DISPLAY_WIDTH && cy + y < DISPLAY_HEIGHT)
            s_display_buf[(cy + y) * DISPLAY_WIDTH + cx + x] = color;
        if (cx - x >= 0 && cy + y < DISPLAY_HEIGHT)
            s_display_buf[(cy + y) * DISPLAY_WIDTH + cx - x] = color;
        if (cx + x < DISPLAY_WIDTH && cy - y >= 0)
            s_display_buf[(cy - y) * DISPLAY_WIDTH + cx + x] = color;
        if (cx - x >= 0 && cy - y >= 0)
            s_display_buf[(cy - y) * DISPLAY_WIDTH + cx - x] = color;

        if (cx + y < DISPLAY_WIDTH && cy + x < DISPLAY_HEIGHT)
            s_display_buf[(cy + x) * DISPLAY_WIDTH + cx + y] = color;
        if (cx - y >= 0 && cy + x < DISPLAY_HEIGHT)
            s_display_buf[(cy + x) * DISPLAY_WIDTH + cx - y] = color;
        if (cx + y < DISPLAY_WIDTH && cy - x >= 0)
            s_display_buf[(cy - x) * DISPLAY_WIDTH + cx + y] = color;
        if (cx - y >= 0 && cy - x >= 0)
            s_display_buf[(cy - x) * DISPLAY_WIDTH + cx - y] = color;

        y++;
        err += 1 + 2 * y;
        if (2 * (err - x) + 1 > 0 && x > 0) {
            x--;
            err += 1 - 2 * x;
        }
    }
    s_dirty = true;
}

/*
 * Draw a filled rectangle.
 */
void display_draw_rect_filled(uint16_t x0, uint16_t y0,
                              uint16_t x1, uint16_t y1,
                              uint16_t color)
{
    if (x0 > x1) { uint16_t t = x0; x0 = x1; x1 = t; }
    if (y0 > y1) { uint16_t t = y0; y0 = y1; y1 = t; }

    for (uint16_t y = y0; y <= y1 && y < DISPLAY_HEIGHT; y++) {
        for (uint16_t x = x0; x <= x1 && x < DISPLAY_WIDTH; x++) {
            s_display_buf[y * DISPLAY_WIDTH + x] = color;
        }
    }
    s_dirty = true;
}

/*
 * Push the display buffer to the screen.
 * Uses window addressing for efficient partial updates.
 */
void display_update(void)
{
    if (!s_dirty) {
        return;
    }

    /* Set column address */
    display_send_cmd(ST7789_CASET);
    display_send_data(0x00U);
    display_send_data(0x00U);
    display_send_data(0x00U);
    display_send_data(0x87U);

    /* Set row address */
    display_send_cmd(ST7789_RASET);
    display_send_data(0x00U);
    display_send_data(0x00U);
    display_send_data(0x00U);
    display_send_data(0xF0U);

    /* Write frame buffer to RAM */
    display_send_cmd(ST7789_RAMWR);

    gpio_bit_reset(DISPLAY_CS_GPIO_PORT, DISPLAY_CS_GPIO_PIN);
    gpio_bit_set(GPIOB, GPIO_PIN_15);

    /* Transfer in chunks to avoid large SPI transfers */
    const uint16_t CHUNK = 320U;
    uint32_t total = (uint32_t)DISPLAY_WIDTH * DISPLAY_HEIGHT;
    for (uint32_t i = 0; i < total; i += CHUNK) {
        uint16_t count = (i + CHUNK > total) ? (total - i) : CHUNK;
        for (uint16_t j = 0; j < count; j++) {
            uint16_t pixel = s_display_buf[i + j];
            /* Send high byte first (big-endian for ST7789) */
            uint8_t hi = (uint8_t)(pixel >> 8);
            uint8_t lo = (uint8_t)(pixel & 0xFF);

            while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
                /* Wait */
            }
            spi_i2s_data_transmit(SPI1, hi);

            while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
                /* Wait */
            }
            spi_i2s_data_transmit(SPI1, lo);
        }
    }

    while (spi_flag_get(SPI1, SPI_FLAG_BSY) != RESET) {
        /* Wait */
    }

    gpio_bit_set(DISPLAY_CS_GPIO_PORT, DISPLAY_CS_GPIO_PIN);
    s_dirty = false;
}

/*
 * Set scroll area for partial screen updates.
 */
void display_set_scroll_area(uint16_t top, uint16_t bot)
{
    /* TFA: Top Fixed Area */
    display_send_cmd(0x33U);  /* VSCRDEF */
    display_send_data((uint8_t)(top >> 8));
    display_send_data((uint8_t)top);
    /* VSA: Vertical Scrolling Area */
    display_send_data((uint8_t)((DISPLAY_HEIGHT - top - bot) >> 8));
    display_send_data((uint8_t)(DISPLAY_HEIGHT - top - bot));
    /* BFA: Bottom Fixed Area */
    display_send_data((uint8_t)(bot >> 8));
    display_send_data((uint8_t)bot);
}
