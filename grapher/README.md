# Grapher

Prometheus(やVictoriaMetrics)からメトリクスを取ってきて、電源周波数変動グラフを描きます。

## build

```:shell
go build .
```

## configuration

起動時引数と設定ファイルがあります。

### 起動時引数

- `-promURL` Prometheus server URL (default `http://localhost:9090/`)
- `-graphConfig` グラフ設定ファイルの場所 (default "graph.json")
- `-graphPath` 吐き出すグラフファイルの場所 (default "graph.png")

### グラフ設定ファイル

例を貼って説明します。

```:json
{
    "tokyo": {
        "legend": "50Hz:TEPCO",
        "origin": 50,
        "color": "#DAB300"
    },
    "kyoto": {
        "legend": "60Hz:KEPCO",
        "origin": 60,
        "color": "#235BC8"
    }
}
```

設定のキーになっている`tokyo`とか`kyoto`は、Prometheusのメトリクスラベル`place`に設定されている値を想定しています。`/firm`ディレクトリにあるファームの中で`place`を設定していて、これをそのままPrometheusに書いています。

- `legend`はグラフの値につく説明です。
- `origin`はグラフに描く変動の基準です。地域によって`50`/`60`を設定することになります。
- `color`はグラフの線をカラーコードで表現します。

## systemd設定例

```systemd.service:ini
[Unit]
Description=freqgraph generator

[Service]
Type=oneshot
ExecStart=/usr/local/bin/freq-grapher -promURL=http://localhost:8428 -graphConfig=/usr/local/etc/freqgraph.json -graphPath=/var/www/freqWatch/freqgraph.png

[Install]
WantedBy=multi-user.target
```

```systemd.timer:ini
[Unit]
Description=freqgraph timer

[Timer]
OnCalendar=*-*-* *:03,13,23,33,43,53:00
Persistent=true

[Install]
WantedBy=timers.target
```
