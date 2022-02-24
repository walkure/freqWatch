# Atmel AVR Firmware

Atmel AVR(ATTinyとか)で周波数観測をするファーム

とものさんのATTiny2313での観測ファームからブザーまわりを消して、インプットキャプチャでのパルス幅に上限下限を設定しただけ。

- [ATTINY2313で商用電源周波数監視](https://tomono.tokyo/2021/01/18/9177/) 2021年1月18日
- [ATTINY版ソフトウェア修正](https://tomono.tokyo/2021/03/31/9596/) 2021年3月31日

## build and write

Microchip Studio(Version 7.0.2052)でビルドして適当なライタつないで書き込みました。

## sender.py

Raspberry PiのUARTに繋いで観測データを流し込み、それを随時アップロードするスクリプトです。`asyncio`で書いたので`apt`で諸々入れる必要がありそう。

設定は `sender.py` にベタ書きです。そのうちsystemdのunit file書きます。

Ambientへのアップロードを使わない場合は、`__send_ambient_async`読んでるとこをコメントアウトしてください。
