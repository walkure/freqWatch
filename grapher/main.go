package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"time"

	"golang.org/x/exp/slog"

	"gonum.org/v1/plot/vg"
)

func main() {
	flag.Parse()
	os.Exit(execute())
}

var promURL = flag.String("promURL", "http://localhost:9090/", "Prometheus server url")
var resultPath = flag.String("graphPath", "graph.png", "result file path")
var graphConfigPath = flag.String("graphConfig", "graph.json", "graph config file path")

func execute() int {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		slog.Error("faiure to load TZ", slog.Any("error", err))
		return -1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conf, err := os.ReadFile(*graphConfigPath)
	if err != nil {
		slog.Error("failure to load graph config file", slog.Any("error", err))
		return -1
	}

	var params GraphConfig
	if err := json.Unmarshal(conf, &params); err != nil {
		slog.Error("failure to parse graph config file", slog.Any("error", err))
		return -1
	}

	now := time.Now()
	//now := time.Unix(1691903795, 0).In(jst)

	matrix, err := getMetrics(ctx, *promURL, now, time.Second*30, time.Hour*24)
	if err != nil {
		slog.Error("failure to load metrics", slog.Any("error", err))
		return -1
	}

	p, err := plotMatrix(matrix, jst, now, params)
	if err != nil {
		slog.Error("failure to plot metrics", slog.Any("error", err))
		return -1
	}

	if err := p.Save(16*vg.Inch, 5*vg.Inch, *resultPath); err != nil {
		slog.Error("failure to save graph file", slog.Any("error", err))
		return -1
	}

	return 0
}
