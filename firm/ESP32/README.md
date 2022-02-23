# ESP32 Firmware

ESP32で周波数監視をするファームウェアです。

## build

Arduino でビルドできるはず。

## config

`config.h-dist`をコピるなどして`config.h`を作り、WiFiのSSIDとかを入れてください。

周波数パルスの入力は無邪気にGPIO34(`mcpwm_gpio_init()`の引数)にしていますが、変えても行けるはず。

## LCDライブラリについて

オリジナルは[キャラクタLCD表示用ライブラリ](http://www.inoshita.jp/freo/view/396)にあります。

## 回路図例

![esp32_schematics](https://user-images.githubusercontent.com/1270667/155331256-906107e5-818b-4b90-8b88-3d23f6a1c457.png)
