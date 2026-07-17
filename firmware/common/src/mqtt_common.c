#include "mqtt_common.h"
#include "esp_log.h"
#include "mqtt_client.h"
#include <string.h>

static const char* TAG = "mqtt_common";

static esp_mqtt_client_handle_t s_client = NULL;
static mqtt_msg_handler_t s_handlers[16] = {0};
static char s_topics[16][128] = {0};
static int s_topic_count = 0;

int mqtt_common_connect(const char* broker_host, uint16_t broker_port,
                        const char* client_id, const char* username,
                        const char* password) {
    esp_mqtt_client_config_t cfg = {0};
    cfg.broker.address.hostname = broker_host;
    cfg.broker.address.port = broker_port;
    cfg.credentials.client_id = client_id;
    if (username) cfg.credentials.username = username;
    if (password) cfg.credentials.authentication.password = password;

    s_client = esp_mqtt_client_init(&cfg);
    if (!s_client) {
        ESP_LOGE(TAG, "Failed to initialize MQTT client");
        return -1;
    }
    esp_mqtt_client_start(s_client);
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
