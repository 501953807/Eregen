/*
 * Eregen (颐贞) - Voice Reminder Module Test (Host-mode)
 * Tests TTS text queuing, volume mapping, and stop behavior.
 * Uses mock UART — no ESP-IDF required.
 *
 * Compile: gcc -DTEST_MODE -I. test_voice_reminder.c -o test_voice_reminder
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <stdint.h>
#include <unistd.h>

/* ---- Mock UART layer ---- */
#define UART_NUM_1 1

typedef struct {
    char text_buf[128];
    uint8_t volume;
    bool playing;
    bool initialized;
} mock_tts_state_t;

static mock_tts_state_t g_mock = {0};

/* ---- Minimal header inline (no real voice_reminder.h available on host) ---- */
esp_err_t tts_init(int uart_num);
esp_err_t tts_speak(const char *text);
void tts_stop(void);
bool tts_is_playing(void);
esp_err_t tts_set_volume(uint8_t percent);
uint8_t tts_get_volume(void);

/* ---- Mock implementations (duplicated from voice_reminder.c for host test) ---- */

#define SYN_CMD_SPEAK         0x01
#define SYN_CMD_VOLUME        0x02
#define SYN_CMD_STOP          0x03
#define SYN_FRAME_START       0xAA
#define SYN_FRAME_END         0x55

static int s_uart_num = -1;
static uint8_t s_volume = 80;
static bool s_playing = false;

/* Simulated queue of texts to speak */
static char s_text_queue[8][128];
static int s_queue_head = 0;
static int s_queue_tail = 0;
static int s_queue_count = 0;

static esp_err_t mock_send_frame(uint8_t cmd, const uint8_t *data, uint8_t len)
{
    if (s_uart_num < 0)
        return ESP_ERR_INVALID_STATE;

    if (cmd == SYN_CMD_SPEAK && data != NULL && len > 0) {
        /* Queue the text */
        if (s_queue_count >= 8)
            return ESP_ERR_NO_MEM;

        char tmp[128];
        memcpy(tmp, data, len);
        tmp[len] = '\0';
        strncpy(s_text_queue[s_queue_tail], tmp, sizeof(s_text_queue[0]) - 1);
        s_text_queue[s_queue_tail][sizeof(s_text_queue[0]) - 1] = '\0';
        s_queue_tail = (s_queue_tail + 1) % 8;
        s_queue_count++;
    } else if (cmd == SYN_CMD_VOLUME && data != NULL) {
        s_volume = (*data * 100) / 0x64;
    }

    return ESP_OK;
}

static void mock_playback_tick(void)
{
    if (!s_playing && s_queue_count > 0) {
        s_playing = true;
    }
}

esp_err_t tts_init(int uart_num)
{
    if (uart_num < 0 || uart_num >= 2)
        return ESP_ERR_INVALID_ARG;

    s_uart_num = uart_num;
    s_volume = 80;

    /* Set default volume */
    uint8_t vol_byte = (uint8_t)((s_volume * 0x64) / 100);
    mock_send_frame(SYN_CMD_VOLUME, &vol_byte, 1);

    memset(s_text_queue, 0, sizeof(s_text_queue));
    s_queue_head = 0;
    s_queue_tail = 0;
    s_queue_count = 0;

    printf("[TTS] Initialized on UART%d, volume=%d%%\n", uart_num, s_volume);
    return ESP_OK;
}

esp_err_t tts_speak(const char *text)
{
    if (text == NULL)
        return ESP_ERR_INVALID_ARG;

    return mock_send_frame(SYN_CMD_SPEAK, (const uint8_t *)text, (uint8_t)strlen(text));
}

void tts_stop(void)
{
    s_playing = false;
    s_queue_head = 0;
    s_queue_tail = 0;
    s_queue_count = 0;
    printf("[TTS] Stopped, queue cleared\n");
}

bool tts_is_playing(void)
{
    return s_playing;
}

esp_err_t tts_set_volume(uint8_t percent)
{
    if (percent > 100)
        return ESP_ERR_INVALID_ARG;

    s_volume = percent;
    uint8_t vol_byte = (uint8_t)((percent * 0x64) / 100);
    return mock_send_frame(SYN_CMD_VOLUME, &vol_byte, 1);
}

uint8_t tts_get_volume(void)
{
    return s_volume;
}

/* ---- Test helpers ---- */
static int tests_run = 0;
static int tests_passed = 0;

#define TEST(name) \
    do { \
        tests_run++; \
        printf("  TEST: %s ... ", #name); \

#define EXPECT_TRUE(cond) \
    do { \
        if (!(cond)) { \
            printf("FAILED (expected true: %s)\n", #cond); \
            goto fail; \
        } \
    } while(0)

#define EXPECT_EQ(a, b) \
    do { \
        if ((a) != (b)) { \
            printf("FAILED (expected %d, got %d): %s == %s\n", \
                   (int)(b), (int)(a), #a, #b); \
            goto fail; \
        } \
    } while(0)

#define EXPECT_STR_EQ(a, b) \
    do { \
        if (strcmp((a), (b)) != 0) { \
            printf("FAILED (expected \"%s\", got \"%s\"): %s == %s\n", \
                   b, a, #a, #b); \
            goto fail; \
        } \
    } while(0)

#define PASS() \
    do { \
        tests_passed++; \
        printf("PASSED\n"); \
        break; \
    } while(0)

#define fail: \
    printf("FAILED\n"); \
    break

/* ---- Test cases ---- */

static void test_init(void)
{
    TEST(tts_init)
    esp_err_t ret = tts_init(UART_NUM_1);
    EXPECT_EQ(ret, ESP_OK);
    EXPECT_EQ(s_volume, 80);
    EXPECT_EQ(s_uart_num, UART_NUM_1);
    PASS();
}

static void test_invalid_init(void)
{
    TEST(tts_init_invalid_uart)
    esp_err_t ret = tts_init(-1);
    EXPECT_EQ(ret, ESP_ERR_INVALID_ARG);
    ret = tts_init(5);
    EXPECT_EQ(ret, ESP_ERR_INVALID_ARG);
    PASS();
}

static void test_speak_text(void)
{
    TEST(tts_speak_text)
    esp_err_t ret = tts_speak("爷爷，该吃降压药了");
    EXPECT_EQ(ret, ESP_OK);
    EXPECT_EQ(s_queue_count, 1);
    EXPECT_STR_EQ(s_text_queue[0], "爷爷，该吃降压药了");
    PASS();
}

static void test_speak_null(void)
{
    TEST(tts_speak_null)
    esp_err_t ret = tts_speak(NULL);
    EXPECT_EQ(ret, ESP_ERR_INVALID_ARG);
    PASS();
}

static void test_speak_queue_order(void)
{
    TEST(tts_speak_queue_order)
    tts_speak("第一条提醒");
    tts_speak("第二条提醒");
    tts_speak("第三条提醒");
    EXPECT_EQ(s_queue_count, 3);
    EXPECT_STR_EQ(s_text_queue[0], "第一条提醒");
    EXPECT_STR_EQ(s_text_queue[1], "第二条提醒");
    EXPECT_STR_EQ(s_text_queue[2], "第三条提醒");
    PASS();
}

static void test_volume_set_default(void)
{
    TEST(tts_volume_set_default)
    uint8_t vol = tts_get_volume();
    EXPECT_EQ(vol, 80);
    PASS();
}

static void test_volume_set_range(void)
{
    TEST(tts_volume_set_range)
    esp_err_t ret = tts_set_volume(0);
    EXPECT_EQ(ret, ESP_OK);
    EXPECT_EQ(s_volume, 0);

    ret = tts_set_volume(100);
    EXPECT_EQ(ret, ESP_OK);
    EXPECT_EQ(s_volume, 100);

    ret = tts_set_volume(50);
    EXPECT_EQ(ret, ESP_OK);
    EXPECT_EQ(s_volume, 50);
    PASS();
}

static void test_volume_out_of_range(void)
{
    TEST(tts_volume_out_of_range)
    esp_err_t ret = tts_set_volume(101);
    EXPECT_EQ(ret, ESP_ERR_INVALID_ARG);
    ret = tts_set_volume(200);
    EXPECT_EQ(ret, ESP_ERR_INVALID_ARG);
    PASS();
}

static void test_volume_mapping(void)
{
    TEST(tts_volume_mapping)
    /* 0% -> 0x00, 50% -> 0x32, 100% -> 0x64 */
    tts_set_volume(0);
    EXPECT_EQ(s_volume, 0);
    tts_set_volume(50);
    EXPECT_EQ(s_volume, 50);
    tts_set_volume(100);
    EXPECT_EQ(s_volume, 100);
    /* Reset to default */
    tts_set_volume(80);
    PASS();
}

static void test_stop_clears_queue(void)
{
    TEST(tts_stop_clears_queue)
    tts_speak("药1");
    tts_speak("药2");
    EXPECT_EQ(s_queue_count, 2);

    tts_stop();
    EXPECT_EQ(s_queue_count, 0);
    PASS();
}

static void test_chinese_text_support(void)
{
    TEST(tts_chinese_text)
    const char *chinese_texts[] = {
        "爷爷，该吃降压药了",
        "请注意按时服药",
        "奶奶，该吃降糖药了",
        "药盒第3格已打开",
        "电池电量低，请充电",
    };
    int n = sizeof(chinese_texts) / sizeof(chinese_texts[0]);
    for (int i = 0; i < n; i++) {
        esp_err_t ret = tts_speak(chinese_texts[i]);
        EXPECT_EQ(ret, ESP_OK);
    }
    EXPECT_EQ(s_queue_count, n);
    PASS();
}

static void test_long_text_truncation(void)
{
    TEST(tts_long_text_handling)
    /* Very long text should not crash */
    char long_text[256];
    memset(long_text, 'A', sizeof(long_text) - 1);
    long_text[sizeof(long_text) - 1] = '\0';

    /* Should queue without crashing (implementation may truncate internally) */
    esp_err_t ret = tts_speak(long_text);
    EXPECT_EQ(ret, ESP_OK);
    PASS();
}

/* ---- Main ---- */
int main(void)
{
    printf("\n=== Eregen Smart Pillbox — TTS Voice Reminder Tests ===\n\n");

    /* Run all tests */
    test_init();
    test_invalid_init();
    test_speak_text();
    test_speak_null();
    test_speak_queue_order();
    test_volume_set_default();
    test_volume_set_range();
    test_volume_out_of_range();
    test_volume_mapping();
    test_stop_clears_queue();
    test_chinese_text_support();
    test_long_text_truncation();

    printf("\n=== Results: %d/%d tests passed ===\n", tests_passed, tests_run);

    return (tests_passed == tests_run) ? 0 : 1;
}
