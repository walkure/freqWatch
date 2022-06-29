#include <Arduino.h>

void dumpSerial(float v1, float v2)
{
    char buff[16];
    dtostrf(v1, 6, 4, buff);
    Serial.print(buff);
    Serial.print(" ");
    dtostrf(v2, 6, 4, buff);
    Serial.println(buff);
}

inline char itoc(const uint8_t i)
{
    return i < 10 ? i + 0x30 : i + 0x57; // 0x57 = 0x61('a') - 0x0a
    // return i < 10 ? i+0x30 : i+0x37; // 0x37 = 0x41('A') - 0x0a
}

// bytes列をhex stringに変換
// strbufのサイズは len*2+1
const char *BytesToHexStr(const uint8_t *const data, char *strbuf, size_t len)
{
    for (size_t i = 0; i < len; i++)
    {
        strbuf[i * 2 + 1] = itoc(data[i] & 0xf);
        strbuf[i * 2] = itoc(data[i] >> 4);
    }
    strbuf[len * 2] = '\0';
    return strbuf;
}

hw_timer_t *TimerSetup(int id, void (*handler)(), int usec)
{
    hw_timer_t *timer;
    timer = timerBegin(id, APB_CLK_FREQ / (1000 * 1000), true);
    timerAttachInterrupt(timer, handler, true);
    timerAlarmWrite(timer, usec, true);
    timerAlarmEnable(timer);

    return timer;
}
