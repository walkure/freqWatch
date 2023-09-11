# freqWatch

周波数観測情報を受信してリアルタイムに表示する試みです。

## build

receiverディレクトリに移動して ``` go build ```

## configuration

環境変数から取ってきます。

- PORT: HTTPを受けるポート (default=`8080`)
- LISTEN: HTTPを受けるホストアドレス (default=`""` INADDR_ANY相当)
- RECV_PATH: ESP32等から周波数情報をを受けるパス (default=`/frecv`)
- SHARE_KEY: ESP32等から周波数情報を受けるときの共通キー (必須)
- METRICS_PATH: OpenMetrics Exporterのパス (default=`/metrics`)
- DUMP_BUFFER: 初回表示時に送るデータポイント数(default=60*10)
- METRIC_BUFFER: OpenMetrics Exporterで処理するデータポイント数(default=30)

### nginx

https通信する場合はnginxのようなリバプロを表に置きます(Let's Encrypt使うとか考えるとそのほうが楽)。

ESP32でTLS通信はちょっと大変なので、データ送信だけhttpで受ける場合の設定例。

```nginx
# http
server{
        listen 80;
        listen [::]:80;

        location /reciverPath/ {
                proxy_pass http://127.0.0.1:8080/frecv?$args;
        }
}

# https
server{
        listen 443;
        listen [::]:443;
        location /freqWatch/ {
                root /var/www/;
        }

        location /freqWatch/d/ {
                if ( $uri = '/freqWatch/d/frecv' ){
                        return 404;
                }
                if ( $uri = '/freqWatch/d/metrics' ){
                        return 404;
                }
                proxy_http_version 1.1;
                proxy_set_header Upgrade $http_upgrade;
                proxy_set_header Connection "Upgrade";
                proxy_set_header Host $host;
                proxy_pass http://127.0.0.1:8080/;
        }
}
```

### systemd

きょうびdaemonはsystemdで管理するので、Unit fileを雑に書く。


```systemd
[Unit]
Description=freq receiver
After=network-online.target nginx.service

[Service]
User=nobody
ExecStart=/usr/local/bin/freq_recv
Environment=SHARE_KEY=fds998fsdfscjdje
Restart=always
RestartSec=30s

[Install]
WantedBy=multi-user.target
```

### Prometheus a.k.a OpenMetrics

`http://localhost:8080/metrics` で OpenMetrics Exporterが動いています。
中長期のデータはPrometheusのようなもので取っておくほうが楽です。

実際にスクレイピングさせると`Error on ingesting out-of-order samples` というwarnログがもりもり流れるので、
取得済みデータを再度取得しないようにクエリパラメタ`flush`をつけてください。Prometheusの場合は以下の設定を`promethues.yml`の`scrape_configs`内`job`に追加します。

```yaml
    params:
      flush: [true]
```

1秒単位のデータは情報量が多すぎるのでまとめる、というときはクエリパラメタ`mode=average`をつけると平均値を返すようになります(こちらには`flush`はありません)。

```yaml
    params:
      mode: ['average']
```

このとき、平均は`METRIC_BUFFER`に保存されているデータで取るので、スクレイピング間隔などに合わせて適宜調整すると良いでしょう。

