# Eregen Community Wristband - ESP32-S3

## Board Configuration
- BOARD = ESP32-S3-DevKitC-1-N8R8 (or compatible)
- Partition scheme: 4MB Flash
- PSRAM: Disabled (not needed for community mode)

## Component Paths
- MAIN_SRC_DIRS = main
- REQUIRES = wifi_mqtt_bridge oled esp_ble_gap esp_event nvs_flash
