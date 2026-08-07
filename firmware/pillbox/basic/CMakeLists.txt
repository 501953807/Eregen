# ESP-IDF v5.3+ Component Registration
# Pillbox Basic - ESP32-C3
# © 2026 Eregen (颐贞). All rights reserved.

idf_component_register(
    SRCS "main.c" "led_gpio.c" "battery_manage.c" "button_input.c"
    INCLUDE_DIRS "."
    REQUIRES
        driver
        esp_timer
        nvs_flash
)

# Require ESP-IDF >= 5.3
idf_build_get_property(idf_target IDF_TARGET)
set_source_files_properties(
    ${CMAKE_CURRENT_LIST_DIR}/main.c
    ${CMAKE_CURRENT_LIST_DIR}/led_gpio.c
    ${CMAKE_CURRENT_LIST_DIR}/battery_manage.c
    ${CMAKE_CURRENT_LIST_DIR}/button_input.c
    PROPERTIES COMPILE_FLAGS "-DESP_IDF_VERSION_MAJOR=${IDF_VER_MAJOR}"
)
