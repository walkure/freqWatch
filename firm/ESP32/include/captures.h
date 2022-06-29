#pragma once

#include <Arduino.h>
#include <driver/mcpwm.h>

void setupCapture();
void get_captured_freqs(float *simple, float *filtered);
bool IRAM_ATTR input_capture_callback(mcpwm_unit_t mcpwm, mcpwm_capture_channel_id_t cap_channel, const cap_event_data_t *edata, void *user_data);

