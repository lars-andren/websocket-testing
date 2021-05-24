package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "prod2.cloudcharge.se", "http service address")

func main() {

	ticker := time.NewTicker(10 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return

			case t := <-ticker.C:
				go func() { doTest() }()
				log.Printf("opening websocket %s", t)
			}
		}
	}()

	time.Sleep(60000 * time.Millisecond)
	ticker.Stop()
	done <- true
	fmt.Println("Ticker stopped")
}

func doTest() {
	conn := openWebsocket()

	request := createNotificationRequest()
	writeToSocket(conn, &request)
}

func openWebsocket() *websocket.Conn {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/sockets/json/" + uuid}
	log.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	return conn
}

func writeToSocket(conn *websocket.Conn, noteReq *NotificationRequest) {

	serialized := SerializeCall(noteReq)

	err := conn.WriteMessage(websocket.TextMessage, *serialized)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func SerializeCall(data interface{}) *[]byte {
	list := []interface{}{2, data}
	bytes, _ := json.Marshal(list)
	return &bytes
}

func createNotificationRequest() NotificationRequest {

	vendor := CiString20Type("Acme")
	model := CiString20Type("ABC123")
	serial := CiString25Type("1")
	imsi := CiString20Type("imsi")

	noteRequest := NotificationRequest{
		Vendor:       &vendor,
		Model:        &model,
		SerialNumber: &serial,
		Imsi:         &imsi,
	}

	return noteRequest
}

type CiString20Type string
type CiString25Type string

type NotificationRequest struct {
	Vendor *CiString20Type `xml:"vendor" json:"vendor"`

	Model *CiString20Type `xml:"model" json:"model"`

	SerialNumber *CiString25Type `xml:"serialNumber,omitempty" json:"serialNumber,omitempty"`

	Imsi *CiString20Type `xml:"imsi,omitempty" json:"imsi,omitempty"`
}

