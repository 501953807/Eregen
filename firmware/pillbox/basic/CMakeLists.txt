# ESP-IDF v5.3+ Component Registration
# Pillbox Basic - ESP32-C3
# © 2026 Eregen (颐贞). All rights reserved.

idf_component_register(
    SRCS "main.c"
    INCLUDE_DIRS "."
    REQUIRES
        esp_wifi
        esp_event
        esp_netif
        nvs_flash
        driver
        esp_timer
)
