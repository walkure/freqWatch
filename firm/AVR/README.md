# Atmel AVR Firmware

Atmel AVR(ATTinyとか)で周波数観測をするファーム

とものさんのATTiny2313での観測ファームを基に改修しました。

- [ATTINY2313で商用電源周波数監視](https://tomono.tokyo/2021/01/18/9177/) 2021年1月18日
- [ATTINY版ソフトウェア修正](https://tomono.tokyo/2021/03/31/9596/) 2021年3月31日

## 主な機能

- Timer1のインプットキャプチャ機能(ICP1ピン:PD6)で信号エッジ間隔を測定
- AVRノイズキャンセラ有効化
- 10ms未満のパルス間隔をフィルタリングしてノイズを除去
- 50サンプルごとに周波数を計算し、前回50サンプルとの移動平均(計100サンプル)で算出

## build and write

Microchip Studio(Version 7.0.2052)でビルドして適当なライタつないで書き込みました。

なお、PlatformIO(on Windows)でも書き込めるのを確認しています。

なお、ライタは[Pololu USB AVRプログラマ v2.1](https://www.switch-science.com/products/3870)を使っています。

### 補正

`clk_base`に32足していますが、わたしの環境で50Hz信号を入れて補正したものです。なので、これは個体補正値になります。

## sender.py

Raspberry PiのUARTに繋いで観測データを流し込み、それを随時アップロードするスクリプトです。`asyncio`で書いたので`apt`で諸々入れる必要がありそう。

設定は `sender.py` にベタ書きです。そのうちsystemdのunit file書きます。

Ambientへのアップロードを使わない場合は、`__send_ambient_async`読んでるとこをコメントアウトしてください。

※GPIOがPL011でなくminiUARTに繋がっていると取りこぼしてデータがもりもり化けるので、気をつけてください。
