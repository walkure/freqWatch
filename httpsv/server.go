package httpsv

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/walkure/freq_recv/databin"
)

func StartHttpServer() {

	notifier := NewNotificationHandler()
	dumper := NewDumperHandler()

	shareKey := os.Getenv("SHARE_KEY")
	if shareKey == "" {
		log.Fatal("SHARE_KEY is mandatory envrions.")
	}

	receiver := &receiverHandler{
		Callback: func(place string, freq float32) {
			log.Printf("place:%s,freq:%f\n", place, freq)

			datum := &databin.FreqDatum{
				Epoch: time.Now().Unix(),
				Freq:  freq,
			}
			dbr := databin.GetRingBuffer(place)
			dbr.PushBack(datum)
			dumper.InvalidateJsonCache(place)
			notifier.Notify(place, datum)
		},
		ShareKey: shareKey,
	}

	recvPath := os.Getenv("RECV_PATH")
	if recvPath == "" {
		recvPath = "/frecv"
	}

	http.Handle(recvPath, receiver)
	http.Handle("/dump", dumper)
	http.Handle("/ws", notifier)

	listener := getListener("", 8080)
	log.Printf("Start listening at %s\n", listener)
	http.ListenAndServe(listener, nil)
}

func getListener(defaultListener string, defaultPort uint16) string {

	envPort := os.Getenv("PORT")
	listener := os.Getenv("LISTEN")

	if listener == "" {
		listener = defaultListener
	}
	port, _ := strconv.Atoi(envPort)
	if port > 65535 || port < 1 {
		port = int(defaultPort)
	}

	return fmt.Sprintf("%s:%d", listener, port)
}
