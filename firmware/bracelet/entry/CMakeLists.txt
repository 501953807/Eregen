# Eregen Bracelet Entry - GD32E230C8T3 Build Configuration
# Target MCU: GD32E230C8T3 (ARM Cortex-M4, 64KB Flash, 20KB SRAM)
# © 2026 Eregen (颐贞). All rights reserved.

cmake_minimum_required(VERSION 3.22)

set(CMAKE_TOOLCHAIN_FILE ${CMAKE_SOURCE_DIR}/../toolchain/gd32.cmake)

project(eregen_bracelet_entry C)

# MCU settings - GD32E230C8T3 is STM32L-series compatible
set(MCU_ARCH cortex-m4)
set(MCU_FLAGS -mcpu=cortex-m4 -mthumb -mfpu=fpv4-sp-d16 -mfloat-abi=hard)
set(COMPILER_FLAGS ${MCU_FLAGS} -ffunction-sections -fdata-sections -fno-common -Wall)

# FreeRTOS source paths
set(FREERTOS_PATH ${CMAKE_SOURCE_DIR}/../../../../third_party/FreeRTOS-Kernel)
set(FREERTOS_SOURCES
    ${FREERTOS_PATH}/portable/GCC/ARM_CM4F/port.c
    ${FREERTOS_PATH}/list.c
    ${FREERTOS_PATH}/queue.c
    ${FREERTOS_PATH}/tasks.c
    ${FREERTOS_PATH}/timers.c
    ${FREERTOS_PATH}/event_groups.c
)

# GD32 standard peripheral library
set(GD32_LIB ${CMAKE_SOURCE_DIR}/../../../../third_party/gd32e230_stdperlib)
set(GD32_SOURCES
    ${GD32_LIB}/source/gd32e230_gpio.c
    ${GD32_LIB}/source/gd32e230_usart.c
    ${GD32_LIB}/source/gd32e230_timer.c
    ${GD32_LIB}/source/gd32e230_i2c.c
    ${GD32_LIB}/source/gd32e230_spi.c
    ${GD32_LIB}/source/gd32e230_rcu.c
    ${GD32_LIB}/source/gd32e230_adc.c
)

# Application sources - all driver files
set(APP_SOURCES
    main.c
    board_init.c
    sensors_ppg.c
    sensors_imu.c
    gps_nmea.c
    cat1_at.c
    display_st7789.c
    battery_adc.c
    sos_button.c
    free_rtos_tasks.c
)

# Linker script
set(LINKER_SCRIPT ${CMAKE_SOURCE_DIR}/../../../../third_party/gd32e230c8t3_flash.ld)

# Build executable
add_executable(${PROJECT_NAME}.elf ${APP_SOURCES} ${FREERTOS_SOURCES} ${GD32_SOURCES})

target_compile_options(${PROJECT_NAME}.elf PRIVATE ${COMPILER_FLAGS})
target_link_options(${PROJECT_NAME}.elf PRIVATE
    -Wl,--gc-sections
    -Wl,-Map=${CMAKE_BINARY_DIR}/${PROJECT_NAME}.map
    -T${LINKER_SCRIPT}
)

target_include_directories(${PROJECT_NAME}.elf PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${FREERTOS_PATH}/include
    ${FREERTOS_PATH}/portable/GCC/ARM_CM4F
    ${GD32_LIB}/include
    ${GD32_LIB}/device
)

# Define device type macro
target_compile_definitions(${PROJECT_NAME}.elf PRIVATE
    -DDEVICE_TYPE_BRACELET_ENTRY
)

# Generate binary and hex
add_custom_command(TARGET ${PROJECT_NAME}.elf POST_BUILD
    COMMAND ${CMAKE_OBJCOPY} -O binary ${PROJECT_NAME}.elf ${PROJECT_NAME}.bin
    COMMAND ${CMAKE_OBJCOPY} -O ihex ${PROJECT_NAME}.elf ${PROJECT_NAME}.hex
    COMMENT "Generating .bin and .hex files for flashing")
