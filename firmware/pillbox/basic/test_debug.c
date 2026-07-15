#include <stdio.h>
#include <string.h>
#include <stdint.h>

typedef int esp_err_t;
#define ESP_OK 0

typedef struct {
    uint8_t duty[3];
    bool initialized;
} ledc_state_t;

static ledc_state_t s_ledc;

static void reset_ledc(void) {
    memset(&s_ledc, 0, sizeof(s_ledc));
    s_ledc.initialized = false;
}

static void mock_ledc_set_duty(uint8_t channel, uint8_t duty) {
    printf("    mock_ledc_set_duty(ch=%d, duty=%d)\n", channel, duty);
    if (channel < 3) {
        s_ledc.duty[channel] = duty;
    }
}

typedef enum { LED_COLOR_GREEN = 0, LED_COLOR_RED = 1, LED_COLOR_BLUE = 2, LED_COLOR_OFF = 3 } led_color_t;

static const uint8_t color_duty[3][3] = {
    {255, 0, 0},   /* RED */
    {0, 255, 0},   /* GREEN */
    {0, 0, 255},   /* BLUE */
};

static void test_led_init(void) {
    reset_ledc();
    s_ledc.initialized = true;
    printf("After init: duty=[%d,%d,%d]\n", s_ledc.duty[0], s_ledc.duty[1], s_ledc.duty[2]);
}

static void test_led_set_color(led_color_t color) {
    if (!s_ledc.initialized) { printf("ERROR: not initialized\n"); return; }
    
    uint8_t r = 0, g = 0, b = 0;
    if (color != LED_COLOR_OFF) {
        r = color_duty[color][0];
        g = color_duty[color][1];
        b = color_duty[color][2];
    }
    printf("Setting color: r=%d g=%d b=%d (color=%d)\n", r, g, b, color);
    
    mock_ledc_set_duty(0, r);
    mock_ledc_set_duty(1, g);
    mock_ledc_set_duty(2, b);
}

int main(void) {
    test_led_init();
    printf("\n--- Test GREEN ---\n");
    test_led_set_color(LED_COLOR_GREEN);
    printf("Result: duty=[%d,%d,%d]\n", s_ledc.duty[0], s_ledc.duty[1], s_ledc.duty[2]);
    
    printf("\n--- Test RED ---\n");
    test_led_init();
    test_led_set_color(LED_COLOR_RED);
    printf("Result: duty=[%d,%d,%d]\n", s_ledc.duty[0], s_ledc.duty[1], s_ledc.duty[2]);
    
    printf("\n--- Test BLUE ---\n");
    test_led_init();
    test_led_set_color(LED_COLOR_BLUE);
    printf("Result: duty=[%d,%d,%d]\n", s_ledc.duty[0], s_ledc.duty[1], s_ledc.duty[2]);
    
    return 0;
}
