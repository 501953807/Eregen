/*
 * Eregen (颐贞) - AP Configuration Mode Implementation
 * Creates a soft-AP + captive portal for WiFi credential provisioning.
 *
 * Copyright (c) 2026 Eregen (颐贞). All rights reserved.
 */

#include "ap_config_mode.h"

#include <string.h>
#include <stdio.h>
#include <stdlib.h>

#include "esp_log.h"
#include "esp_wifi.h"
#include "esp_event.h"
#include "esp_netif.h"
#include "nvs_flash.h"

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include "lwip/err.h"
#include "lwip/sockets.h"
#include "lwip/sys.h"
#include <lwip/netdb.h>

static const char *TAG = "ap_config";

/* AP configuration */
#define AP_SSID_PREFIX      "eregen-pixel-"
#define AP_SSID_MAX_LEN     32
#define AP_PASSWORD         "eregen1234"       /* WPA2, 8+ chars */
#define AP_CHANNEL          1
#define AP_MAX_CONNECTIONS  4
#define AP_BEacon_INTERVAL  100                  /* 100ms */

/* Captive portal HTTP server */
#define AP_HTTP_PORT        80
#define AP_RECV_BUF_SIZE    1024

/* NVS keys */
#define NVS_KEY_WIFI_SSID   "wifi_ssid"
#define NVS_KEY_WIFI_PASS   "wifi_pass"

/* Internal state */
static bool s_ap_active = false;
static char s_saved_ssid[33] = {0};
static char s_saved_password[65] = {0};
static bool s_credentials_valid = false;

/* ---- Captive portal HTML ---- */

static const char *CAPTIVE_HTML =
    "<!DOCTYPE html>"
    "<html lang='zh-CN'>"
    "<head><meta charset='utf-8'>"
    "<meta name='viewport' content='width=device-width,initial-scale=1'>"
    "<title>Eregen 配网</title>"
    "<style>"
    "*{margin:0;padding:0;box-sizing:border-box}"
    "body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;"
    "background:#f5f5f0;color:#2c2c2c;display:flex;align-items:center;"
    "justify-content:center;min-height:100vh}"
    ".card{background:#fff;border-radius:16px;padding:40px 32px;"
    "max-width:360px;width:90%;box-shadow:0 2px 20px rgba(0,0,0,.08)}"
    "h1{text-align:center;font-size:22px;margin-bottom:6px;color:#1a5c2a}"
    ".sub{text-align:center;font-size:13px;color:#888;margin-bottom:28px}"
    "label{display:block;font-size:13px;font-weight:600;margin-bottom:4px;color:#555}"
    "input{width:100%;padding:10px 12px;border:1px solid #ddd;border-radius:8px;"
    "font-size:15px;margin-bottom:16px;outline:none}"
    "input:focus{border-color:#1a5c2a}"
    "button{width:100%;padding:12px;background:#1a5c2a;color:#fff;border:none;"
    "border-radius:8px;font-size:16px;font-weight:600;cursor:pointer}"
    "button:active{background:#145022}"
    ".footer{text-align:center;font-size:11px;color:#aaa;margin-top:20px}"
    "</style></head>"
    "<body>"
    "<div class='card'>"
    "<h1>Eregen 颐贞</h1>"
    "<p class='sub'>为智能药盒连接 Wi-Fi 网络</p>"
    "<form action='/connect' method='POST'>"
    "<label>Wi-Fi 名称 (SSID)</label>"
    "<input name='ssid' required placeholder='输入你的 Wi-Fi 名称'>"
    "<label>Wi-Fi 密码</label>"
    "<input name='password' type='password' required placeholder='输入密码'>"
    "<button type='submit'>保存并连接</button>"
    "</form>"
    "<p class='footer'>设备会自动重启并完成配网</p>"
    "</div></body></html>";

/* ---- Forward declarations ---- */

static void ap_config_task(void *pvParameter);
static void http_server_task(void *pvParameter);
static void extract_ap_suffix(char *suffix, size_t len);

/**
 * Start AP configuration mode.
 * Runs in its own FreeRTOS task so it doesn't block app_main().
 */
void ap_config_start(void)
{
    if (s_ap_active) return;

    xTaskCreate(ap_config_task, "ap_cfg", 4096, NULL, 5, NULL);
}

/**
 * Stop AP configuration mode.
 */
void ap_config_stop(void)
{
    s_ap_active = false;

    esp_err_t ret = esp_wifi_set_mode(WIFI_MODE_STA);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Failed to set STA mode: %s", esp_err_to_name(ret));
    }

    esp_wifi_disconnect();
    ESP_LOGI(TAG, "AP configuration mode stopped");
}

bool ap_config_is_active(void)
{
    return s_ap_active;
}

bool ap_config_get_credentials(char *ssid, char *password, size_t max_len)
{
    if (!s_credentials_valid) return false;

    if (ssid) {
        strncpy(ssid, s_saved_ssid, max_len - 1);
        ssid[max_len - 1] = '\0';
    }
    if (password) {
        strncpy(password, s_saved_password, max_len - 1);
        password[max_len - 1] = '\0';
    }
    return true;
}

bool ap_config_save_credentials(const char *ssid, const char *password)
{
    if (!ssid || !password) return false;
    if (strlen(ssid) == 0 || strlen(ssid) > 32) return false;
    if (strlen(password) > 64) return false;

    nvs_handle_t handle;
    esp_err_t ret = nvs_open("pillbox", NVS_READWRITE, &handle);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS open failed: %s", esp_err_to_name(ret));
        return false;
    }

    ret = nvs_set_str(handle, NVS_KEY_WIFI_SSID, ssid);
    if (ret == ESP_OK) {
        ret = nvs_set_str(handle, NVS_KEY_WIFI_PASS, password);
    }
    if (ret == ESP_OK) {
        ret = nvs_commit(handle);
    }
    nvs_close(handle);

    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS save failed: %s", esp_err_to_name(ret));
        return false;
    }

    /* Store in local buffer too */
    strncpy(s_saved_ssid, ssid, sizeof(s_saved_ssid) - 1);
    s_saved_ssid[sizeof(s_saved_ssid) - 1] = '\0';
    strncpy(s_saved_password, password, sizeof(s_saved_password) - 1);
    s_saved_password[sizeof(s_saved_password) - 1] = '\0';
    s_credentials_valid = true;

    ESP_LOGI(TAG, "Credentials saved to NVS: %s", ssid);
    return true;
}

/* ---- AP setup task ---- */

static void ap_config_task(void *pvParameter)
{
    (void)pvParameter;

    /* Set AP mode */
    esp_err_t ret = esp_wifi_set_mode(WIFI_MODE_APSTA);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Failed to set APSTA mode: %s", esp_err_to_name(ret));
        vTaskDelete(NULL);
        return;
    }

    /* Build AP SSID from MAC suffix */
    char ap_ssid[AP_SSID_MAX_LEN];
    char mac_suffix[5] = {0};
    extract_ap_suffix(mac_suffix, sizeof(mac_suffix));
    snprintf(ap_ssid, sizeof(ap_ssid), "%s%s", AP_SSID_PREFIX, mac_suffix);

    wifi_config_t ap_conf = {
        .ap = {
            .ssid_len = strlen(ap_ssid),
            .channel = AP_CHANNEL,
            .authmode = WIFI_AUTH_WPA2_PSK,
            .ssid_hidden = 0,
            .max_connection = AP_MAX_CONNECTIONS,
            .beacon_interval = AP_BEacon_INTERVAL,
        },
    };
    memcpy(ap_conf.ap.ssid, ap_ssid, strlen(ap_ssid));
    memcpy(ap_conf.ap.password, AP_PASSWORD, strlen(AP_PASSWORD));

    ret = esp_wifi_set_config(WIFI_IF_AP, &ap_conf);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Failed to set AP config: %s", esp_err_to_name(ret));
        vTaskDelete(NULL);
        return;
    }

    ret = esp_wifi_start();
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Failed to start WiFi: %s", esp_err_to_name(ret));
        vTaskDelete(NULL);
        return;
    }

    s_ap_active = true;

    /* Get AP IP address */
    esp_netif_ip_info_t ip_info;
    esp_netif_t *ap_netif = esp_netif_get_handle_from_ifkey("WIFI_AP_DEF");
    if (ap_netif && esp_netif_get_ip_info(ap_netif, &ip_info) == ESP_OK) {
        ESP_LOGI(TAG, "AP started: %s  IP: %s", ap_ssid, inet_ntoa(ip_info.gw));
    } else {
        ESP_LOGI(TAG, "AP started: %s", ap_ssid);
    }

    /* Start captive portal HTTP server */
    xTaskCreate(http_server_task, "http_srv", 4096, NULL, 5, NULL);

    /* Wait indefinitely — user will configure via browser */
    for (;;) {
        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/* ---- Captive portal HTTP server ---- */

static void http_server_task(void *pvParameter)
{
    (void)pvParameter;

    int server_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (server_fd < 0) {
        ESP_LOGE(TAG, "Socket creation failed");
        vTaskDelete(NULL);
        return;
    }

    /* Allow address reuse */
    int opt = 1;
    setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));

    struct sockaddr_in addr;
    memset(&addr, 0, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = INADDR_ANY;
    addr.sin_port = htons(AP_HTTP_PORT);

    if (bind(server_fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        ESP_LOGE(TAG, "Bind failed on port %d", AP_HTTP_PORT);
        close(server_fd);
        vTaskDelete(NULL);
        return;
    }

    if (listen(server_fd, 1) < 0) {
        ESP_LOGE(TAG, "Listen failed");
        close(server_fd);
        vTaskDelete(NULL);
        return;
    }

    ESP_LOGI(TAG, "Captive portal HTTP server listening on port %d", AP_HTTP_PORT);

    char recv_buf[AP_RECV_BUF_SIZE];

    for (;;) {
        struct sockaddr_in client_addr;
        socklen_t addr_len = sizeof(client_addr);
        int client_fd = accept(server_fd, (struct sockaddr *)&client_addr, &addr_len);

        if (client_fd < 0) {
            ESP_LOGE(TAG, "Accept failed");
            continue;
        }

        /* Read request */
        int bytes_read = recv(client_fd, recv_buf, sizeof(recv_buf) - 1, 0);
        if (bytes_read <= 0) {
            close(client_fd);
            continue;
        }
        recv_buf[bytes_read] = '\0';

        ESP_LOGD(TAG, "HTTP request (%d bytes): %.200s...", recv_buf);

        /* Check if this is a POST /connect (credential submission) */
        if (strstr(recv_buf, "POST /connect") != NULL) {
            /* Extract SSID and password from form body */
            char *body = strstr(recv_buf, "\r\n\r\n");
            if (!body) body = strstr(recv_buf, "\n\n");
            if (body) {
                body += 4;  /* skip past CRLF+CRLF */

                char post_ssid[33] = {0};
                char post_pass[65] = {0};

                /* Parse ssid=...&password=... */
                char *ssid_pos = strstr(body, "ssid=");
                char *pass_pos = strstr(body, "password=");

                if (ssid_pos) {
                    ssid_pos += 5;
                    char *amp = strchr(ssid_pos, '&');
                    size_t slen = amp ? (size_t)(amp - ssid_pos) : strlen(ssid_pos);
                    if (slen > 32) slen = 32;
                    strncpy(post_ssid, ssid_pos, slen);
                    post_ssid[slen] = '\0';
                }
                if (pass_pos) {
                    pass_pos += 9;
                    size_t plen = strlen(pass_pos);
                    if (plen > 64) plen = 64;
                    strncpy(post_pass, pass_pos, plen);
                    post_pass[plen] = '\0';
                }

                if (strlen(post_ssid) > 0) {
                    if (ap_config_save_credentials(post_ssid, post_pass)) {
                        /* Success: redirect to done page */
                        const char *redirect_html =
                            "<html><head><meta http-equiv='refresh' content='3;url=/done'></head>"
                            "<body style='font-family:sans-serif;text-align:center;padding:60px 20px;"
                            "<h1>配网成功</h1><p>正在连接到您的 Wi-Fi...</p>"
                            "<p>设备将在 3 秒后重启</p></body></html>";

                        const char *resp_ok =
                            "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n"
                            "Content-Length: %d\r\nConnection: close\r\n\r\n%s";
                        char hdr[128];
                        snprintf(hdr, sizeof(hdr), resp_ok,
                                 (int)strlen(redirect_html), redirect_html);
                        send(client_fd, hdr, strlen(hdr), 0);
                        send(client_fd, redirect_html, strlen(redirect_html), 0);
                    } else {
                        /* Save failed — retry */
                        const char *resp_fail =
                            "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n"
                            "Content-Length: %d\r\nConnection: close\r\n\r\n"
                            "<html><body style='font-family:sans-serif;text-align:center;padding:60px 20px'>"
                            "<h1>保存失败</h1><p>请重试</p></body></html>";
                        char hdr[128];
                        snprintf(hdr, sizeof(hdr), resp_fail, 117);
                        send(client_fd, hdr, strlen(hdr), 0);
                        send(client_fd,
                            "<html><body style='font-family:sans-serif;text-align:center;padding:60px 20px'>"
                            "<h1>保存失败</h1><p>请重试</p></body></html>",
                            117, 0);
                    }
                }
            }
        } else if (strstr(recv_buf, "GET /connect") ||
                   strstr(recv_buf, "GET /") ||
                   strstr(recv_buf, "GET /favicon")) {
            /* Serve captive portal page */
            const char *resp =
                "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n"
                "Content-Length: %d\r\nConnection: close\r\n\r\n%s";
            char hdr[128];
            snprintf(hdr, sizeof(hdr), resp,
                     (int)strlen(CAPTIVE_HTML), CAPTIVE_HTML);
            send(client_fd, hdr, strlen(hdr), 0);
            send(client_fd, CAPTIVE_HTML, strlen(CAPTIVE_HTML), 0);
        }

        close(client_fd);
    }

    close(server_fd);
    vTaskDelete(NULL);
}

/* ---- Helpers ---- */

static void extract_ap_suffix(char *suffix, size_t len)
{
    uint8_t mac[6];
    esp_read_mac(mac, ESP_MAC_WIFI_STA);
    snprintf(suffix, len, "%02X%02X", mac[4], mac[5]);
}
