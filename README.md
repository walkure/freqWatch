# freq_recv

周波数観測情報を受信してリアルタイムに表示する試みです。

## configuration

環境変数から取ってきます。

- PORT: HTTPを受けるポート (default=8080)
- LISTEN: HTTPを受けるホストアドレス(default="" INADDR_ANY相当)
- RECV_PATH: ESP32等から周波数情報をを受けるパス(default=/frecv)
- SHARE_KEY: ESP32等から周波数情報を受けるときの共通キー(必須)
