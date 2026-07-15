/*
 * Eregen (颐贞) - Board Initialization Header
 * GD32E230C8T3 board init: GPIO, clocks, UART, I2C, SPI, timers
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef BOARD_INIT_H
#define BOARD_INIT_H

#include <stdint.h>
#include <stdbool.h>

/* Clock configuration - 120MHz HCLK max for GD32E230 */
#define BOARD_SYSTEM_CLOCK_HZ     120000000UL

/* UART debug configuration */
#define BOARD_DEBUG_UART_BAUD      115200

/* GPIO pin definitions */
/* LEDs: PA0 (green), PA1 (blue) */
#define LED_GREEN_GPIO_PORT        GPIOA
#define LED_GREEN_GPIO_PIN         GPIO_PIN_0
#define LED_BLUE_GPIO_PORT         GPIOA
#define LED_BLUE_GPIO_PIN          GPIO_PIN_1

/* Buttons: PA2 (user button), PA3 (SOS) */
#define USER_BUTTON_GPIO_PORT      GPIOA
#define USER_BUTTON_GPIO_PIN       GPIO_PIN_2
#define SOS_BUTTON_GPIO_PORT       GPIOA
#define SOS_BUTTON_GPIO_PIN        GPIO_PIN_3

/* I2C: PB8 (SCL), PB9 (SDA) */
#define BOARD_I2C_PORT             I2C1
#define BOARD_I2C_SCL_PIN          GPIO_PIN_8
#define BOARD_I2C_SDA_PIN          GPIO_PIN_9

/* SPI: PB3 (SPI1_SCK), PB4 (SPI1_MISO), PB5 (SPI1_MOSI) */
#define BOARD_SPI_PORT             SPI1
#define BOARD_SPI_SCK_PIN          GPIO_PIN_3
#define BOARD_SPI_MISO_PIN         GPIO_PIN_4
#define BOARD_SPI_MOSI_PIN         GPIO_PIN_5

/* SPI chip selects */
#define IMU_CS_GPIO_PORT           GPIOB
#define IMU_CS_GPIO_PIN            GPIO_PIN_12
#define GPS_CS_GPIO_PORT           GPIOB
#define GPS_CS_GPIO_PIN            GPIO_PIN_13
#define DISPLAY_CS_GPIO_PORT       GPIOB
#define DISPLAY_CS_GPIO_PIN        GPIO_PIN_14

/* ADC: PA5 for battery voltage measurement */
#define BATTERY_ADC_GPIO_PORT      GPIOA
#define BATTERY_ADC_GPIO_PIN       GPIO_PIN_5
#define BATTERY_ADC_CHANNEL        GPIO_ADC_IN_5

/* Timer for system tick and peripheral timing */
#define BOARD_SYS_TIMER            TIMER0

/* Function declarations */
void board_clock_init(void);
void board_gpio_init(void);
void board_uart_debug_init(uint32_t baud);
void board_i2c_init(void);
void board_spi_init(void);
void board_timer_init(void);
void board_init_all(void);

#endif /* BOARD_INIT_H */
