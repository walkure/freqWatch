# ESP32 Firmware

ESP32で周波数監視をするファームウェアです。

## build

Arduino でビルドできるはず。

## config

`config.h-dist`をコピるなどして`config.h`を作り、WiFiのSSIDとかを入れてください。

周波数パルスの入力は無邪気にGPIO34(`mcpwm_gpio_init()`の引数)にしていますが、変えても行けるはず。

## LCDライブラリについて

オリジナルは[キャラクタLCD表示用ライブラリ](http://www.inoshita.jp/freo/view/396)にあります。