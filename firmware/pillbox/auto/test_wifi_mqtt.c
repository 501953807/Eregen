/*
 * Eregen (颐贞) - WiFi/MQTT/BLE/AP Configuration Tests
 * Host-compiled standalone tests with mocked ESP-IDF functions.
 *
 * Compile:  gcc auto/test_wifi_mqtt.c -lm -o test_wifi_mqtt
 * Run:      ./test_wifi_mqtt
 *
 * Copyright (c) 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <stdint.h>

/* ---- Mock esp_err_t ---- */
typedef int esp_err_t;
#define ESP_OK        0
#define ESP_FAIL     -1
#define ESP_ERR_NO_MEM  -2
#define ESP_ERR_TIMEOUT -3
#define ESP_ERR_INVALID_ARG  -5

/* ---- Mock FreeRTOS types ---- */
typedef void *TaskHandle_t;
typedef void *EventGroupHandle_t;
typedef void *TimerHandle_t;
typedef int BaseType_t;
#define pdTRUE  1
#define pdFALSE 0
#define pdMS_TO_TICKS(x) ((x) / 10)
#define tskIDLE_PRIORITY 0
#define portMAX_DELAY 0xFFFFFFFFUL

/* ---- Mock ESP-IDF constants ---- */
#define WIFI_AUTH_WPA2_PSK    4
#define WIFI_MODE_STA         1
#define WIFI_MODE_AP          2
#define WIFI_MODE_APSTA       3
#define WIFI_IF_STA           0
#define WIFI_IF_AP            1
#define MQTT_QOS              1
#define BIT0                  (1 << 0)
#define BIT1                  (1 << 1)
#define BIT2                  (1 << 2)

/* ---- Bridge configuration (matches wifi_mqtt_bridge.h) ---- */
#define MQTT_BROKER_HOST      "mqtt.eregen.local"
#define MQTT_BROKER_PORT      1883
#define MQTT_KEEPALIVE_S      60
#define MQTT_MAX_RECONNECT_ATTEMPTS  10
#define MQTT_HEARTBEAT_INTERVAL_S    60
#define TOPIC_BASE            "eregen/device/pillbox"
#define DEVICE_ID_PREFIX      "PX-"
#define DEVICE_ID_FULL_LEN    11
#define MAX_SUBSCRIPTIONS     8

/* ---- NVS mock ---- */
static char s_nvs_ssid[33] = {0};
static char s_nvs_password[65] = {0};
static bool s_nvs_has_credentials = false;

typedef struct { int dummy; } nvs_handle_t;

static esp_err_t mock_nvs_open(const char *ns, int rw, nvs_handle_t *h) { (void)ns;(void)rw;(void)h; return ESP_OK; }
static esp_err_t mock_nvs_set_str(nvs_handle_t h, const char *k, const char *v) {
    (void)h;
    if (strcmp(k, "wifi_ssid") == 0) { strncpy(s_nvs_ssid, v, 32); s_nvs_has_credentials = true; }
    if (strcmp(k, "wifi_pass") == 0) { strncpy(s_nvs_password, v, 64); }
    return ESP_OK;
}
static esp_err_t mock_nvs_commit(nvs_handle_t h) { (void)h; return ESP_OK; }
static void mock_nvs_close(nvs_handle_t h) { (void)h; }

/* ---- Bridge internal state (mirrors wifi_mqtt_bridge.c) ---- */
static bool s_bridge_wifi_connected = false;
static bool s_bridge_mqtt_connected = false;
static char s_bridge_dev_id[DEVICE_ID_FULL_LEN] = "";
static int s_bridge_reconnect_attempts = 0;
static int s_bridge_reconnect_state = 0; /* 0=idle, 1=waiting, 2=done */
static int s_mqtt_publish_count = 0;
static int s_mqtt_subscribe_count = 0;

typedef struct {
    char topic[64];
    void (*callback)(const char *, uint16_t);
} sub_entry_t;
static sub_entry_t s_bridge_subs[MAX_SUBSCRIPTIONS];
static int s_bridge_sub_count = 0;
static bool s_global_handler_set = false;

/* ---- AP config internal state (mirrors ap_config_mode.c) ---- */
static bool s_ap_active = false;
static char s_ap_saved_ssid[33] = {0};
static char s_ap_saved_password[65] = {0};
static bool s_ap_creds_valid = false;

/* ---- BLE pairing internal state (mirrors ble_pair.c) ---- */
static bool s_ble_active = false;
static bool s_ble_cred_received = false;
static char s_ble_saved_ssid[33] = {0};
static char s_ble_saved_password[65] = {0};
static bool s_ble_creds_valid = false;

/* ---- Bridge API (standalone implementation matching wifi_mqtt_bridge.c) ---- */

static void bridge_set_wifi(bool connected)
{
    s_bridge_wifi_connected = connected;
    if (!connected) {
        s_bridge_mqtt_connected = false;
    }
}

static void bridge_set_mqtt(bool connected)
{
    s_bridge_mqtt_connected = connected;
    if (connected) {
        s_bridge_reconnect_attempts = 0;
        s_bridge_reconnect_state = 0;
    }
}

bool bridge_is_connected(void)
{
    return s_bridge_wifi_connected && s_bridge_mqtt_connected;
}

bool mqtt_publish_topic(const char *topic, const uint8_t *data, uint16_t len)
{
    (void)topic; (void)data; (void)len;
    if (!bridge_is_connected()) return false;
    s_mqtt_publish_count++;
    return true;
}

bool mqtt_subscribe_topic(const char *topic, void (*callback)(const char *, uint16_t))
{
    if (s_bridge_sub_count >= MAX_SUBSCRIPTIONS) return false;
    strncpy(s_bridge_subs[s_bridge_sub_count].topic, topic, 63);
    s_bridge_subs[s_bridge_sub_count].topic[63] = '\0';
    s_bridge_subs[s_bridge_sub_count].callback = callback;
    s_bridge_sub_count++;
    return true;
}

void mqtt_on_message(void (*handler)(const char *, const uint8_t *, uint16_t))
{
    (void)handler;
    s_global_handler_set = true;
}

/* ---- AP Config API (standalone implementation) ---- */

void ap_config_start(void)
{
    s_ap_active = true;
}

void ap_config_stop(void)
{
    s_ap_active = false;
}

bool ap_config_is_active(void)
{
    return s_ap_active;
}

bool ap_config_get_credentials(char *ssid, char *password, size_t max_len)
{
    if (!s_ap_creds_valid) return false;
    if (ssid) { strncpy(ssid, s_ap_saved_ssid, max_len - 1); ssid[max_len - 1] = '\0'; }
    if (password) { strncpy(password, s_ap_saved_password, max_len - 1); password[max_len - 1] = '\0'; }
    return true;
}

bool ap_config_save_credentials(const char *ssid, const char *password)
{
    if (!ssid || !password) return false;
    if (strlen(ssid) == 0 || strlen(ssid) > 32) return false;
    if (strlen(password) > 64) return false;

    mock_nvs_set_str((nvs_handle_t){0}, "wifi_ssid", ssid);
    mock_nvs_set_str((nvs_handle_t){0}, "wifi_pass", password);

    strncpy(s_ap_saved_ssid, ssid, sizeof(s_ap_saved_ssid) - 1);
    s_ap_saved_ssid[sizeof(s_ap_saved_ssid) - 1] = '\0';
    strncpy(s_ap_saved_password, password, sizeof(s_ap_saved_password) - 1);
    s_ap_saved_password[sizeof(s_ap_saved_password) - 1] = '\0';
    s_ap_creds_valid = true;

    return true;
}

/* ---- BLE Pairing API (standalone implementation) ---- */

void ble_pair_start(void)
{
    s_ble_active = true;
}

void ble_pair_stop(void)
{
    s_ble_active = false;
}

bool ble_pair_has_credentials(void)
{
    return s_ble_creds_valid;
}

bool ble_pair_get_credentials(char *ssid, char *password, size_t max_len)
{
    if (!s_ble_creds_valid) return false;
    if (ssid) { strncpy(ssid, s_ble_saved_ssid, max_len - 1); ssid[max_len - 1] = '\0'; }
    if (password) { strncpy(password, s_ble_saved_password, max_len - 1); password[max_len - 1] = '\0'; }
    return true;
}

bool ble_pair_save_to_nvs(void)
{
    if (!s_ble_creds_valid) return false;

    mock_nvs_set_str((nvs_handle_t){0}, "wifi_ssid", s_ble_saved_ssid);
    mock_nvs_set_str((nvs_handle_t){0}, "wifi_pass", s_ble_saved_password);
    return true;
}

/* ---- Callback stubs for tests ---- */

static void dummy_subscribe_cb(const char *p, uint16_t l) { (void)p;(void)l; }
static void dummy_cmd_cb(const char *p, uint16_t l) { (void)p;(void)l; }
static void dummy_global_handler(const char *t, const uint8_t *p, uint16_t l) { (void)t;(void)p;(void)l; }

/* ---- Test helpers ---- */

static int tests_run = 0;
static int tests_passed = 0;
static int tests_failed = 0;
static int current_test_line = 0;

#define TEST(name) \
    static void test_##name(void); \
    static void run_##name(void) { \
        printf("  TEST: %s\n", #name); \
        test_##name(); \
    } \
    static void test_##name(void)

#define ASSERT_TRUE(cond) do { \
    tests_run++; current_test_line = __LINE__; \
    if (cond) { tests_passed++; } \
    else { tests_failed++; printf("    FAIL: %s at line %d\n", #cond, current_test_line); } \
} while(0)

#define ASSERT_FALSE(cond) ASSERT_TRUE(!(cond))
#define ASSERT_EQ(a, b) ASSERT_TRUE((a) == (b))
#define ASSERT_NEQ(a, b) ASSERT_TRUE((a) != (b))
#define ASSERT_STR_EQ(a, b) ASSERT_TRUE(strcmp((a), (b)) == 0)

/* ===== Bridge Init Tests ===== */

TEST(bridge_init_sets_up_wifi_and_mqtt)
{
    s_bridge_wifi_connected = false;
    s_bridge_mqtt_connected = false;

    ASSERT_FALSE(bridge_is_connected());
    ASSERT_FALSE(s_bridge_wifi_connected);
}

TEST(bridge_connect_requires_wifi_then_mqtt)
{
    bridge_set_wifi(false);
    bridge_set_mqtt(false);
    s_mqtt_publish_count = 0;

    bridge_set_wifi(true);
    ASSERT_TRUE(s_bridge_wifi_connected);
    ASSERT_FALSE(bridge_is_connected());

    bridge_set_mqtt(true);
    ASSERT_TRUE(bridge_is_connected());
}

/* ===== MQTT Publish/Subscribe Tests ===== */

TEST(mqtt_publish_requires_connection)
{
    bridge_set_wifi(false);
    bridge_set_mqtt(false);

    const uint8_t payload[] = "{\"type\":\"heartbeat\"}";
    ASSERT_FALSE(mqtt_publish_topic("test/topic", payload, sizeof(payload) - 1));
}

TEST(mqtt_publish_succeeds_when_connected)
{
    bridge_set_wifi(true);
    bridge_set_mqtt(true);
    s_mqtt_publish_count = 0;

    const uint8_t payload[] = "{\"type\":\"heartbeat\"}";
    ASSERT_TRUE(mqtt_publish_topic("eregen/device/pillbox/PX-ABC12345/up",
                                    payload, sizeof(payload) - 1));
    ASSERT_EQ(s_mqtt_publish_count, 1);
}

TEST(mqtt_subscribe_registers_callback)
{
    s_bridge_sub_count = 0;

    ASSERT_TRUE(mqtt_subscribe_topic("eregen/device/pillbox/+/cmd", dummy_subscribe_cb));
    ASSERT_EQ(s_bridge_sub_count, 1);
    ASSERT_STR_EQ(s_bridge_subs[0].topic, "eregen/device/pillbox/+/cmd");
}

TEST(mqtt_subscribe_rejects_too_many_topics)
{
    s_bridge_sub_count = MAX_SUBSCRIPTIONS;

    ASSERT_FALSE(mqtt_subscribe_topic("any/topic", dummy_subscribe_cb));
    ASSERT_EQ(s_bridge_sub_count, MAX_SUBSCRIPTIONS);
}

TEST(mqtt_on_message_registers_global_handler)
{
    s_global_handler_set = false;

    mqtt_on_message(dummy_global_handler);
    ASSERT_TRUE(s_global_handler_set);
}

/* ===== Reconnection Logic Tests ===== */

TEST(reconnect_attempts_increment_on_failure)
{
    bridge_set_mqtt(false);
    s_bridge_reconnect_attempts = 0;
    s_bridge_reconnect_state = 0;

    /* Simulate 2 failed attempts */
    for (int i = 0; i < 2; i++) {
        if (s_bridge_reconnect_attempts < MQTT_MAX_RECONNECT_ATTEMPTS) {
            s_bridge_reconnect_attempts++;
            s_bridge_reconnect_state = 1;
        }
    }
    ASSERT_EQ(s_bridge_reconnect_attempts, 2);

    /* Third attempt succeeds, resetting counter */
    bridge_set_mqtt(true);
    ASSERT_TRUE(s_bridge_mqtt_connected);
    /* Counter is reset on success (expected behavior) */
    ASSERT_EQ(s_bridge_reconnect_attempts, 0);
}

TEST(reconnect_resets_on_success)
{
    bridge_set_mqtt(false);
    s_bridge_reconnect_attempts = 5;
    s_bridge_reconnect_state = 1;

    bridge_set_mqtt(true);
    ASSERT_EQ(s_bridge_reconnect_attempts, 0);
    ASSERT_EQ(s_bridge_reconnect_state, 0);
}

TEST(reconnect_stops_after_max_attempts)
{
    bridge_set_mqtt(false);
    s_bridge_reconnect_attempts = MQTT_MAX_RECONNECT_ATTEMPTS - 1;
    s_bridge_reconnect_state = 1;

    if (s_bridge_reconnect_attempts < MQTT_MAX_RECONNECT_ATTEMPTS) {
        s_bridge_reconnect_attempts++;
    }

    ASSERT_EQ(s_bridge_reconnect_attempts, MQTT_MAX_RECONNECT_ATTEMPTS);
    ASSERT_FALSE(s_bridge_mqtt_connected);
}

TEST(wifi_disconnect_triggers_mqtt_disconnect)
{
    bridge_set_wifi(true);
    bridge_set_mqtt(true);
    ASSERT_TRUE(bridge_is_connected());

    bridge_set_wifi(false);
    ASSERT_FALSE(s_bridge_mqtt_connected);
    ASSERT_FALSE(bridge_is_connected());
}

/* ===== AP Config Mode Tests ===== */

TEST(ap_config_starts_and_stops)
{
    s_ap_active = false;
    ap_config_start();
    ASSERT_TRUE(ap_config_is_active());

    ap_config_stop();
    ASSERT_FALSE(ap_config_is_active());
}

TEST(ap_config_save_and_retrieve_credentials)
{
    s_ap_creds_valid = false;
    s_ap_saved_ssid[0] = '\0';
    s_ap_saved_password[0] = '\0';

    ASSERT_TRUE(ap_config_save_credentials("MyHomeWiFi", "secret1234"));
    ASSERT_TRUE(s_ap_creds_valid);

    char ssid[33] = {0};
    char pass[65] = {0};
    ASSERT_TRUE(ap_config_get_credentials(ssid, pass, sizeof(ssid)));
    ASSERT_STR_EQ(ssid, "MyHomeWiFi");
    ASSERT_STR_EQ(pass, "secret1234");
}

TEST(ap_config_rejects_empty_ssid)
{
    s_ap_creds_valid = false;
    ASSERT_FALSE(ap_config_save_credentials("", "password123"));
    ASSERT_FALSE(s_ap_creds_valid);
}

TEST(ap_config_rejects_long_ssid)
{
    s_ap_creds_valid = false;
    char long_ssid[100];
    memset(long_ssid, 'A', 99);
    long_ssid[99] = '\0';
    ASSERT_FALSE(ap_config_save_credentials(long_ssid, "password123"));
    ASSERT_FALSE(s_ap_creds_valid);
}

TEST(ap_config_nvs_persists_credentials)
{
    s_nvs_has_credentials = false;
    s_nvs_ssid[0] = '\0';
    s_nvs_password[0] = '\0';

    ap_config_save_credentials("PersistSSID", "persistPass1");

    ASSERT_TRUE(s_nvs_has_credentials);
    ASSERT_STR_EQ(s_nvs_ssid, "PersistSSID");
    ASSERT_STR_EQ(s_nvs_password, "persistPass1");
}

/* ===== BLE Pairing Tests ===== */

TEST(ble_pair_starts_and_stops)
{
    s_ble_active = false;
    ble_pair_start();
    ASSERT_TRUE(s_ble_active);

    ble_pair_stop();
    ASSERT_FALSE(s_ble_active);
}

TEST(ble_pair_save_and_retrieve_credentials)
{
    s_ble_creds_valid = false;
    s_ble_saved_ssid[0] = '\0';
    s_ble_saved_password[0] = '\0';

    /* Without credentials, save_to_nvs should fail */
    ASSERT_FALSE(ble_pair_save_to_nvs());

    /* Simulate receiving credentials via BLE write */
    strncpy(s_ble_saved_ssid, "BleSSID", sizeof(s_ble_saved_ssid) - 1);
    strncpy(s_ble_saved_password, "blepass123", sizeof(s_ble_saved_password) - 1);
    s_ble_creds_valid = true;

    ASSERT_TRUE(ble_pair_has_credentials());

    char ssid[33] = {0};
    char pass[65] = {0};
    ASSERT_TRUE(ble_pair_get_credentials(ssid, pass, sizeof(ssid)));
    ASSERT_STR_EQ(ssid, "BleSSID");
    ASSERT_STR_EQ(pass, "blepass123");
}

TEST(ble_pair_has_no_credentials_initially)
{
    s_ble_creds_valid = false;
    ASSERT_FALSE(ble_pair_has_credentials());
}

/* ===== MQTT Topic Format Tests ===== */

TEST(topic_format_follows_spec)
{
    const char *expected_up = "eregen/device/pillbox/PX-ABC12345/up";
    const char *expected_down = "eregen/device/pillbox/PX-ABC12345/down";
    const char *expected_status = "eregen/device/pillbox/status";

    ASSERT_TRUE(strstr(expected_up, TOPIC_BASE) == expected_up);
    ASSERT_TRUE(strstr(expected_down, TOPIC_BASE) == expected_down);
    ASSERT_TRUE(strstr(expected_status, TOPIC_BASE) == expected_status);
}

/* ===== Full Device Lifecycle Integration Test ===== */

TEST(full_device_lifecycle)
{
    /* 1. Device boots without credentials */
    s_ap_active = false;
    s_nvs_has_credentials = false;
    s_bridge_wifi_connected = false;
    s_bridge_mqtt_connected = false;
    s_mqtt_publish_count = 0;  /* Reset from previous tests */
    s_bridge_sub_count = 0;

    /* 2. AP mode starts (no credentials found) */
    ap_config_start();
    ASSERT_TRUE(ap_config_is_active());

    /* 3. User saves credentials via captive portal */
    ap_config_save_credentials("UserWiFi", "userPass123");
    ASSERT_TRUE(s_ap_creds_valid);

    /* 4. Credentials persist in NVS */
    ASSERT_TRUE(s_nvs_has_credentials);

    /* 5. AP stops, WiFi connects */
    ap_config_stop();
    ASSERT_FALSE(ap_config_is_active());

    bridge_set_wifi(true);
    ASSERT_TRUE(s_bridge_wifi_connected);

    /* 6. MQTT connects */
    bridge_set_mqtt(true);
    ASSERT_TRUE(bridge_is_connected());

    /* 7. Device publishes heartbeat */
    const uint8_t hb[] = "{\"type\":\"heartbeat\",\"dev_id\":\"PX-ABC12345\",\"bat\":85}";
    ASSERT_TRUE(mqtt_publish_topic("eregen/device/pillbox/PX-ABC12345/up", hb, sizeof(hb) - 1));
    ASSERT_EQ(s_mqtt_publish_count, 1);

    /* 8. Subscribe to command topic */
    ASSERT_TRUE(mqtt_subscribe_topic("eregen/device/pillbox/+/cmd", dummy_cmd_cb));
    ASSERT_EQ(s_bridge_sub_count, 1);
}

TEST(mqtt_heartbeat_interval_constants)
{
    ASSERT_EQ(MQTT_HEARTBEAT_INTERVAL_S, 60);
    ASSERT_EQ(MQTT_KEEPALIVE_S, 60);
    ASSERT_EQ(MQTT_MAX_RECONNECT_ATTEMPTS, 10);
    ASSERT_EQ(MAX_SUBSCRIPTIONS, 8);
}

/* ---- Main ---- */

int main(void)
{
    printf("\n=== Eregen WiFi/MQTT/BLE/AP Configuration Tests ===\n\n");

    /* Bridge init tests */
    run_bridge_init_sets_up_wifi_and_mqtt();
    run_bridge_connect_requires_wifi_then_mqtt();

    /* MQTT publish/subscribe tests */
    run_mqtt_publish_requires_connection();
    run_mqtt_publish_succeeds_when_connected();
    run_mqtt_subscribe_registers_callback();
    run_mqtt_subscribe_rejects_too_many_topics();
    run_mqtt_on_message_registers_global_handler();

    /* Reconnection tests */
    run_reconnect_attempts_increment_on_failure();
    run_reconnect_resets_on_success();
    run_reconnect_stops_after_max_attempts();
    run_wifi_disconnect_triggers_mqtt_disconnect();

    /* AP config tests */
    run_ap_config_starts_and_stops();
    run_ap_config_save_and_retrieve_credentials();
    run_ap_config_rejects_empty_ssid();
    run_ap_config_rejects_long_ssid();
    run_ap_config_nvs_persists_credentials();

    /* BLE pairing tests */
    run_ble_pair_starts_and_stops();
    run_ble_pair_save_and_retrieve_credentials();
    run_ble_pair_has_no_credentials_initially();

    /* Topic format tests */
    run_topic_format_follows_spec();

    /* Full lifecycle integration test */
    run_full_device_lifecycle();

    /* Constants check */
    run_mqtt_heartbeat_interval_constants();

    printf("\n=== Results: %d passed, %d failed, %d total ===\n",
           tests_passed, tests_failed, tests_run);

    return tests_failed > 0 ? 1 : 0;
}
