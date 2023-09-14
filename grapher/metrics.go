package main

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"golang.org/x/exp/slog"
)

func getMetrics(ctx context.Context, url string, endAt time.Time, step, width time.Duration) (model.Matrix, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	api := v1.NewAPI(client)

	query := "median_over_time(power_freq[10m])"

	result, warning, err := api.QueryRange(ctx, query, v1.Range{Start: endAt.Add(-width), End: endAt, Step: step})
	if err != nil {
		return nil, fmt.Errorf("error get query range: %w", err)
	}

	if len(warning) > 0 {
		for _, w := range warning {
			slog.Warn("Warning QueryRange", slog.String("warning", w))
		}
	}

	if result.Type() != model.ValMatrix {
		return nil, fmt.Errorf("unknown result type: %v", result.Type())
	}
	matrix, ok := result.(model.Matrix)

	if !ok {
		return nil, fmt.Errorf("unknown result type: %T", result)
	}

	return matrix, nil
}
