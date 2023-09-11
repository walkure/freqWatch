package databin

import "encoding/json"

type FreqDatum struct {
	Epoch int64   `json:"t"`
	Freq  float64 `json:"f"`
}

func (d FreqDatum) ToJSON() string {
	b, err := json.Marshal(d)
	if err != nil {
		return "{}"
	}
	return string(b)
}
