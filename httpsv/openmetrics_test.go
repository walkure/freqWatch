package httpsv

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/walkure/freqWatch/databin"
)

func Test_openmetrics_handlerNil(t *testing.T) {
	h := NewOpenMetricsHandler(databin.NewDataBin(10))

	req := httptest.NewRequest("GET", "/metrics", nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Errorf("unexpected status code:%d\n", res.Code)
	}

	expected := ""

	if got := res.Body.String(); got != expected {
		t.Errorf("want [%s], but [%s]", expected, got)
	}
}

func Test_openmetrics_handlerSingle(t *testing.T) {
	h := NewOpenMetricsHandler(databin.NewDataBin(10))

	req := httptest.NewRequest("GET", "/metrics", nil)
	res := httptest.NewRecorder()

	h.Update("test1", &databin.FreqDatum{Freq: 60.012, Epoch: 1})
	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("unexpected status code:%d\n", res.Code)
	}

	expected := `# HELP power_freq The frequency of power line.
# TYPE power_freq gauge
power_freq{place="test1"} 60.012000 1000
`
	if got := res.Body.String(); got != expected {
		t.Errorf("want %s, but %s", expected, got)
	}
}

func Test_openmetrics_handlerUpdate(t *testing.T) {
	h := NewOpenMetricsHandler(databin.NewDataBin(10))

	req := httptest.NewRequest("GET", "/metrics", nil)
	res := httptest.NewRecorder()

	h.Update("test1", &databin.FreqDatum{Freq: 60.012, Epoch: 1})
	h.Update("test1", &databin.FreqDatum{Freq: 50.012, Epoch: 4})
	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("unexpected status code:%d\n", res.Code)
	}

	expected := `# HELP power_freq The frequency of power line.
# TYPE power_freq gauge
power_freq{place="test1"} 50.012000 4000
power_freq{place="test1"} 60.012000 1000
`
	if got := res.Body.String(); got != expected {
		t.Errorf("want %s, but %s", expected, got)
	}
}

func Test_openmetrics_handlerMultiple(t *testing.T) {
	h := NewOpenMetricsHandler(databin.NewDataBin(10))

	req := httptest.NewRequest("GET", "/metrics", nil)
	res := httptest.NewRecorder()

	h.Update("test1", &databin.FreqDatum{Freq: 60.0120, Epoch: 2})
	h.Update("test2", &databin.FreqDatum{Freq: 50.0120, Epoch: 4})
	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("unexpected status code:%d\n", res.Code)
	}

	// range returns random order
	expected1 := `# HELP power_freq The frequency of power line.
# TYPE power_freq gauge
power_freq{place="test1"} 60.012000 2000
power_freq{place="test2"} 50.012000 4000
`
	expected2 := `# HELP power_freq The frequency of power line.
# TYPE power_freq gauge
power_freq{place="test2"} 50.012000 4000
power_freq{place="test1"} 60.012000 2000
`
	got := res.Body.String()
	if got != expected1 && got != expected2 {
		t.Errorf("want %s or %s, but %s", expected1, expected2, got)
	}
}

func Test_openmetrics_handlerAverage(t *testing.T) {
	h := NewOpenMetricsHandler(databin.NewDataBin(10))

	req := httptest.NewRequest("GET", "/metrics?mode=average", nil)
	res := httptest.NewRecorder()

	h.Update("test1", &databin.FreqDatum{Freq: 60.0120, Epoch: 2})
	h.Update("test1", &databin.FreqDatum{Freq: 50.0120, Epoch: 4})
	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("unexpected status code:%d\n", res.Code)
	}

	expected := `# HELP power_freq The frequency of power line.
# TYPE power_freq gauge
power_freq{place="test1"} 55.0120
`
	got := res.Body.String()
	if got != expected {
		t.Errorf("want [%s] , but [%s]", expected, got)
	}
}
