// original: http://www.inoshita.jp/freo/view/396

#include "AQM1602LCD.h"
#include <Wire.h>

constexpr uint8_t AQMI2CLCD_ADDR = 0x3E;

void AQMI2CLCDClass::_writeCommand(byte command) {
    Wire.beginTransmission(AQMI2CLCD_ADDR);    // スタートコンディション
    Wire.write(0x00);                        // control byte の送信(コマンドを指定)
    Wire.write(command);                    // command byte の送信
    Wire.endTransmission();                    // ストップコンディション
    delay(20);
}

void AQMI2CLCDClass::setup() {
    setup(3);
}
void AQMI2CLCDClass::setup(int volt) {
    
    delay(100);
    _writeCommand(0x38); // Function set
    _writeCommand(0x39); // IS=1
    _writeCommand(0x14); // Internal OSC frequency

    if(volt == 3) {
        // 3.3V
        _writeCommand(0x73); // Contrast set
        _writeCommand(0x56); // POWER/ICON/Contrast control
    } else /* if(volt == 5) */ {
        // 5V
        _writeCommand(0x7A); // Contrast set
        _writeCommand(0x52); // POWER/ICON/Contrast control
    }
    _writeCommand(0x6C); // Follower control
    _writeCommand(0x38); // Function set
    _writeCommand(0x01); // Clear Display
    _writeCommand(0x0C); // Display ON
    _writeCommand(0x06); // Entry Mode
}

// LCDモジュールの画面をクリア
//
void AQMI2CLCDClass::clearScreen(void) {
    _writeCommand(0x01);                // Clear Display
}

// LCDモジュール画面内のカーソル位置を移動
//    col : 横(列)方向のカーソル位置(0-15)
//    row : 縦(行)方向のカーソル位置(0-1)
//
void AQMI2CLCDClass::setLocate(int col, int row) {
    static int row_offsets[] = { 0x00, 0x40 } ;
    // Set DDRAM Adddress : 00H-0FH,40H-4FH
    _writeCommand(0x80 | (col + row_offsets[row]));
}

// LCDにデータを1バイト出力
//      c :  出力する文字データを指定
//
size_t AQMI2CLCDClass::write(uint8_t c) {
  Wire.beginTransmission(AQMI2CLCD_ADDR);
  Wire.write(0x40);                        // control byte の送信(データを指定)
  Wire.write(c);                        // data byte の送信
  Wire.endTransmission();                    // ストップコンディション
  delay(20);
	return 1;
}

void AQMI2CLCDClass::scroll() {
	_writeCommand(0x1B);
}

void AQMI2CLCDClass::returnHome() {
	_writeCommand(0x02);
}

void AQMI2CLCDClass::createChar(const uint8_t location, const byte* data)
{
	// address maybe 00 -> 07
	_writeCommand(0x40 | (location << 3));
	
	//write 8bytes
    Wire.beginTransmission(AQMI2CLCD_ADDR);
    Wire.write(0x40);                        // control byte の送信(データを指定)
    Wire.write(data,8);                        // data byte の送信
    Wire.endTransmission();                    // ストップコンディション
    delay(1);
}

size_t AQMI2CLCDClass::write(const uint8_t *buffer, size_t size){
    Wire.beginTransmission(AQMI2CLCD_ADDR);
    for(int i = 0 ; i < size ; i++){
      if(i < size - 1){
        Wire.write(0xC0); // Co = 1 RS = 1
      }else{
        Wire.write(0x40); // Co = 0 RS = 1
      }
      Wire.write(buffer[i]);
    }
    Wire.endTransmission();
    delay(20);
    
    return size;
}


AQMI2CLCDClass AQMI2CLCD;
