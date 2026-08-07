cmake_minimum_required(VERSION 3.16)

include($ENV{IDF_PATH}/tools/cmake/project.cmake)

# Component sources
file(GLOB_RECURSE SOURCES
    "main/*.c"
    "../common/src/*.c"
)

idf_component_register(
    SRCS ${SOURCES}
    INCLUDE_DIRS "main"
                 "main/patient_store"
                 "main/nfc_server"
                 "main/display_oled"
                 "main/led_indicator"
                 "main/nvs_manager"
                 "main/security"
                 "main/protocol"
    REQUIRES
        esp_wifi
        esp_event
        esp_netif
        nvs_flash
        driver
        esp_timer
        mqtt
        bt
        esp_https_ota
        esp_http_client
        sha256
        oled_ssd1306
)

# Require ESP-IDF >= 5.3
idf_build_get_property(idf_target IDF_TARGET)
set_source_files_properties(
    ${SOURCES}
    PROPERTIES COMPILE_FLAGS "-DESP_IDF_VERSION_MAJOR=${IDF_VER_MAJOR}"
)
