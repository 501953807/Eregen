/*
 * Eregen (颐贞) - Pro Board Hardware Abstraction Layer
 * GD32E230C8T3 pin definitions and hardware init for the Pro tier bracelet.
 *
 * Differences from Entry/Plus:
 *   - Dedicated 3.3V LDO for analog sensor section (PA4)
 *   - I2C1 (PB8/PB9) routed to ECG chip (ADAS1000) instead of PPG
 *   - SPI1 (PB3-PB5) dedicated to AMOLED display (ST7701S)
 *   - GPS on USART3 (PB10/PB11) with u-blox NEO-M9N GNSS module
 *   - Separate CS pins for ECG, Display, and GPS
 *   - Additional LED (amber) for ECG measurement status (PC13)
 *
 * This board file is a drop-in replacement for entry/board_init.c/h
 * but targets the Pro hardware layout for metal housing.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef BOARD_PRO_H
#define BOARD_PRO_H

#include <stdint.h>
#include <stdbool.h>

/* ----------------------------------------------------------------
 * System clock configuration
 * ---------------------------------------------------------------- */
#define BOARD_PRO_SYSTEM_CLOCK_HZ    120000000UL

/* Debug UART baud rate */
#define BOARD_PRO_DEBUG_UART_BAUD    115200

/* ----------------------------------------------------------------
 * Power management pins
 * ---------------------------------------------------------------- */
/* PA4: 3.3V Analog LDO enable (active-high, metal housing req.) */
#define BOARD_PRO_ANA_LDO_PORT       GPIOA
#define BOARD_PRO_ANA_LDO_PIN        GPIO_PIN_4

/* PC14: Main VCC rail enable for digital section */
#define BOARD_PRO_VCC_RAIL_PORT      GPIOC
#define BOARD_PRO_VCC_RAIL_PIN       GPIO_PIN_14

/* ----------------------------------------------------------------
 * LED indicators
 * ---------------------------------------------------------------- */
/* PA0: Green - system alive (same as entry) */
#define BOARD_PRO_LED_GREEN_PORT     GPIOA
#define BOARD_PRO_LED_GREEN_PIN      GPIO_PIN_0

/* PA1: Blue - connectivity status (same as entry) */
#define BOARD_PRO_LED_BLUE_PORT      GPIOA
#define BOARD_PRO_LED_BLUE_PIN       GPIO_PIN_1

/* PC13: Amber - ECG measurement active */
#define BOARD_PRO_LED_AMBER_PORT     GPIOC
#define BOARD_PRO_LED_AMBER_PIN      GPIO_PIN_13

/* ----------------------------------------------------------------
 * Button inputs
 * ---------------------------------------------------------------- */
/* PA2: User button */
#define BOARD_PRO_USER_BTN_PORT      GPIOA
#define BOARD_PRO_USER_BTN_PIN       GPIO_PIN_2

/* PA3: SOS button (same as entry) */
#define BOARD_PRO_SOS_BTN_PORT       GPIOA
#define BOARD_PRO_SOS_BTN_PIN        GPIO_PIN_3

/* ----------------------------------------------------------------
 * I2C bus - ECG chip (ADAS1000EVM compatible)
 * PB8 = SCL, PB9 = SDA
 * ---------------------------------------------------------------- */
#define BOARD_PRO_I2C                I2C1
#define BOARD_PRO_I2C_SCL_PIN        GPIO_PIN_8
#define BOARD_PRO_I2C_SDA_PIN        GPIO_PIN_9
#define BOARD_PRO_I2C_GPIO_PORT      GPIOB

/* ECG chip I2C address (ADAS1000 default) */
#define BOARD_PRO_ECG_I2C_ADDR       0x68U

/* ----------------------------------------------------------------
 * SPI1 bus - AMOLED display (ST7701S)
 * PB3 = SCK, PB4 = MISO, PB5 = MOSI
 * ---------------------------------------------------------------- */
#define BOARD_PRO_SPI                SPI1
#define BOARD_PRO_SPI_SCK_PIN        GPIO_PIN_3
#define BOARD_PRO_SPI_MISO_PIN       GPIO_PIN_4
#define BOARD_PRO_SPI_MOSI_PIN       GPIO_PIN_5
#define BOARD_PRO_SPI_GPIO_PORT      GPIOB

/* Display CS and control pins */
#define BOARD_PRO_DISPLAY_CS_PORT    GPIOB
#define BOARD_PRO_DISPLAY_CS_PIN     GPIO_PIN_12
#define BOARD_PRO_DISPLAY_DC_PORT    GPIOB
#define BOARD_PRO_DISPLAY_DC_PIN     GPIO_PIN_14
#define BOARD_PRO_DISPLAY_RST_PORT   GPIOB
#define BOARD_PRO_DISPLAY_RST_PIN    GPIO_PIN_15

/* ----------------------------------------------------------------
 * UART3 - GNSS module (u-blox NEO-M9N)
 * PB10 = TX, PB11 = RX
 * ---------------------------------------------------------------- */
#define BOARD_PRO_GNSS_UART          USART3
#define BOARD_PRO_GNSS_TX_PIN        GPIO_PIN_10
#define BOARD_PRO_GNSS_RX_PIN        GPIO_PIN_11
#define BOARD_PRO_GNSS_GPIO_PORT     GPIOB

/* GNSS module enable pin */
#define BOARD_PRO_GNSS_EN_PORT       GPIOC
#define BOARD_PRO_GNSS_EN_PIN        GPIO_PIN_0

/* GNSS baud rate (NEO-M9N default) */
#define BOARD_PRO_GNSS_BAUD          9600

/* ----------------------------------------------------------------
 * Cat1 cellular module (shared with entry)
 * USART1 on PB6/PB7
 * ---------------------------------------------------------------- */
#define BOARD_PRO_CAT1_UART          USART1
#define BOARD_PRO_CAT1_TX_PIN        GPIO_PIN_6
#define BOARD_PRO_CAT1_RX_PIN        GPIO_PIN_7
#define BOARD_PRO_CAT1_GPIO_PORT     GPIOB
#define BOARD_PRO_CAT1_EN_PIN        GPIO_PIN_1  /* PC1 */
#define BOARD_PRO_CAT1_RST_PIN       GPIO_PIN_2  /* PC2 */

/* ----------------------------------------------------------------
 * Battery ADC (shared)
 * PA5
 * ---------------------------------------------------------------- */
#define BOARD_PRO_BATT_ADC_PORT      GPIOA
#define BOARD_PRO_BATT_ADC_PIN       GPIO_PIN_5
#define BOARD_PRO_BATT_ADC_CHANNEL   GPIO_ADC_IN_5

/* ----------------------------------------------------------------
 * Function declarations
 * ---------------------------------------------------------------- */
void board_pro_clock_init(void);
void board_pro_gpio_init(void);
void board_pro_analog_power_on(void);
void board_pro_digital_power_on(void);
void board_pro_uart_debug_init(uint32_t baud);
void board_pro_uart_gnss_init(uint32_t baud);
void board_pro_uart_cat1_init(uint32_t baud);
void board_pro_i2c_init(void);
void board_pro_spi_init(void);
void board_pro_timer_init(void);
void board_pro_exti_init(void);
void board_pro_init_all(void);

#endif /* BOARD_PRO_H */
