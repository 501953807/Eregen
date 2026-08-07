/*
 * Eregen Medical Wristband - Patient Store Header
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef PATIENT_STORE_H
#define PATIENT_STORE_H

#include "esp_err.h"
#include <stdint.h>

#define MAX_PATIENT_ID_LEN  64
#define MAX_NAME_LEN        32
#define MAX_ADMISSION_LEN   32

typedef struct {
    char patient_id[MAX_PATIENT_ID_LEN];
    char admission_no[MAX_ADMISSION_LEN];
    char name[MAX_NAME_LEN];
    char department[MAX_NAME_LEN];
    char bed_number[MAX_NAME_LEN];
    uint8_t bind_timestamp;
} patient_info_t;

esp_err_t patient_store_init(void);
esp_err_t patient_store_save(const patient_info_t *info);
esp_err_t patient_store_load(patient_info_t *info);
esp_err_t patient_store_clear(void);
void patient_store_deinit(void);

#endif /* PATIENT_STORE_H */
