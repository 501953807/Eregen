# Eregen (颐贞) - Pro Bracelet Firmware Build Configuration
# GD32E230C8T3 (ARM Cortex-M4), FreeRTOS, C99
#
# © 2026 Eregen (颐贞). All rights reserved.

cmake_minimum_required(VERSION 3.22)

project(eregen_bracelet_pro C)

# ----------------------------------------------------------------
# Toolchain configuration (GD32 / ARM GCC)
# ----------------------------------------------------------------
set(CMAKE_SYSTEM_NAME Generic)
set(CMAKE_SYSTEM_PROCESSOR ARM)

set(CMAKE_TRY_COMPILE_TARGET_TYPE STATIC_LIBRARY)

# Compiler flags for Cortex-M4 with FPU
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} \
    -mcpu=cortex-m4 \
    -mthumb \
    -mfpu=fpv4-sp-d16 \
    -mfloat-abi=hard \
    -fsingle-precision-float \
")

# Optimization: -O2 for production, -Og for debug builds
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

set(CMAKE_C_FLAGS_RELEASE "-O2 -DNDEBUG")
set(CMAKE_C_FLAGS_DEBUG   "-Og -g3 -DDEBUG")

# Common warnings
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} \
    -Wall \
    -Wextra \
    -Werror \
    -Wpedantic \
    -Wconversion \
    -std=c99 \
")

# Linker flags
set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} \
    -mcpu=cortex-m4 \
    -mthumb \
    -mfpu=fpv4-sp-d16 \
    -mfloat-abi=hard \
    -T${CMAKE_SOURCE_DIR}/linker/eregen_cm4.ld \
    -nostartfiles \
    -lc \
    -lm \
    -lnosys \
")

# Include paths
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${CMAKE_CURRENT_SOURCE_DIR}/../entry
    ${CMAKE_CURRENT_SOURCE_DIR}/../common
    ${CMAKE_SOURCE_DIR}/FreeRTOS-Kernel/include
    ${CMAKE_SOURCE_DIR}/FreeRTOS-Kernel/portable/GCC/ARM_CM4F
    ${CMAKE_SOURCE_DIR}/gd32e230_drivers/include
    ${CMAKE_SOURCE_DIR}/gd32e230_drivers/source
    ${CMAKE_SOURCE_DIR}/shared/protocol
    ${CMAKE_SOURCE_DIR}/shared/crypto
)

# ----------------------------------------------------------------
# Source files
# ----------------------------------------------------------------
set(PRO_SOURCES
    # Board hardware abstraction
    board_pro.c

    # Pro-specific drivers
    ecg_driver.c
    display_amoled.c
    gps_gnss.c
    power_optimizer.c

    # FreeRTOS task management
    free_rtos_tasks.c

    # Main entry point
    main.c

    # Shared common libraries
    ../common/crc16.c
    ../common/ring_buffer.c
    ../common/log.c

    # Entry-level shared drivers (PPG, IMU, battery, SOS, Cat1)
    ../entry/sensors_ppg.c
    ../entry/sensors_imu.c
    ../entry/battery_adc.c
    ../entry/sos_button.c
    ../entry/cat1_at.c

    # Protocol libraries
    ../protocol/message_encode.c
    ../protocol/heartbeat.c
)

# ----------------------------------------------------------------
# Build targets
# ----------------------------------------------------------------
add_executable(${PROJECT_NAME}.elf ${PRO_SOURCES})

# Generate binary and hex for flashing
add_custom_command(TARGET ${PROJECT_NAME}.elf POST_BUILD
    COMMAND ${CMAKE_OBJCOPY} -O binary ${PROJECT_NAME}.elf ${PROJECT_NAME}.bin
    COMMAND ${CMAKE_OBJCOPY} -O ihex ${PROJECT_NAME}.elf ${PROJECT_NAME}.hex
    COMMENT "Generating firmware binaries: .bin and .hex"
)

# Generate size report
add_custom_command(TARGET ${PROJECT_NAME}.elf POST_BUILD
    COMMAND ${CMAKE_SIZE} ${PROJECT_NAME}.elf
    COMMENT "Firmware size report"
)

# ----------------------------------------------------------------
# Flashing (via st-flash or OpenOCD for GD32)
# ----------------------------------------------------------------
add_custom_target(flash
    COMMAND st-flash write ${PROJECT_NAME}.bin 0x08000000
    DEPENDS ${PROJECT_NAME}.elf
    COMMENT "Flashing firmware via st-flash"
)

add_custom_target(openocd-flash
    COMMAND openocd -f interface/stlink-v2-1.cfg -f target/gd32e230.cfg \
        -c "program ${PROJECT_NAME}.elf verify reset exit"
    DEPENDS ${PROJECT_NAME}.elf
    COMMENT "Flashing firmware via OpenOCD + GD32 adapter"
)
