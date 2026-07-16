/*
 * Eregen (颐贞) - Pro Board Hardware Implementation
 * Pin mappings, clock trees, and peripheral init for Pro variant.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "board_pro.h"
#include "../common/log.h"
#include <string.h>

/* ----------------------------------------------------------------
 * Clock initialization
 * ---------------------------------------------------------------- */

void board_pro_clock_init(void)
{
    /* Enable HXTAL (external high-speed crystal) */
    rcu_osci_on(RCU_HXTAL);
    rcu_osci_stab_wait(RCU_HXTAL_STAB);

    /* Select HXTAL as PLL source */
    rcu_cksel_config(
        RCU_CKSEL_HXTAL,
        0U,
        0U,
        0U,
        RCU_CDIV_DIV_1,
        RCU_HXTAL_DIV_1
    );

    /* Configure PLL: HXTAL 8MHz * 15 = 120MHz */
    rcu_pll_config(RCU_PLL_MUL15);

    /* Enable PLL */
    rcu_osci_on(RCU_PLL);
    rcu_osci_stab_wait(RCU_PLL_STAB);

    /* Flash wait states for 120MHz */
    flash_prefetch_enable();
    flash_instr_latency_set(FLASH_ILFT_1_5);
    flash_data_latency_set(FLASH_DLT_1_5);

    /* Select PLL as SYSCLK source */
    rcu_clock_extend_config(RCU_CFG0_CSSD_EN_MASK, true);
    rcu_sysclk_config(RCU_SRCCLK_PLL);

    uint32_t timeout = 0xFFFFFU;
    while ((rcu_flag_get(RCU_FLAG_HXTAL_STABLE) == RESET) && (timeout-- > 0)) {
        /* Wait for HXTAL stable */
    }

    log_info("BoardPro: SYSCLK = %lu Hz", (unsigned long)BOARD_PRO_SYSTEM_CLOCK_HZ);
}

/* ----------------------------------------------------------------
 * GPIO initialization
 * ---------------------------------------------------------------- */

void board_pro_gpio_init(void)
{
    /* Enable GPIO clocks */
    rcu_periph_clock_enable(RCU_GPIOA);
    rcu_periph_clock_enable(RCU_GPIOB);
    rcu_periph_clock_enable(RCU_GPIOC);

    /* LEDs */
    gpio_init(BOARD_PRO_LED_GREEN_PORT, GPIO_MODE_OUT_PP,
              GPIO_OSPEED_2MHZ, BOARD_PRO_LED_GREEN_PIN);
    gpio_bit_set(BOARD_PRO_LED_GREEN_PORT, BOARD_PRO_LED_GREEN_PIN);

    gpio_init(BOARD_PRO_LED_BLUE_PORT, GPIO_MODE_OUT_PP,
              GPIO_OSPEED_2MHZ, BOARD_PRO_LED_BLUE_PIN);
    gpio_bit_set(BOARD_PRO_LED_BLUE_PORT, BOARD_PRO_LED_BLUE_PIN);

    gpio_init(BOARD_PRO_LED_AMBER_PORT, GPIO_MODE_OUT_PP,
              GPIO_OSPEED_2MHZ, BOARD_PRO_LED_AMBER_PIN);
    gpio_bit_set(BOARD_PRO_LED_AMBER_PORT, BOARD_PRO_LED_AMBER_PIN);

    /* Buttons */
    gpio_init(BOARD_PRO_USER_BTN_PORT, GPIO_MODE_IPU,
              GPIO_OSPEED_50MHZ, BOARD_PRO_USER_BTN_PIN);
    gpio_init(BOARD_PRO_SOS_BTN_PORT, GPIO_MODE_IPU,
              GPIO_OSPEED_50MHZ, BOARD_PRO_SOS_BTN_PIN);

    /* Power rails - start disabled, enabled explicitly */
    gpio_init(BOARD_PRO_ANA_LDO_PORT, GPIO_MODE_OUT_PP,
              GPIO_OSPEED_50MHZ, BOARD_PRO_ANA_LDO_PIN);
    gpio_bit_reset(BOARD_PRO_ANA_LDO_PORT, BOARD_PRO_ANA_LDO_PIN);

    gpio_init(BOARD_PRO_VCC_RAIL_PORT, GPIO_MODE_OUT_PP,
              GPIO_OSPEED_50MHZ, BOARD_PRO_VCC_RAIL_PIN);
    gpio_bit_reset(BOARD_PRO_VCC_RAIL_PORT, BOARD_PRO_VCC_RAIL_PIN);

    /* SPI CS / control pins (high = deselected) */
    gpio_init(BOARD_PRO_DISPLAY_CS_PORT, GPIO_MODE_OUT_PP,
              GPIO_OSPEED_50MHZ, BOARD_PRO_DISPLAY_CS_PIN);
    gpio_bit_set(BOARD_PRO_DISPLAY_CS_PORT, BOARD_PRO_DISPLAY_CS_PIN);

    gpio_init(BOARD_PRO_DISPLAY_DC_PORT, GPIO_MODE_OUT_PP,
              GPIO_OSPEED_50MHZ, BOARD_PRO_DISPLAY_DC_PIN);
    gpio_bit_set(BOARD_PRO_DISPLAY_DC_PORT, BOARD_PRO_DISPLAY_DC_PIN);

    gpio_init(BOARD_PRO_DISPLAY_RST_PORT, GPIO_MODE_OUT_PP,
              GPIO_OSPEED_50MHZ, BOARD_PRO_DISPLAY_RST_PIN);
    gpio_bit_set(BOARD_PRO_DISPLAY_RST_PORT, BOARD_PRO_DISPLAY_RST_PIN);

    /* GNSS enable (disabled at boot) */
    gpio_init(BOARD_PRO_GNSS_EN_PORT, GPIO_MODE_OUT_PP,
              GPIO_OSPEED_50MHZ, BOARD_PRO_GNSS_EN_PIN);
    gpio_bit_reset(BOARD_PRO_GNSS_EN_PORT, BOARD_PRO_GNSS_EN_PIN);

    /* Cat1 module pins */
    gpio_init(GPIOC, GPIO_MODE_OUT_PP, GPIO_OSPEED_50MHZ, GPIO_PIN_1);
    gpio_bit_reset(GPIOC, GPIO_PIN_1);
    gpio_init(GPIOC, GPIO_MODE_OUT_PP, GPIO_OSPEED_50MHZ, GPIO_PIN_2);
    gpio_bit_set(GPIOC, GPIO_PIN_2);
}

/* ----------------------------------------------------------------
 * Power rail helpers
 * ---------------------------------------------------------------- */

void board_pro_analog_power_on(void)
{
    rcu_periph_clock_enable(RCU_GPIOA);
    gpio_bit_set(BOARD_PRO_ANA_LDO_PORT, BOARD_PRO_ANA_LDO_PIN);
    /* Wait for LDO output to stabilize (~1ms typical) */
    for (volatile uint32_t d = 0; d < 12000U; d++) {
        /* spin ~1ms at 120MHz */
        (void)d;
    }
    log_info("BoardPro: Analog LDO enabled");
}

void board_pro_digital_power_on(void)
{
    rcu_periph_clock_enable(RCU_GPIOC);
    gpio_bit_set(BOARD_PRO_VCC_RAIL_PORT, BOARD_PRO_VCC_RAIL_PIN);
    log_info("BoardPro: Digital VCC rail enabled");
}

/* ----------------------------------------------------------------
 * UART initializations
 * ---------------------------------------------------------------- */

void board_pro_uart_debug_init(uint32_t baud)
{
    rcu_periph_clock_enable(RCU_GPIOA);
    rcu_periph_clock_enable(RCU_USART0);

    gpio_init(GPIOA, GPIO_MODE_AF_PP, GPIO_OSPEED_50MHZ, GPIO_PIN_9);
    gpio_init(GPIOA, GPIO_MODE_INPUT, GPIO_OSPEED_50MHZ, GPIO_PIN_10);

    usart_deinit(USART0);
    usart_baudrate_set(USART0, baud);
    usart_word_length_set(USART0, USART_WL_8BIT);
    usart_stop_bit_set(USART0, USART_STB_1BIT);
    usart_parity_config(USART0, USART_PM_NONE);
    usart_hardware_flow_rts_config(USART0, USART_RTS_DISABLE);
    usart_hardware_flow_cts_config(USART0, USART_CTS_DISABLE);
    usart_transmit_config(USART0, USART_TRANSMIT_ENABLE);
    usart_receive_config(USART0, USART_RECEIVE_ENABLE);
    usart_enable(USART0);

    while (usart_flag_get(USART0, USART_FLAG_RBNE) == RESET) {
        /* Wait */
    }
}

void board_pro_uart_gnss_init(uint32_t baud)
{
    rcu_periph_clock_enable(RCU_GPIOB);
    rcu_periph_clock_enable(RCU_USART3);

    gpio_init(GPIOB, GPIO_MODE_AF_PP, GPIO_OSPEED_50MHZ, BOARD_PRO_GNSS_TX_PIN);
    gpio_init(GPIOB, GPIO_MODE_INPUT, GPIO_OSPEED_50MHZ, BOARD_PRO_GNSS_RX_PIN);

    usart_deinit(BOARD_PRO_GNSS_UART);
    usart_baudrate_set(BOARD_PRO_GNSS_UART, baud);
    usart_word_length_set(BOARD_PRO_GNSS_UART, USART_WL_8BIT);
    usart_stop_bit_set(BOARD_PRO_GNSS_UART, USART_STB_1BIT);
    usart_parity_config(BOARD_PRO_GNSS_UART, USART_PM_NONE);
    usart_hardware_flow_rts_config(BOARD_PRO_GNSS_UART, USART_RTS_DISABLE);
    usart_hardware_flow_cts_config(BOARD_PRO_GNSS_UART, USART_CTS_DISABLE);
    usart_transmit_config(BOARD_PRO_GNSS_UART, USART_TRANSMIT_ENABLE);
    usart_receive_config(BOARD_PRO_GNSS_UART, USART_RECEIVE_ENABLE);
    usart_enable(BOARD_PRO_GNSS_UART);

    log_info("BoardPro: GNSS UART initialized at %lu baud", (unsigned long)baud);
}

void board_pro_uart_cat1_init(uint32_t baud)
{
    rcu_periph_clock_enable(RCU_GPIOB);
    rcu_periph_clock_enable(RCU_USART1);

    gpio_init(GPIOB, GPIO_MODE_AF_PP, GPIO_OSPEED_50MHZ, BOARD_PRO_CAT1_TX_PIN);
    gpio_init(GPIOB, GPIO_MODE_INPUT, GPIO_OSPEED_50MHZ, BOARD_PRO_CAT1_RX_PIN);

    usart_deinit(BOARD_PRO_CAT1_UART);
    usart_baudrate_set(BOARD_PRO_CAT1_UART, baud);
    usart_word_length_set(BOARD_PRO_CAT1_UART, USART_WL_8BIT);
    usart_stop_bit_set(BOARD_PRO_CAT1_UART, USART_STB_1BIT);
    usart_parity_config(BOARD_PRO_CAT1_UART, USART_PM_NONE);
    usart_hardware_flow_rts_config(BOARD_PRO_CAT1_UART, USART_RTS_DISABLE);
    usart_hardware_flow_cts_config(BOARD_PRO_CAT1_UART, USART_CTS_DISABLE);
    usart_transmit_config(BOARD_PRO_CAT1_UART, USART_TRANSMIT_ENABLE);
    usart_receive_config(BOARD_PRO_CAT1_UART, USART_RECEIVE_ENABLE);
    usart_enable(BOARD_PRO_CAT1_UART);

    log_info("BoardPro: Cat1 UART initialized at %lu baud", (unsigned long)baud);
}

/* ----------------------------------------------------------------
 * I2C initialization (for ECG chip)
 * ---------------------------------------------------------------- */

void board_pro_i2c_init(void)
{
    rcu_periph_clock_enable(RCU_GPIOB);
    rcu_periph_clock_enable(RCU_I2C1);

    /* PB8 = SCL, PB9 = SDA, alternate function open-drain */
    gpio_init(GPIOB, GPIO_MODE_AF_OD, GPIO_OSPEED_50MHZ, BOARD_PRO_I2C_SCL_PIN);
    gpio_init(GPIOB, GPIO_MODE_AF_OD, GPIO_OSPEED_50MHZ, BOARD_PRO_I2C_SDA_PIN);

    i2c_deinit(BOARD_PRO_I2C);
    i2c_clock_config(BOARD_PRO_I2C, 400000, I2C_CKCTL_DHSL_ENABLE);
    i2c_mode_addr_config(BOARD_PRO_I2C, I2C_MODE_I2C, I2C_ADDFORMAT_7BIT);
    i2c_ack_config(BOARD_PRO_I2C, I2C_ACK_ENABLE);
    i2c_enable(BOARD_PRO_I2C);

    log_info("BoardPro: I2C1 initialized at 400kHz (fast mode)");
}

/* ----------------------------------------------------------------
 * SPI initialization (for AMOLED display)
 * ---------------------------------------------------------------- */

void board_pro_spi_init(void)
{
    rcu_periph_clock_enable(RCU_GPIOB);
    rcu_periph_clock_enable(RCU_SPI1);

    /* PB3 = SCK (AF push-pull), PB5 = MOSI (AF push-pull) */
    gpio_init(GPIOB, GPIO_MODE_AF_PP, GPIO_OSPEED_50MHZ, BOARD_PRO_SPI_SCK_PIN);
    gpio_init(GPIOB, GPIO_MODE_AF_PP, GPIO_OSPEED_50MHZ, BOARD_PRO_SPI_MOSI_PIN);
    /* PB4 = MISO (AF input) */
    gpio_init(GPIOB, GPIO_MODE_INPUT, GPIO_OSPEED_50MHZ, BOARD_PRO_SPI_MISO_PIN);

    spi_deinit(SPI1);
    spi_master_config(SPI1, SPI_FRF_NONE, SPI_SMCR_NONE,
                      SPI_NSS_SOFT, SPI_CLKPOL_LOW, SPI_CLKPH_1EDGE,
                      SPI_DATADIST_8BIT);
    spi_sclk_speed_set(SPI1, SPI_SCCLK_DIV8); /* 15 MHz max for 120 MHz SYSCLK */
    spi_enable(SPI1);

    log_info("BoardPro: SPI1 initialized at 15 MHz (master)");
}

/* ----------------------------------------------------------------
 * Timer and EXTI
 * ---------------------------------------------------------------- */

void board_pro_timer_init(void)
{
    rcu_periph_clock_enable(RCU_TIMER0);
    /* TIMER0 used by FreeRTOS tick and general timing */
    log_info("BoardPro: TIMER0 enabled for FreeRTOS");
}

void board_pro_exti_init(void)
{
    rcu_periph_clock_enable(RCU_AF);

    /* EXTI3 for SOS button (PA3) */
    gpio_exti_source_select(GPIO_PORT_SOURCE_GPIOA, GPIO_PIN_SOURCE_3);
    exti_init(GPIO_LINE_3, EXTI_INTERRUPT, EXTI_TRIG_BOTH);
    /* EXTI3 interrupt handler should be set up in FreeRTOS tasks */
    log_info("BoardPro: EXTI3 configured for SOS button");
}

/* ----------------------------------------------------------------
 * Master init: power rails -> peripherals -> everything
 * ---------------------------------------------------------------- */

void board_pro_init_all(void)
{
    board_pro_clock_init();
    board_pro_gpio_init();

    /* Power up analog section first (sensors need clean supply) */
    board_pro_analog_power_on();
    vTaskDelay(pdMS_TO_TICKS(2));

    /* Power up digital section */
    board_pro_digital_power_on();
    vTaskDelay(pdMS_TO_TICKS(1));

    board_pro_uart_debug_init(BOARD_PRO_DEBUG_UART_BAUD);
    board_pro_uart_gnss_init(BOARD_PRO_GNSS_BAUD);
    board_pro_uart_cat1_init(9600);

    board_pro_i2c_init();
    board_pro_spi_init();
    board_pro_timer_init();
    board_pro_exti_init();

    log_info("BoardPro: All peripherals initialized");
}
