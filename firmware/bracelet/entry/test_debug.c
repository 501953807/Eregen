#include <stdio.h>
#include <stdint.h>
#include <stdbool.h>
#define TEST_MODE
#include "sos_button.c"

int main(void) {
    sos_init();
    
    /* Simulate 10 iterations */
    for (int i = 1; i <= 10; i++) {
        sos_set_mock_state(true);
        sos_task();
        printf("iter=%d stable=%u hold_ms=%lu pressed=%d long=%d\n",
               i, s_sos.stable_count, (unsigned long)s_sos.hold_time_ms,
               s_sos.just_pressed, s_sos.just_long_press);
    }
    
    /* Release */
    sos_set_mock_state(false);
    sos_task();
    printf("released: hold_ms=%lu\n", (unsigned long)s_sos.hold_time_ms);
    
    return 0;
}
