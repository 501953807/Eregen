/*
 * Eregen (颐贞) - LED Control Test
 * Host-compiled tests with mocked LEDC functions
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <stdint.h>

/* Mock esp_err_t */
typedef int esp_err_t;
#define ESP_OK        0
#define ESP_FAIL     -1

/* ---- Mock LEDC state ---- */
typedef struct {
    uint8_t duty[3];       /* Red, Green, Blue channels */
    bool timer_configured;
    bool channels_configured[3];
    int set_duty_calls[3];
} ledc_state_t;

static ledc_state_t s_ledc;
static bool s_led_initialized = false;

/* Reset mock state */
static void reset_ledc(void)
{
    memset(&s_ledc, 0, sizeof(s_ledc));
    s_led_initialized = false;
}

/* ---- Mock LEDC functions ---- */

esp_err_t mock_ledc_timer_config(uint8_t duty_res, uint32_t freq,
                                   uint8_t timer, uint8_t speed_mode)
{
    (void)duty_res; (void)freq; (void)speed_mode;
    s_ledc.timer_configured = true;
    return ESP_OK;
}

esp_err_t mock_ledc_channel_config(uint8_t channel, uint8_t gpio_num,
                                    uint8_t timer_sel)
{
    if (channel < 3) {
        s_ledc.channels_configured[channel] = true;
    }
    return ESP_OK;
}

void mock_ledc_set_duty(uint8_t channel, uint8_t duty)
{
    if (channel < 3) {
        s_ledc.duty[channel] = duty;
        s_ledc.set_duty_calls[channel]++;
    }
}

void mock_ledc_update_duty(uint8_t channel)
{
    (void)channel;
    /* In real hardware this pushes the duty to the hardware register */
}

/* ---- Functions under test (inline implementations for testing) ---- */

typedef enum {
    LED_COLOR_RED,
    LED_COLOR_GREEN,
    LED_COLOR_BLUE,
    LED_COLOR_OFF
} led_color_t;

typedef enum {
    LED_PATTERN_SOLID,
    LED_PATTERN_SLOW_BLINK,
    LED_PATTERN_FAST_BLINK,
    LED_PATTERN_PULSE
} led_pattern_t;

static const uint8_t color_duty[3][3] = {
    {255,   0,     0},     /* RED    */
    {0,     255,   0},     /* GREEN  */
    {0,     0,     255},   /* BLUE   */
};

static esp_err_t test_led_init(void)
{
    reset_ledc();

    /* Simulate LED initialization */
    mock_ledc_timer_config(8, 5000, 0, 0);
    mock_ledc_channel_config(0, 3, 0);  /* Red */
    mock_ledc_channel_config(1, 4, 0);  /* Green */
    mock_ledc_channel_config(2, 5, 0);  /* Blue */

    s_led_initialized = true;
    return ESP_OK;
}

static esp_err_t test_led_set_color(led_color_t color)
{
    if (!s_led_initialized) return ESP_FAIL;

    uint8_t r = 0, g = 0, b = 0;
    if (color != LED_COLOR_OFF) {
        r = color_duty[color][0];
        g = color_duty[color][1];
        b = color_duty[color][2];
    }

    mock_ledc_set_duty(0, r);
    mock_ledc_set_duty(1, g);
    mock_ledc_set_duty(2, b);
    mock_ledc_update_duty(0);
    mock_ledc_update_duty(1);
    mock_ledc_update_duty(2);

    return ESP_OK;
}

static esp_err_t test_led_off(void)
{
    return test_led_set_color(LED_COLOR_OFF);
}

static esp_err_t test_led_blink(led_color_t color, led_pattern_t pattern)
{
    (void)pattern;
    return test_led_set_color(color);
}

/* ---- Test helpers ---- */
static int tests_run = 0;
static int tests_passed = 0;
static int tests_failed = 0;

#define TEST(name) \
    static void test_##name(void); \
    static void run_##name(void) { \
        printf("  TEST: %s\n", #name); \
        test_##name(); \
    } \
    static void test_##name(void)

#define ASSERT_TRUE(cond) do { \
    tests_run++; \
    if (cond) { tests_passed++; } \
    else { tests_failed++; printf("    FAIL: %s at line %d\n", #cond, __LINE__); } \
} while(0)

#define ASSERT_FALSE(cond) ASSERT_TRUE(!(cond))
#define ASSERT_EQ(a, b) ASSERT_TRUE((a) == (b))
#define ASSERT_NEQ(a, b) ASSERT_TRUE((a) != (b))

/* ---- Tests ---- */

TEST(led_init_configures_timer_and_channels)
{
    reset_ledc();
    esp_err_t ret = test_led_init();

    ASSERT_EQ(ret, ESP_OK);
    ASSERT_TRUE(s_led_initialized);
    ASSERT_TRUE(s_ledc.timer_configured);
    ASSERT_TRUE(s_ledc.channels_configured[0]);
    ASSERT_TRUE(s_ledc.channels_configured[1]);
    ASSERT_TRUE(s_ledc.channels_configured[2]);
}

TEST(led_set_color_green)
{
    test_led_init();
    test_led_set_color(LED_COLOR_GREEN);

    ASSERT_EQ(s_ledc.duty[0], 0);     /* Red off */
    ASSERT_EQ(s_ledc.duty[1], 255);   /* Green full */
    ASSERT_EQ(s_ledc.duty[2], 0);     /* Blue off */
}

TEST(led_set_color_red)
{
    test_led_init();
    test_led_set_color(LED_COLOR_RED);

    ASSERT_EQ(s_ledc.duty[0], 255);   /* Red full */
    ASSERT_EQ(s_ledc.duty[1], 0);     /* Green off */
    ASSERT_EQ(s_ledc.duty[2], 0);     /* Blue off */
}

TEST(led_set_color_blue)
{
    test_led_init();
    test_led_set_color(LED_COLOR_BLUE);

    ASSERT_EQ(s_ledc.duty[0], 0);     /* Red off */
    ASSERT_EQ(s_ledc.duty[1], 0);     /* Green off */
    ASSERT_EQ(s_ledc.duty[2], 255);   /* Blue full */
}

TEST(led_set_color_off)
{
    test_led_init();
    test_led_set_color(LED_COLOR_OFF);

    ASSERT_EQ(s_ledc.duty[0], 0);
    ASSERT_EQ(s_ledc.duty[1], 0);
    ASSERT_EQ(s_ledc.duty[2], 0);
}

TEST(led_blink_sets_color)
{
    test_led_init();
    test_led_blink(LED_COLOR_RED, LED_PATTERN_SLOW_BLINK);

    ASSERT_EQ(s_ledc.duty[0], 255);
    ASSERT_EQ(s_ledc.duty[1], 0);
    ASSERT_EQ(s_ledc.duty[2], 0);
}

TEST(led_off_turns_all_channels_off)
{
    test_led_init();
    test_led_off();

    ASSERT_EQ(s_ledc.duty[0], 0);
    ASSERT_EQ(s_ledc.duty[1], 0);
    ASSERT_EQ(s_ledc.duty[2], 0);
}

TEST(led_set_color_before_init_fails)
{
    reset_ledc();
    /* Don't call init — led_set_color should fail */
    esp_err_t ret = test_led_set_color(LED_COLOR_GREEN);
    /* Our inline impl returns ESP_FAIL when not initialized */
    ASSERT_NEQ(ret, ESP_OK);
}

TEST(led_duty_updates_are_pushed)
{
    test_led_init();
    test_led_set_color(LED_COLOR_RED);   /* Triggers mock_ledc_set_duty */

    /* Verify that set_duty was called for each channel */
    ASSERT_TRUE(s_ledc.set_duty_calls[0] > 0);
    ASSERT_TRUE(s_ledc.set_duty_calls[1] > 0);
    ASSERT_TRUE(s_ledc.set_duty_calls[2] > 0);
}

TEST(all_three_colors_produce_unique_patterns)
{
    test_led_init();

    test_led_set_color(LED_COLOR_RED);
    uint8_t red_r = s_ledc.duty[0];
    uint8_t red_g = s_ledc.duty[1];
    uint8_t red_b = s_ledc.duty[2];

    test_led_set_color(LED_COLOR_GREEN);
    uint8_t grn_r = s_ledc.duty[0];
    uint8_t grn_g = s_ledc.duty[1];
    uint8_t grn_b = s_ledc.duty[2];

    test_led_set_color(LED_COLOR_BLUE);
    uint8_t blu_r = s_ledc.duty[0];
    uint8_t blu_g = s_ledc.duty[1];
    uint8_t blu_b = s_ledc.duty[2];

    /* Each color should have exactly one channel at 255 */
    ASSERT_EQ(red_r + red_g + red_b, 255);
    ASSERT_EQ(grn_r + grn_g + grn_b, 255);
    ASSERT_EQ(blu_r + blu_g + blu_b, 255);

    /* Red and green should differ */
    ASSERT_NEQ(red_r, grn_r);
    ASSERT_NEQ(red_g, grn_g);
}

/* ---- Main ---- */

int main(void)
{
    printf("\n=== LED Control Tests ===\n\n");

    run_led_init_configures_timer_and_channels();
    run_led_set_color_green();
    run_led_set_color_red();
    run_led_set_color_blue();
    run_led_set_color_off();
    run_led_blink_sets_color();
    run_led_off_turns_all_channels_off();
    run_led_set_color_before_init_fails();
    run_led_duty_updates_are_pushed();
    run_all_three_colors_produce_unique_patterns();

    printf("\n=== Results: %d passed, %d failed, %d total ===\n",
           tests_passed, tests_failed, tests_run);

    return tests_failed > 0 ? 1 : 0;
}
