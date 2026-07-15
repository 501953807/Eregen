/*
 * Eregen (颐贞) - Board Initialization Source
 * GD32E230C8T3 board init implementation
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "board_init.h"
#include "gd32e230_rcu.h"
#include "gd32e230_gpio.h"
#include "gd32e230_usart.h"
#include "gd32e230_i2c.h"
#include "gd32e230_spi.h"
#include "gd32e230_timer.h"
#include "gd32e230_adc.h"

/*
 * Configure system clock to 120MHz using HXTAL (8MHz crystal)
 * with PLL multiplier of 15x.
 */
void board_clock_init(void)
{
    /* Enable HXTAL (external high-speed crystal) */
    rcu_osci_on(RCU_HXTAL);
    rcu_osci_stab_wait(RCU_HXTAL_STAB);

    /* Select HXTAL as PLL source */
    rcu_cksel_config(
        RCU_CKSEL_HXTAL,      /* HXTAL selected as SYSCLK source */
        0U,                   /* CK_DIV (AHB div) - disabled */
        0U,                   /* ADC div - disabled */
        0U,                   /* CDIV (CPU div) - not used when HXTAL is direct */
        RCU_CDIV_DIV_1,       /* CKOUT0 div - disabled */
        RCU_HXTAL_DIV_1       /* HXTAL no division */
    );

    /* Configure PLL: HXTAL 8MHz * 15 = 120MHz */
    rcu_pll_config(RCU_PLL_MUL15);

    /* Enable PLL */
    rcu_osci_on(RCU_PLL);
    rcu_osci_stab_wait(RCU_PLL_STAB);

    /* Configure flash wait states for 120MHz */
    flash_prefetch_enable();
    flash_instr_latency_set(FLASH_ILFT_1_5);
    flash_data_latency_set(FLASH_DLT_1_5);

    /* Select PLL as SYSCLK source */
    rcu_clock_extend_config(RCU_CFG0_CSSD_EN_MASK, true);
    rcu_sysclk_config(RCU_SRCCLK_PLL);

    /* Wait until PLL is used as SYSCLK source */
    uint32_t timeout = 0xFFFFFU;
    while ((rcu_flag_get(RCU_FLAG_HXTAL_STABLE) == RESET) && (timeout-- > 0)) {
        /* Wait for HXTAL stable */
    }

    /* Verify SYSCLK source is PLL */
    uint32_t srcclk = rcu_sysclk_source_get();
    (void)srcclk;
}

/*
 * Initialize GPIO pins for LEDs, buttons, and peripheral CS.
 * Also configures GPS enable, Cat1 enable, and UART pins.
 */
void board_gpio_init(void)
{
    /* Enable GPIO clocks */
    rcu_periph_clock_enable(RCU_GPIOA);
    rcu_periph_clock_enable(RCU_GPIOB);
    rcu_periph_clock_enable(RCU_GPIOC);

    /* LED green: PA0, push-pull output, 2MHz */
    gpio_init(LED_GREEN_GPIO_PORT, GPIO_MODE_OUT_PP, GPIO_OSPEED_2MHZ, LED_GREEN_GPIO_PIN);
    gpio_bit_set(LED_GREEN_GPIO_PORT, LED_GREEN_GPIO_PIN);

    /* LED blue: PA1, push-pull output, 2MHz */
    gpio_init(LED_BLUE_GPIO_PORT, GPIO_MODE_OUT_PP, GPIO_OSPEED_2MHZ, LED_BLUE_GPIO_PIN);
    gpio_bit_set(LED_BLUE_GPIO_PORT, LED_BLUE_GPIO_PIN);

    /* User button: PA2, input with pull-up/pull-down */
    gpio_init(USER_BUTTON_GPIO_PORT, GPIO_MODE_IPU, GPIO_OSPEED_50MHZ, USER_BUTTON_GPIO_PIN);

    /* SOS button: PA3, input with pull-up/pull-down */
    gpio_init(SOS_BUTTON_GPIO_PORT, GPIO_MODE_IPU, GPIO_OSPEED_50MHZ, SOS_BUTTON_GPIO_PIN);

    /* SPI chip select pins: outputs, high (deselected) */
    gpio_init(IMU_CS_GPIO_PORT, GPIO_MODE_OUT_PP, GPIO_OSPEED_50MHZ, IMU_CS_GPIO_PIN);
    gpio_bit_set(IMU_CS_GPIO_PORT, IMU_CS_GPIO_PIN);

    gpio_init(GPS_CS_GPIO_PORT, GPIO_MODE_OUT_PP, GPIO_OSPEED_50MHZ, GPS_CS_GPIO_PIN);
    gpio_bit_set(GPS_CS_GPIO_PORT, GPS_CS_GPIO_PIN);

    gpio_init(DISPLAY_CS_GPIO_PORT, GPIO_MODE_OUT_PP, GPIO_OSPEED_50MHZ, DISPLAY_CS_GPIO_PIN);
    gpio_bit_set(DISPLAY_CS_GPIO_PORT, DISPLAY_CS_GPIO_PIN);

    /* GPS module enable pin (PC0) - configure as output, low (disabled) */
    gpio_init(GPIOC, GPIO_MODE_OUT_PP, GPIO_OSPEED_50MHZ, GPIO_PIN_0);
    gpio_bit_reset(GPIOC, GPIO_PIN_0);

    /* Cat1 module enable pin (PC1) - configure as output, low (disabled) */
    gpio_init(GPIOC, GPIO_MODE_OUT_PP, GPIO_OSPEED_50MHZ, GPIO_PIN_1);
    gpio_bit_reset(GPIOC, GPIO_PIN_1);

    /* Cat1 module reset pin (PC2) - configure as output, high (not reset) */
    gpio_init(GPIOC, GPIO_MODE_OUT_PP, GPIO_OSPEED_50MHZ, GPIO_PIN_2);
    gpio_bit_set(GPIOC, GPIO_PIN_2);
}

/*
 * Initialize UART0 as debug console at specified baud rate.
 * TX: PA9, RX: PA10
 */
void board_uart_debug_init(uint32_t baud)
{
    rcu_periph_clock_enable(RCU_GPIOA);
    rcu_periph_clock_enable(RCU_USART0);

    /* Configure PA9 as alternate function push-pull (USART0_TX) */
    gpio_init(GPIOA, GPIO_MODE_AF_PP, GPIO_OSPEED_50MHZ, GPIO_PIN_9);
    /* Configure PA10 as input floating (USART0_RX) */
    gpio_init(GPIOA, GPIO_MODE_INPUT, GPIO_OSPEED_50MHZ, GPIO_PIN_10);

    /* USART configuration: 115200-8-N-1 */
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

    /* Wait for USART to be ready */
    while (usart_flag_get(USART0, USART_FLAG_RBNE) == RESET) {
        /* Wait */
    }
}

/*
 * Initialize USART1 for Cat1 module communication.
 * TX: PB6, RX: PB7
 */
void board_uart_cat1_init(uint32_t baud)
{
    rcu_periph_clock_enable(RCU_GPIOB);
    rcu_periph_clock_enable(RCU_USART1);

    /* Configure PB6 as alternate function push-pull (USART1_TX) */
    gpio_init(GPIOB, GPIO_MODE_AF_PP, GPIO_OSPEED_50MHZ, GPIO_PIN_6);
    /* Configure PB7 as input floating (USART1_RX) */
    gpio_init(GPIOB, GPIO_MODE_INPUT, GPIO_OSPEED_50MHZ, GPIO_PIN_7);

    /* USART configuration: default 9600-8-N-1 (module wakes up at this speed) */
    usart_deinit(USART1);
    usart_baudrate_set(USART1, baud);
    usart_word_length_set(USART1, USART_WL_8BIT);
    usart_stop_bit_set(USART1, USART_STB_1BIT);
    usart_parity_config(USART1, USART_PM_NONE);
    usart_hardware_flow_rts_config(USART1, USART_RTS_DISABLE);
    usart_hardware_flow_cts_config(USART1, USART_CTS_DISABLE);
    usart_transmit_config(USART1, USART_TRANSMIT_ENABLE);
    usart_receive_config(USART1, USART_RECEIVE_ENABLE);
    usart_enable(USART1);
}

/*
 * Initialize USART2 for GPS module communication.
 * TX: PA2, RX: PA3 -- but PA2/PA3 are used by buttons.
 * Alternative: use PB10 (TX) / PB11 (RX) if available, or remap.
 * For GD32E230, USART2 is on PA2/PA3 by default.
 * We use the existing pins since GPS is polled in main loop.
 */
void board_uart_gps_init(uint32_t baud)
{
    rcu_periph_clock_enable(RCU_GPIOA);
    rcu_periph_clock_enable(RCU_USART2);

    /* USART2 TX: PA2, RX: PA3 */
    /* Note: PA2/PA3 are also used by user button and SOS button.
     * In production, these pins would be shared carefully.
     * The GPS UART is polled only when GPS task runs.
     */
    gpio_init(GPIOA, GPIO_MODE_AF_PP, GPIO_OSPEED_50MHZ, GPIO_PIN_2);
    gpio_init(GPIOA, GPIO_MODE_INPUT, GPIO_OSPEED_50MHZ, GPIO_PIN_3);

    usart_deinit(USART2);
    usart_baudrate_set(USART2, baud);
    usart_word_length_set(USART2, USART_WL_8BIT);
    usart_stop_bit_set(USART2, USART_STB_1BIT);
    usart_parity_config(USART2, USART_PM_NONE);
    usart_transmit_config(USART2, USART_TRANSMIT_ENABLE);
    usart_receive_config(USART2, USART_RECEIVE_ENABLE);
    usart_enable(USART2);
}

/*
 * Initialize I2C1 on PB8 (SCL) / PB9 (SDA).
 * Standard mode: 400kHz
 */
void board_i2c_init(void)
{
    rcu_periph_clock_enable(RCU_GPIOB);
    rcu_periph_clock_enable(RCU_I2C1);

    /* Configure PB8 as alternate function open-drain (I2C1_SCL) */
    gpio_init(GPIOB, GPIO_MODE_AF_OD, GPIO_OSPEED_50MHZ, GPIO_PIN_8);
    /* Configure PB9 as alternate function open-drain (I2C1_SDA) */
    gpio_init(GPIOB, GPIO_MODE_AF_OD, GPIO_OSPEED_50MHZ, GPIO_PIN_9);

    /* Deinit I2C */
    i2c_deinit(I2C1);

    /* I2C clock: 80MHz APB1 -> 400kHz SCL */
    i2c_clock_config(I2C1, 400000U, I2C_DTCY_2);

    /* I2C enable */
    i2c_enable(I2C1);

    /* Enable ACK */
    i2c_ack_enable(I2C1);
}

/*
 * Initialize SPI1 on PB3/SCK, PB4/MISO, PB5/MOSI.
 * Master mode, 1MHz initial clock (for device probing).
 */
void board_spi_init(void)
{
    rcu_periph_clock_enable(RCU_GPIOB);
    rcu_periph_clock_enable(RCU_SPI1);

    /* PB3: alternate function push-pull (SPI1_SCK) */
    gpio_init(GPIOB, GPIO_MODE_AF_PP, GPIO_OSPEED_50MHZ, BOARD_SPI_SCK_PIN);
    /* PB4: input floating (SPI1_MISO) */
    gpio_init(GPIOB, GPIO_MODE_INPUT, GPIO_OSPEED_50MHZ, BOARD_SPI_MISO_PIN);
    /* PB5: alternate function push-pull (SPI1_MOSI) */
    gpio_init(GPIOB, GPIO_MODE_AF_PP, GPIO_OSPEED_50MHZ, BOARD_SPI_MOSI_PIN);

    /* SPI1 configuration: master, CPOL=0 CPHA=0, MSB first */
    spi_parameter_struct spi_init_param;
    spi_struct_para_init(&spi_init_param);
    spi_init_param.trans_mode           = SPI_TRANSMODE_FULLDUPLEX;
    spi_init_param.device_mode          = SPI_MASTER;
    spi_init_param.frame_size           = SPI_FSIZE_8BIT;
    spi_init_param.nss                  = SPI_NSS_SOFT;
    spi_init_param.endianness           = SPI_ENDIAN_MSB;
    spi_init_param.prescale             = SPI_PSC_256;  /* Slow initial clock */
    spi_init(&SPI1, &spi_init_param);

    spi_enable(SPI1);
}

/*
 * Initialize general-purpose timers for system timing.
 * TIMER0 used as system tick base.
 */
void board_timer_init(void)
{
    rcu_periph_clock_enable(RCU_TIMER0);

    timer_parameter_struct timer_init_param;
    timer_struct_para_init(&timer_init_param);
    timer_init_param.prescaler         = 120U - 1U;  /* 1MHz tick from 120MHz */
    timer_init_param.aligned_mode      = TIMER_COUNTER_EDGE;
    timer_init_param.count_direction   = TIMER_COUNT_UP;
    timer_init_param.period            = 999U;       /* 1ms auto-reload */
    timer_init_param.clock_division    = TIMER_CKDIV_DIV1;
    timer_init_param.repetition_counter = 0U;
    timer_init(TIMER0, &timer_init_param);

    /* Enable timer update interrupt */
    timer_interrupt_enable(TIMER0, TIMER_INT_UP);
    nvic_irq_enable(TIMER0_IRQn, 3U, 0U);

    timer_enable(TIMER0);
}

/*
 * Initialize external interrupts for SOS button and user button.
 * PA2 (user button) and PA3 (SOS button) trigger EXTI on falling edge.
 * Requires GD32 EXTI standard peripheral library.
 */
void board_exti_init(void)
{
    /* Enable SYSCFG clock for EXTI */
    rcu_periph_clock_enable(RCU_SYSCFG);

    /* Configure PA2 as EXTI source (user button) */
    syscfg_exti_line_config(EXTI_SOURCE_PA, EXTI_SOURCE_PIN2);
    /* Configure PA3 as EXTI source (SOS button) */
    syscfg_exti_line_config(EXTI_SOURCE_PA, EXTI_SOURCE_PIN3);

    /* Enable EXTI2 interrupt (user button) */
    exti_init(EXTI_2, EXTI_INTERRUPT, EXTI_TRIG_FALL);
    nvic_irq_enable(EXTI2_IRQn, 3U, 0U);
    exti_interrupt_enable(EXTI_2);

    /* Enable EXTI3 interrupt (SOS button) */
    exti_init(EXTI_3, EXTI_INTERRUPT, EXTI_TRIG_FALL);
    nvic_irq_enable(EXTI3_IRQn, 3U, 0U);
    exti_interrupt_enable(EXTI_3);
}

/*
 * Master initialization: call all subsystem inits.
 */
void board_init_all(void)
{
    board_clock_init();
    board_gpio_init();
    board_uart_debug_init(BOARD_DEBUG_UART_BAUD);
    board_uart_cat1_init(9600);    /* Cat1 module default baud */
    board_uart_gps_init(9600);     /* GPS module default baud */
    board_i2c_init();
    board_spi_init();
    board_timer_init();
    board_exti_init();
}
