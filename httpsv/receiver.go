package httpsv

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/walkure/freq_recv/databin"
)

type ReceiveCallback func(place string, datum *databin.FreqDatum)

type receiverHandler struct {
	Callback ReceiveCallback
	ShareKey string
}

func (h *receiverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	signParam := r.FormValue("sign")
	if len(signParam) != 32 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	place := r.FormValue("place")
	freqParam := r.FormValue("freq")
	hash := md5.New()
	defer hash.Reset()
	hash.Write([]byte(place))
	hash.Write([]byte(freqParam))
	hash.Write([]byte(h.ShareKey))
	sign := hex.EncodeToString(hash.Sum(nil))

	if sign != signParam {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if place == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	freq, err := strconv.ParseFloat(freqParam, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")

	datum := &databin.FreqDatum{
		Epoch: time.Now().Unix(),
		Freq:  freq,
	}

	if h.Callback != nil {
		h.Callback(place, datum)
	}
}
