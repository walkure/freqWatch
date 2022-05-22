// original: http://www.inoshita.jp/freo/view/396

#pragma once

#include <Arduino.h>
#include <Print.h>

class AQMI2CLCDClass : public Print {
  public:
    void setup();
    void setup(int);
    void clearScreen(void);
    void setLocate(int, int);
    size_t write(uint8_t);
    size_t write(const uint8_t *buffer, size_t size);
    void scroll();
    void returnHome();
    void createChar(const uint8_t ,const byte*);
  private:
    static void _writeCommand(byte command);
};

extern AQMI2CLCDClass AQMI2CLCD;
