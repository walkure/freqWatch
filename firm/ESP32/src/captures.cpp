#include <Arduino.h>

#include <captures.h>
#include <constexpr.h>

void setupCapture()
{
    // GPIO 34
    mcpwm_gpio_init(MCPWM_UNIT_0, MCPWM_CAP_0, 34);
    mcpwm_capture_config_t input_capture_setup;
    input_capture_setup.cap_edge = MCPWM_NEG_EDGE;
    input_capture_setup.cap_prescale = 1;
    input_capture_setup.capture_cb = input_capture_callback;
    input_capture_setup.user_data = NULL;
    mcpwm_capture_enable_channel(MCPWM_UNIT_0, MCPWM_SELECT_CAP0, &input_capture_setup);
}

portMUX_TYPE caputureCallbackMux = portMUX_INITIALIZER_UNLOCKED;
volatile uint32_t _cap_value = 0;

volatile uint32_t _cap_sum_value = 0;
volatile uint16_t _cap_sum_count = 0;

volatile uint32_t _cap_filtered_sum_value = 0;
volatile uint16_t _cap_filtered_sum_count = 0;

constexpr uint32_t filter_floor = APB_CLK_FREQ / (pivot_freq + 6); // 1230769
constexpr uint32_t filter_ceil = APB_CLK_FREQ / (pivot_freq - 6);  // 1777777

bool IRAM_ATTR input_capture_callback(mcpwm_unit_t mcpwm, mcpwm_capture_channel_id_t cap_channel, const cap_event_data_t *edata, void *user_data)
{
    static uint32_t old_cap_value = 0;

    portENTER_CRITICAL_ISR(&caputureCallbackMux);
    const uint32_t value = edata->cap_value;
    _cap_value = value - old_cap_value;
    old_cap_value = value;
    _cap_sum_value += _cap_value;
    _cap_sum_count++;

    if (_cap_value > filter_floor && _cap_value < filter_ceil)
    {
        _cap_filtered_sum_value += _cap_value;
        _cap_filtered_sum_count++;
    }

    portEXIT_CRITICAL_ISR(&caputureCallbackMux);

    return false;
}

void get_captured_freqs(float *simple, float *filtered)
{
    portENTER_CRITICAL_ISR(&caputureCallbackMux);
    auto sum_value = _cap_sum_value;
    auto sum_count = _cap_sum_count;
    auto filtered_sum_value = _cap_filtered_sum_value;
    auto filtered_sum_count = _cap_filtered_sum_count;

    _cap_sum_count = 0;
    _cap_sum_value = 0;
    _cap_filtered_sum_value = 0;
    _cap_filtered_sum_count = 0;
    portEXIT_CRITICAL_ISR(&caputureCallbackMux);

    auto average_count = (float)sum_value / sum_count;
    *simple = (float)APB_CLK_FREQ / average_count;

    if (filtered_sum_count > 0)
    {
        average_count = (float)filtered_sum_value / filtered_sum_count;
        *filtered = (float)APB_CLK_FREQ / average_count;
    }
    else
    {
        *filtered = 0.;
    }
}
