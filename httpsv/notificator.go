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
	mu           sync.RWMutex
	clients      map[*websocket.Conn]chan *databin.FreqDatum
	places       map[chan *databin.FreqDatum]string
	bufferGetter func(place string) *databin.DataRingBuffer
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
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

func NewNotificationHandler(db *databin.DataBin) *notificatorHandler {
	return &notificatorHandler{
		clients:      make(map[*websocket.Conn]chan *databin.FreqDatum),
		places:       make(map[chan *databin.FreqDatum]string),
		bufferGetter: db.LookupRingBuffer,
	}
}

func (h *notificatorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	place := r.FormValue("place")

	if h.bufferGetter(place) == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print(err)
		// http error wrote in Upgrade()
		return
	}
	h.wsNotification(place, ws)

}

type notifyData struct {
	databin.FreqDatum
	Clients int `json:"c"`
}

func (h *notificatorHandler) wsNotification(place string, ws *websocket.Conn) {

	updateMessage := make(chan *databin.FreqDatum, 1)
	func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.clients[ws] = updateMessage
		h.places[updateMessage] = place
	}()

	pingResult := make(chan string, 1)

	go wsReceiver(ws, pingResult)
	go h.wsTransmitter(ws, updateMessage, pingResult)
}

func (h *notificatorHandler) wsTransmitter(ws *websocket.Conn, updateMessage chan *databin.FreqDatum, pingResult chan string) {
	defer ws.Close()

	// writer loop
	pinger := time.NewTicker(pingPeriod)
	defer pinger.Stop()
	for loop := true; loop; {
		select {
		case <-pinger.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(writeWait)); err != nil {
				ws.Close()
				loop = false
			}
		case updateMsg, ok := <-updateMessage:
			if ok {
				data := notifyData{
					FreqDatum: databin.FreqDatum{
						Epoch: updateMsg.Epoch,
						Freq:  updateMsg.Freq,
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
		case pongMsg, ok := <-pingResult:
			if ok {
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteMessage(websocket.TextMessage, []byte(pongMsg)); err != nil {
					ws.Close()
					loop = false
				}
			}
		}
	}

	func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.clients, ws)
		delete(h.places, updateMessage)

		//should close channel after removes from  clients list.
		close(updateMessage)
	}()
}

// receiver goroutine waiting for control/ping message or close websocket.
func wsReceiver(ws *websocket.Conn, pingResult chan string) {
	defer ws.Close()
	defer close(pingResult)
	ws.SetReadLimit(6) // maximum ping message size
	ws.SetReadDeadline(time.Now().Add(pongWait))

	// handle pong(send ping from us).
	ws.SetPongHandler(func(msg string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// process any messages(process control message completes internally. )
		_, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		if string(message) == "\"ping\"" {
			pingResult <- "\"pong\""
		}
	}
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
