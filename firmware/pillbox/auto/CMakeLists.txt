idf_component_register(
    SRCS "state_machine.c" "dispensing.c" "schedule_engine.c"
         "med_rule_parser.c" "opto_sensor.c" "empty_detector.c"
         "nvs_store.c" "ble_pair.c" "ap_config_mode.c"
    INCLUDE_DIRS "."
                 "../common/include"
                 "../../common/include"
    REQUIRES driver nvs_flash esp_wifi mqtt esp_netif esp_event esp_timer esp_adc_cal esp_https_ota esp_http_client cJSON sha256
)
