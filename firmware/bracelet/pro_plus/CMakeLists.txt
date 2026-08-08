/*
 * Eregen (颐贞) - Pro+ Bracelet Firmware Build Configuration
 * 在 Pro 基础上增加：电化学试纸检测 + BLE 血压计连接 + 慢病管理调度
 *
 * 目标 MCU: GD32E230C8T3 (ARM Cortex-M4, 128KB Flash, 32KB SRAM)
 * 宏定义:   TARGET_PRO_PLUS
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

cmake_minimum_required(VERSION 3.22)

project(eregen_bracelet_pro_plus C)

# ----------------------------------------------------------------
# 继承 Pro 的工具链配置
# ----------------------------------------------------------------
set(CMAKE_SYSTEM_NAME Generic)
set(CMAKE_SYSTEM_PROCESSOR ARM)
set(CMAKE_TRY_COMPILE_TARGET_TYPE STATIC_LIBRARY)

# 编译器标志（与 pro 一致，Cortex-M4 + FPU）
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} \
    -mcpu=cortex-m4 \
    -mthumb \
    -mfpu=fpv4-sp-d16 \
    -mfloat-abi=hard \
    -fsingle-precision-float \
    -DTARGET_PRO_PLUS \
")

if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

set(CMAKE_C_FLAGS_RELEASE "-O2 -DNDEBUG")
set(CMAKE_C_FLAGS_DEBUG   "-Og -g3 -DDEBUG")

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} \
    -Wall \
    -Wextra \
    -Werror \
    -Wpedantic \
    -Wconversion \
    -std=c99 \
")

# 链接器标志
set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} \
    -mcpu=cortex-m4 \
    -mthumb \
    -mfpu=fpv4-sp-d16 \
    -mfloat-abi=hard \
    -T${CMAKE_SOURCE_DIR}/../pro/linker/eregen_cm4.ld \
    -nostartfiles \
    -lc \
    -lm \
    -lnosys \
")

# 包含路径
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${CMAKE_CURRENT_SOURCE_DIR}/../pro          # Pro 头文件（board_pro.h, ecg_driver.h 等）
    ${CMAKE_CURRENT_SOURCE_DIR}/../entry        # Entry 公共模块
    ${CMAKE_CURRENT_SOURCE_DIR}/../common       # 公共工具（log, crc16, ring_buffer）
    ${CMAKE_SOURCE_DIR}/FreeRTOS-Kernel/include
    ${CMAKE_SOURCE_DIR}/FreeRTOS-Kernel/portable/GCC/ARM_CM4F
    ${CMAKE_SOURCE_DIR}/gd32e230_drivers/include
    ${CMAKE_SOURCE_DIR}/gd32e230_drivers/source
)

# ----------------------------------------------------------------
# 来源文件：以 Pro 为基础，添加 Pro+ 专属模块
# ----------------------------------------------------------------
set(PRO_SOURCES
    # ---- 继承自 Pro ----
    ../pro/board_pro.c
    ../pro/ecg_driver.c
    ../pro/display_amoled.c
    ../pro/gps_gnss.c
    ../pro/power_optimizer.c
    ../pro/free_rtos_tasks.c
    ../pro/ota/boot_switch.c
    ../pro/ota/ota_download.c
    ../pro/ota/ota_verify.c
    ../pro/ota/ota_update.c

    # ---- 继承自 Entry ----
    ../entry/sensors_ppg.c
    ../entry/sensors_imu.c
    ../entry/battery_adc.c
    ../entry/sos_button.c
    ../entry/cat1_at.c
    ../entry/protocol/message_encode.c
    ../entry/protocol/heartbeat.c
    ../entry/common/crc16.c
    ../entry/common/ring_buffer.c
    ../entry/common/log.c

    # ---- Pro+ 专属模块 ----
    # 协议定义
    protocol/strip_type.h
    protocol/electrochemical_msg.h
    # 传感器驱动
    sensors/electrochemical.c
    # BLE 外设
    bt_peripheral/bp_device.c
    # 应用层
    app/chronic_manager.c
)

# 自身 main.c（Pro+ 专属入口）
set(PRO_PLUS_MAIN main.c)

# ----------------------------------------------------------------
# 构建目标
# ----------------------------------------------------------------
add_executable(${PROJECT_NAME}.elf
    ${PRO_SOURCES}
    ${PRO_PLUS_MAIN}
)

# 生成 .bin 和 .hex
add_custom_command(TARGET ${PROJECT_NAME}.elf POST_BUILD
    COMMAND ${CMAKE_OBJCOPY} -O binary ${PROJECT_NAME}.elf ${PROJECT_NAME}.bin
    COMMAND ${CMAKE_OBJCOPY} -O ihex  ${PROJECT_NAME}.elf ${PROJECT_NAME}.hex
    COMMENT "Generating Pro+ firmware binaries (.bin / .hex)"
)

add_custom_command(TARGET ${PROJECT_NAME}.elf POST_BUILD
    COMMAND ${CMAKE_SIZE} ${PROJECT_NAME}.elf
    COMMENT "Pro+ firmware size report"
)

# ----------------------------------------------------------------
# 烧录目标
# ----------------------------------------------------------------
add_custom_target(flash
    COMMAND st-flash write ${PROJECT_NAME}.bin 0x08000000
    DEPENDS ${PROJECT_NAME}.elf
    COMMENT "Flashing Pro+ firmware via st-flash"
)

add_custom_target(openocd-flash
    COMMAND openocd -f interface/stlink-v2-1.cfg \
                    -f target/gd32e230.cfg \
                    -c "program ${PROJECT_NAME}.elf verify reset exit"
    DEPENDS ${PROJECT_NAME}.elf
    COMMENT "Flashing Pro+ firmware via OpenOCD"
)
