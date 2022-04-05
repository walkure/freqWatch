package httpsv

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/walkure/freq_recv/databin"
)

func Test_notificator_handlerData(t *testing.T) {

	h := NewNotificationHandler()

	hsv := httptest.NewServer(h)
	defer hsv.Close()

	epochWants := int64(2)
	freqWants := 810.514
	placeWants := "testplace"
	clientsWants := 1

	wsc1 := connectWSClient(t, hsv, "/")
	defer wsc1.Close()

	h.Notify(placeWants, &databin.FreqDatum{Epoch: epochWants, Freq: freqWants})
	data := recvWSNotifyData(t, wsc1)
	if data.Clients != clientsWants || data.Epoch != epochWants || data.Freq != freqWants || data.Place != placeWants {
		t.Errorf("invalid datum: Client[want %d,got:%d],Epoch[want %d,got:%d] Freq[want %f,got:%f] Place[wants %s,got:%s]",
			clientsWants, data.Clients, epochWants, data.Epoch, freqWants, data.Freq, placeWants, data.Place)
	}

	wsc2 := connectWSClient(t, hsv, "/")
	defer wsc2.Close()

	epochWants = int64(3)
	freqWants = 514.810
	placeWants = "test----place"
	clientsWants = 2

	h.Notify(placeWants, &databin.FreqDatum{Epoch: epochWants, Freq: freqWants})
	data = recvWSNotifyData(t, wsc2)
	if data.Clients != clientsWants || data.Epoch != epochWants || data.Freq != freqWants || data.Place != placeWants {
		t.Errorf("invalid datum: Client[want %d,got:%d],Epoch[want %d,got:%d] Freq[want %f,got:%f] Place[wants %s,got:%s]",
			clientsWants, data.Clients, epochWants, data.Epoch, freqWants, data.Freq, placeWants, data.Place)
	}
}

func Test_notificator_handlerJsonPing(t *testing.T) {

	h := NewNotificationHandler()

	hsv := httptest.NewServer(h)
	defer hsv.Close()

	wsc := connectWSClient(t, hsv, "/")
	defer wsc.Close()

	if err := wsc.WriteMessage(websocket.TextMessage, []byte("\"ping\"")); err != nil {
		t.Fatalf("WSWriteMessage:%v", err)
	}

	got := recvWSMessage(t, wsc)
	wants := "\"pong\""

	if got != wants {
		t.Errorf("wants [%s] got [%s]\n", wants, got)
	}
}

func connectWSClient(t *testing.T, h *httptest.Server, path string) *websocket.Conn {
	t.Helper()

	r := strings.NewReplacer("http://", "ws://", "http://", "wss://")
	wsUri := r.Replace(h.URL)

	// connect WebSocket and returns client
	wsc, _, err := websocket.DefaultDialer.Dial(wsUri+path, nil)
	if err != nil {
		t.Fatalf("WSClent:%v", err)
	}

	return wsc
}

func recvWSMessage(t *testing.T, ws *websocket.Conn) string {
	t.Helper()

	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("WSRecv:%v", err)
	}

	return string(msg)
}

func recvWSNotifyData(t *testing.T, ws *websocket.Conn) *notifyData {
	t.Helper()

	resp := &notifyData{}
	if err := json.Unmarshal([]byte(recvWSMessage(t, ws)), &resp); err != nil {
		t.Fatalf("WSJSON:%v", err)
	}

	return resp
}
