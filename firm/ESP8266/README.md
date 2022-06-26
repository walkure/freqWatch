# ESP8266 Firmware

ESP8266で周波数監視をするファームウェアです。コアが同じESP8285でも動きます。

## build

[PlatformIO](https://platformio.org/) でビルドできます。

## config

`config.cpp-dist`をコピるなどして`config.cpp`を作り、WiFiのSSIDとかを入れてください。

周波数パルスの入力は適当にGPIO12(`FREQ_INPUT_PIN`で定義)にしていますが、特にこだわりはないです。IO0とかにするとしんどそう。

## LCDライブラリについて

オリジナルは[キャラクタLCD表示用ライブラリ](http://www.inoshita.jp/freo/view/396)にあります。

## 回路図例

![schematic.png](https://user-images.githubusercontent.com/1270667/169705496-27c31516-46e5-45e6-a244-8673f12e2baf.png)

## 使用ライブラリ

- [AsyncHTTPRequest_Generic](https://github.com/khoih-prog/AsyncHTTPRequest_Generic)
  - [ESPAsyncTCP](https://github.com/me-no-dev/ESPAsyncTCP)

### ライブラリのバグについて

ESPAsyncTCPとAsyncHTTPRequest_Genericともにバグが残っています。とくにESPAsyncTCPは長期間運用するには辛いバグがあるのでオリジナルの使用はおすすめできない。

#### ESPAsyncTCP

正常に送信完了できなかった場合の処理にバグがあって、いちど正常に送信できないと挙動が怪しくなる。

- <https://github.com/serek4/ESPAsyncTCP>
- <https://github.com/walkure/ESPAsyncTCP>

#### AsyncHTTPRequest_Generic
ログを有効にした状態で接続先がWindowsのとき、内部状態が期待しない遷移をしてクラッシュした(Linux相手だと再現しない…。)

- <https://github.com/walkure/AsyncHTTPRequest_Generic>
