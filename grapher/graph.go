package main

import (
	"fmt"
	"math"
	"time"

	"github.com/prometheus/common/model"
	"go-hep.org/x/hep/hplot"
	"golang.org/x/exp/slog"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func plotMatrix(matrix model.Matrix, loc *time.Location, updatedAt time.Time, params GraphConfig) (*plot.Plot, error) {

	plotter.DefaultLineStyle.Width = vg.Points(0.7)
	p := plot.New()

	p.Add(plotter.NewGrid())
	p.Add(hplot.HLine(0, nil, nil))
	p.Title.Text = "Power Frequency Differences(10mins median) at " + updatedAt.Format("Monday, 2 January, 2006 3:04 PM")

	p.X.Tick.Marker = plot.TickerFunc(func(min, max float64) []plot.Tick {
		ticks := []plot.Tick{}
		step := time.Hour * 2
		tf := func(i float64) time.Time { return time.Unix(int64(i), 0).In(loc) }
		tmin := tf(min)
		tmax := tf(max)
		tmin = time.Date(tmin.Year(), tmin.Month(), tmin.Day(), tmin.Hour(), 0, 0, 0, tmin.Location())
		for i := tmin; !tmax.Before(i); i = i.Add(step) {
			ticks = append(ticks, plot.Tick{Value: float64(i.Unix()), Label: i.Format("2006-01-02\n15:04 -0700")})
		}
		return ticks
	})

	/*
		p.X.Tick.Marker = plot.TimeTicks{
			Format: "2006-01-02\n15:04 -0700",
			Time: func(t float64) time.Time {
				return time.Unix(int64(t), 0).In(loc)
			},
		}
	*/
	p.Y.Label.Text = "Hz"
	p.Y.Tick.Marker = plot.TickerFunc(func(min, max float64) []plot.Tick {
		ticks := []plot.Tick{}
		step := math.Round(((max-min)/10)*1000) / 1000
		for i := min; i <= max; i += step {
			ticks = append(ticks, plot.Tick{Value: i, Label: fmt.Sprintf("%+.4f", i)})
		}
		return ticks
	})

	for _, sample := range matrix {

		graphParams, ok := params[sample.Metric["place"]]
		if !ok {
			continue
		}

		// check color format
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Warn("invalid color format",
						slog.String("color", string(graphParams.Color)),
						slog.String("place", string(sample.Metric["place"])),
					)
					// fallback color
					graphParams.Color = HexColorCode("#000000")
				}
			}()
			graphParams.Color.RGBA()
		}()

		pts := make(plotter.XYs, len(sample.Values))

		for i, v := range sample.Values {
			pts[i].X = float64(v.Timestamp.Unix())
			pts[i].Y = float64(v.Value) - graphParams.Origin
		}

		line, _, err := plotter.NewLinePoints(pts)
		if err != nil {
			return nil, fmt.Errorf("error creating line: %w", err)
		}
		line.Color = graphParams.Color
		//line.LineStyle.Width = vg.Points(0.6)
		p.Add(line)
		p.Legend.Add(graphParams.Legend, line)
	}

	return p, nil
}
