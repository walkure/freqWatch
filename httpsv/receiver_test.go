package httpsv

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/walkure/freq_recv/databin"
)

func Test_recv_handler(t *testing.T) {
	type args struct {
		method string
		target string
		body   string
	}
	type resp struct {
		code  int
		value float64
		place string
	}
	tests := []struct {
		name string
		args args
		resp resp
	}{
		{
			"none params",
			args{http.MethodGet, "/hoge", ""},
			resp{http.StatusBadRequest, 0., ""},
		},
		{
			"insufficient",
			args{http.MethodGet, "/hoge?freq=114514&sign=646da9ae5d90e6b51b06ede01b9fed67", ""},
			resp{http.StatusBadRequest, 0., ""},
		},
		{
			"invalid signature",
			args{http.MethodGet, "/hoge?place=hoge&freq=114.514&sign=646da9ae5d90e6b51b06ede01b9fed67", ""},
			resp{http.StatusBadRequest, 0., ""},
		},
		{
			"valid signature",
			args{http.MethodGet, "/hoge?place=hoge&freq=114.514&sign=92a438ee33a8b3bf07f27c9d118f2586", ""},
			resp{http.StatusOK, 114.514, "hoge"},
		},
		{
			"invalid method",
			args{http.MethodPost, "/hoge?place=hoge&freq=114514&sign=92a438ee33a8b3bf07f27c9d118f2586", ""},
			resp{http.StatusBadRequest, 0., ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.args.method, tt.args.target, bytes.NewBufferString(tt.args.body))
			res := httptest.NewRecorder()
			receiver := &receiverHandler{
				Callback: func(place string, datum *databin.FreqDatum) {
					if datum.Freq != tt.resp.value {
						t.Errorf("want %f, but %f", tt.resp.value, datum.Freq)
					}
					if place != tt.resp.place {
						t.Errorf("want %s, but %s", tt.resp.place, place)
					}
				},
				ShareKey: "giog890dfg7098sdfgsffdvd34",
			}
			receiver.ServeHTTP(res, req)
			if res.Code != int(tt.resp.code) {
				t.Errorf("want %d, but %d", tt.resp.code, res.Code)
			}
		})
	}
}
