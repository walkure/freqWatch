package httpsv

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/walkure/freq_recv/databin"
)

type notificatorHandler struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]chan *databin.FreqDatum
	places  map[chan *databin.FreqDatum]string
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 1 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 2 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

func NewNotificationHandler() *notificatorHandler {
	return &notificatorHandler{
		clients: make(map[*websocket.Conn]chan *databin.FreqDatum),
		places:  make(map[chan *databin.FreqDatum]string),
	}
}

func (h *notificatorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	place := r.FormValue("place")

	if databin.LookupRingBuffer(place) == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.notification(place, w, r)

}

type notifyData struct {
	databin.FreqDatum
	Clients int `json:"c"`
}

func (h *notificatorHandler) notification(place string, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print(err)
		// http error wrote in Upgrade()
		return
	}
	defer ws.Close()

	pinger := time.NewTicker(pingPeriod)
	defer pinger.Stop()

	updateMessage := make(chan *databin.FreqDatum, 1)
	func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.clients[ws] = updateMessage
		h.places[updateMessage] = place
	}()

	// receiver goroutine waiting for close websocket.
	go func() {
		defer ws.Close()

		ws.SetReadLimit(3)
		ws.SetReadDeadline(time.Now().Add(pongWait))
		ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

		// waiting for something receive or close socket.
		ws.ReadMessage()
	}()

	// writer loop
	for loop := true; loop; {
		select {
		case <-pinger.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Println("ping:", err)
			}
		case msg, ok := <-updateMessage:
			if ok {
				data := notifyData{
					FreqDatum: databin.FreqDatum{
						Epoch: msg.Epoch,
						Freq:  msg.Freq,
					},
					Clients: len(h.clients),
				}

				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteJSON(data); err != nil {
					ws.Close()
					loop = false
				}
			} else {
				ws.Close()
				loop = false
			}
		}
	}
	func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.clients, ws)
		delete(h.places, updateMessage)
		close(updateMessage)
	}()
}

func (h *notificatorHandler) Notify(place string, datum *databin.FreqDatum) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, ch := range h.clients {
		if h.places[ch] == place {
			ch <- datum
		}
	}
}
