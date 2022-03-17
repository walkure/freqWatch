package httpsv

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/walkure/freq_recv/databin"
)

func StartHttpServer() {
	shareKey := os.Getenv("SHARE_KEY")
	if shareKey == "" {
		log.Fatal("SHARE_KEY is mandatory envrions.")
	}

	// 60(seconds) * 10(minutes)
	dumpBuffer := initDumper(60*10, "DUMP_BUFFER")

	notifier := NewNotificationHandler(dumpBuffer)
	dumper := NewDumperHandler(dumpBuffer)

	openmetrics := NewOpenMetricsHandler()

	receiver := &receiverHandler{
		Callback: func(place string, datum *databin.FreqDatum) {
			dumper.Update(place, datum)
			notifier.Notify(place, datum)
			openmetrics.Update(place, datum)

			//log.Printf("place:%s freq:%f\n", place, freq)
		},
		ShareKey: shareKey,
	}

	recvPath := os.Getenv("RECV_PATH")
	if recvPath == "" {
		recvPath = "/frecv"
	} else {
		log.Printf("receive at %s\n", recvPath)
	}

	metricsPath := os.Getenv("METRICS_PATH")
	if metricsPath == "" {
		metricsPath = "/metrics"
	} else {
		log.Printf("metrics at %s\n", metricsPath)
	}

	http.Handle(recvPath, receiver)
	http.Handle(metricsPath, openmetrics)
	http.Handle("/dump", dumper)
	http.Handle("/ws", notifier)

	listener := getListener("", 8080)
	log.Printf("Start listening at %s\n", listener)
	http.ListenAndServe(listener, nil)
}

func initDumper(defaultSize int, keyName string) *databin.DataBin {
	size := defaultSize
	sizeVal := os.Getenv(keyName)
	if sizeVal != "" {
		parsedSize, err := strconv.Atoi(sizeVal)
		if err == nil {
			size = parsedSize
		}
	}

	log.Printf("DataBin(%s) size:%d\n", keyName, size)

	return databin.NewDataBin(size)
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
