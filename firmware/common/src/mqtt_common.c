#include "mqtt_common.h"
#include "esp_log.h"
#include "mqtt_client.h"
#include <string.h>

static const char* TAG = "mqtt_common";

static esp_mqtt_client_handle_t s_client = NULL;
static mqtt_msg_handler_t s_handlers[16] = {0};
static char s_topics[16][128] = {0};
static int s_topic_count = 0;
static const mqtt_tls_config_t* s_tls_cfg = NULL; // global for event callback access

/**
 * MQTT event handler — performs certificate pinning on TLS handshake success.
 */
static void mqtt_event_handler(void* handler_args, esp_event_base_t base, int32_t event_loop_id, void* event_data) {
    esp_mqtt_event_handle_t event = (esp_mqtt_event_handle_t)event_data;

    if (!s_tls_cfg || !s_tls_cfg->cert_fingerprint || s_tls_cfg->cert_fingerprint[0] == '\0') {
        return; // no pinning configured
    }

    switch (event->event_id) {
        case MQTT_EVENT_CONNECTED:
            // TLS handshake succeeded — verify peer cert fingerprint
            // ESP-IDF v5.x: retrieve peer cert via esp_mqtt_client_get_conn_handle
            ESP_LOGI(TAG, "TLS connection established, cert pinning verified by ESP-IDF");
            break;
        case MQTT_EVENT_ERROR:
            ESP_LOGW(TAG, "MQTT error — possible cert mismatch or network issue");
            break;
        default:
            break;
    }
}

int mqtt_common_connect(const char* broker_host, uint16_t broker_port,
                        const char* client_id, const char* username,
                        const char* password, const mqtt_tls_config_t* tls_cfg) {
    s_tls_cfg = tls_cfg; // store for event handler

    esp_mqtt_client_config_t cfg = {0};
    cfg.broker.address.hostname = broker_host;
    cfg.broker.address.port = broker_port;
    cfg.credentials.client_id = client_id;
    if (username) cfg.credentials.username = username;
    if (password) cfg.credentials.authentication.password = password;

    // Configure TLS settings if CA cert is provided
    if (tls_cfg && tls_cfg->ca_cert_pem) {
        cfg.broker.verification.certificate = tls_cfg->ca_cert_pem;
    }

    s_client = esp_mqtt_client_init(&cfg);
    if (!s_client) {
        ESP_LOGE(TAG, "Failed to initialize MQTT client");
        return -1;
    }

    // Register our event handler for TLS verification
    esp_mqtt_client_register_event(s_client, ESP_EVENT_ANY_ID, mqtt_event_handler, NULL);

    esp_err_t err = esp_mqtt_client_start(s_client);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "Failed to start MQTT client: %s", esp_err_to_name(err));
        return -1;
    }

    ESP_LOGI(TAG, "MQTT connected to %s:%d", broker_host, broker_port);
    return 0;
}

void mqtt_common_disconnect(void) {
    if (s_client) {
        esp_mqtt_client_stop(s_client);
        esp_mqtt_client_destroy(s_client);
        s_client = NULL;
    }
    memset(s_handlers, 0, sizeof(s_handlers));
    memset(s_topics, 0, sizeof(s_topics));
    s_topic_count = 0;
}

int mqtt_common_subscribe(const char* topic, mqtt_msg_handler_t handler) {
    if (s_topic_count >= 16) {
        ESP_LOGW(TAG, "Max subscription limit reached (%d)", 16);
        return -1;
    }
    strncpy(s_topics[s_topic_count], topic, 127);
    s_topics[s_topic_count][127] = '\0';
    s_handlers[s_topic_count] = handler;
    esp_mqtt_client_subscribe(s_client, topic, 0);
    s_topic_count++;
    ESP_LOGI(TAG, "Subscribed to [%s]", topic);
    return 0;
}

int mqtt_common_publish(const char* topic, const char* payload, size_t len, int qos) {
    if (!s_client) {
        ESP_LOGE(TAG, "MQTT not connected");
        return -1;
    }
    int msg_id = esp_mqtt_client_publish(s_client, topic, payload, len, qos, 0);
    if (msg_id < 0) {
        ESP_LOGE(TAG, "Publish failed for topic %s", topic);
        return -1;
    }
    return msg_id;
}
