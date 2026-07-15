# ESP-IDF v5.3+ Component Registration
# Pillbox Smart - ESP32-C3 | Smart tier with TTS, OLED, APP linkage
# © 2026 Eregen (颐贞). All rights reserved.

idf_component_register(
    SRCS "main.c" "voice_reminder.c" "reminder_scheduler.c"
         "app_link.c" "oled_status.c" "volume_control.c"
    INCLUDE_DIRS "."
    REQUIRES
        esp_wifi
        esp_event
        esp_netif
        nvs_flash
        driver
        esp_timer
        cJSON
)

# Require ESP-IDF >= 5.3
idf_build_get_property(idf_target IDF_TARGET)
set_source_files_properties(
    ${CMAKE_CURRENT_LIST_DIR}/main.c
    ${CMAKE_CURRENT_LIST_DIR}/voice_reminder.c
    ${CMAKE_CURRENT_LIST_DIR}/reminder_scheduler.c
    ${CMAKE_CURRENT_LIST_DIR}/app_link.c
    ${CMAKE_CURRENT_LIST_DIR}/oled_status.c
    ${CMAKE_CURRENT_LIST_DIR}/volume_control.c
    PROPERTIES COMPILE_FLAGS "-DESP_IDF_VERSION_MAJOR=${IDF_VER_MAJOR}"
)
