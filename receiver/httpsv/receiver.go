package httpsv

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/walkure/freqWatch/databin"
)

type ReceiveCallback func(place string, datum *databin.FreqDatum)

type receiverHandler struct {
	Callback ReceiveCallback
	ShareKey string
}

const (
	ceilFreq  = 60 + 5
	floorFreq = 50 - 5
)

func (h *receiverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	const badReq = "Bad Request"

	signParam := r.FormValue("sign")
	if len(signParam) != 32 {
		http.Error(w, badReq, http.StatusBadRequest)
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
		http.Error(w, badReq, http.StatusBadRequest)
		return
	}

	if place == "" {
		http.Error(w, badReq, http.StatusBadRequest)
		return
	}

	freq, err := strconv.ParseFloat(freqParam, 64)
	if err != nil {
		http.Error(w, badReq, http.StatusBadRequest)
		return
	}

	if freq < floorFreq || freq > ceilFreq {
		http.Error(w, badReq, http.StatusBadRequest)
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
