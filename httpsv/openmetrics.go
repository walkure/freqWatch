package httpsv

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/walkure/freq_recv/databin"
)

type openmetriucsHandler struct {
	mu   sync.RWMutex
	data map[string]*databin.FreqDatum
}

func NewOpenMetricsHandler() *openmetriucsHandler {
	return &openmetriucsHandler{
		data: make(map[string]*databin.FreqDatum),
	}
}

func (h *openmetriucsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.data) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprintln(w, "# HELP power_freq The frequency of power line.\n# TYPE power_freq gauge")
	for place, it := range h.data {
		fmt.Fprintf(w, "power_freq{place=\"%s\"} %f %d\n", place, it.Freq, it.Epoch*1000)
	}
}

func (h *openmetriucsHandler) Update(place string, datum *databin.FreqDatum) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.data[place] = datum
}
