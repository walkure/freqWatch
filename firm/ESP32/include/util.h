#pragma once
#include <stdint.h>
#include <stddef.h>

const char *BytesToHexStr(const uint8_t *const data, char *strbuf, size_t len);
hw_timer_t *TimerSetup(int id, void (*handler)(), int usec);
