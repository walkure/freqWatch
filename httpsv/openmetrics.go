package httpsv

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/walkure/freq_recv/databin"
)

type openmetriucsHandler struct {
	mu sync.RWMutex
	db *databin.DataBin
}

func NewOpenMetricsHandler(db *databin.DataBin) *openmetriucsHandler {
	return &openmetriucsHandler{
		db: db,
	}
}

func (h *openmetriucsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	places := h.db.ListRingBuffer()
	if len(places) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	mode := getMode(r)

	switch mode {
	case "all":
		flush := r.URL.Query().Has("flush")
		h.sendAllMetrics(w, flush)
	case "average":
		h.sendAverageMetrics(w)
	default:
		panic("unknown mode:" + mode)
	}
}

func (h *openmetriucsHandler) sendAverageMetrics(w http.ResponseWriter) {
	places := h.db.ListRingBuffer()
	fmt.Fprintln(w, "# HELP power_freq The frequency of power line.\n# TYPE power_freq gauge")
	for _, place := range places {
		drb := h.db.LookupRingBuffer(place)
		if drb == nil {
			continue
		}
		sum := float64(0)
		count := 0
		drb.PeekAll(true, func(it *databin.FreqDatum) bool {
			sum += it.Freq
			count++
			return false
		})
		fmt.Fprintf(w, "power_freq{place=\"%s\"} %.4f\n", place, sum/float64(count))
	}

}

func (h *openmetriucsHandler) sendAllMetrics(w http.ResponseWriter, flush bool) {
	places := h.db.ListRingBuffer()
	fmt.Fprintln(w, "# HELP power_freq The frequency of power line.\n# TYPE power_freq gauge")
	for _, place := range places {
		drb := h.db.LookupRingBuffer(place)
		if drb == nil {
			continue
		}

		drb.PeekAll(true, func(it *databin.FreqDatum) bool {
			fmt.Fprintf(w, "power_freq{place=\"%s\"} %f %d\n", place, it.Freq, it.Epoch*1000)
			return false
		})
		if flush {
			drb.Init(-1)
		}
	}

}

func getMode(r *http.Request) string {
	mode := r.FormValue("mode")

	if mode == "average" {
		return "average"
	}

	return "all"

}

func (h *openmetriucsHandler) Update(place string, datum *databin.FreqDatum) {
	h.mu.Lock()
	defer h.mu.Unlock()
	db := h.db.GetRingBuffer(place)
	db.PushBack(datum)
}
