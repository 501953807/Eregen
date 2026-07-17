idf_component_register(
    SRCS "src/mqtt_common.c" "src/brand_boot_logo.c"
    INCLUDE_DIRS "include"
    REQUIRES mqtt nvs_flash driver esp_timer
)
