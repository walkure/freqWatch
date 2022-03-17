package httpsv

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/walkure/freq_recv/databin"
)

type dumperHandler struct {
	db        *databin.DataBin
	mu        sync.RWMutex
	jsonCache map[string]string
}

func NewDumperHandler(db *databin.DataBin) *dumperHandler {
	return &dumperHandler{
		jsonCache: make(map[string]string),
		db:        db,
	}
}

func (h *dumperHandler) InvalidateJsonCache(place string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.jsonCache, place)
}

func (h *dumperHandler) generateJSON(place string) string {

	if jsonBody := func(place string) string {
		h.mu.RLock()
		defer h.mu.RUnlock()
		if cache, ok := h.jsonCache[place]; ok {
			return cache
		}
		return ""
	}(place); jsonBody != "" {
		return jsonBody
	}

	dbr := h.db.LookupRingBuffer(place)
	if dbr == nil {
		return ""
	}

	jsonData := make([]string, 0, 600)
	dbr.PeekAll(false, func(d *databin.FreqDatum) bool {
		jsonData = append(jsonData, d.ToJSON())
		return false
	})

	jsonBody := fmt.Sprintf("[%s]", strings.Join(jsonData, ","))

	h.mu.Lock()
	defer h.mu.Unlock()
	h.jsonCache[place] = jsonBody

	return jsonBody
}

func (h *dumperHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	place := r.FormValue("place")

	js := h.generateJSON(place)
	if js == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "%s not found.", place)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprint(len(js)))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, js)
}

func (h *dumperHandler) Update(place string, datum *databin.FreqDatum) {
	dbr := h.db.GetRingBuffer(place)
	dbr.PushBack(datum)
	h.InvalidateJsonCache(place)
}
