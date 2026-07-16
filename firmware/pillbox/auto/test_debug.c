#include <stdio.h>
#include "dispensing.h"
#include "opto_sensor.h"
#include "motor_control.h"
#include "state_machine.h"

static int g_motor_step_calls = 0;
bool motor_control_step(uint8_t steps) { g_motor_step_calls++; return true; }
void motor_control_init(void) { g_motor_step_calls = 0; }
bool motor_control_is_ready(void) { return true; }
void motor_control_home(void) { g_motor_step_calls = 0; }

static void mock_delay_fn(uint32_t ms) { (void)ms; }

int main(void) {
    printf("1. init\n");
    opto_sensor_init();
    motor_control_init();
    state_machine_init();
    state_machine_force_state(STATE_IDLE);
    
    printf("2. set mock\n");
    dispensing_set_mock_delay(mock_delay_fn);
    dispensing_set_mock_poll_hook(NULL);
    opto_sensor_set_mock_state(false);
    
    printf("3. calling dispense_medication(0, 1)\n");
    dispense_result_t result = dispense_medication(0, 1);
    
    printf("4. result=%d\n", (int)result);
    return 0;
}
