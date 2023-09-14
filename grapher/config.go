package main

import (
	"strconv"

	"github.com/prometheus/common/model"
)

// HexColorCode is a color.Color hex color code of the form "#RRGGBB".
type HexColorCode string

// RGBA implements the color.Color interface.
func (c HexColorCode) RGBA() (r, g, b, a uint32) {
	if len(c) != 7 || c[0] != '#' {
		panic("invalid color format")
	}

	parseHexColor := func(s HexColorCode) (uint32, error) {
		v, err := strconv.ParseUint(string(s), 16, 8)
		return uint32(v) | uint32(v)<<8, err
	}

	var err error
	r, err = parseHexColor(c[1:3])
	if err != nil {
		panic(err)
	}
	g, err = parseHexColor(c[3:5])
	if err != nil {
		panic(err)
	}
	b, err = parseHexColor(c[5:7])
	if err != nil {
		panic(err)
	}
	a = 0xffff
	return
}

// GraphConfigItem is a graph config item.
type GraphConfigItem struct {
	Legend string       `json:"legend"`
	Origin float64      `json:"origin"`
	Color  HexColorCode `json:"color"`
}

// GraphConfig is a graph config.
type GraphConfig map[model.LabelValue]GraphConfigItem
