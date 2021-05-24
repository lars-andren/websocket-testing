package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

var addr2 = flag.String("addr2", "staging.internet.com:8080", "http service address")

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for i := 0; i < 1; i++ {
		time.Sleep(200 * time.Millisecond)
		path := fmt.Sprintf("/websockets/json/PREFIX%v", i)

		go func(){
			var reqMap = map[string]time.Time{}
			flag.Parse()
			log.SetFlags(0)

			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)

			u := url.URL{Scheme: "ws", Host: *addr2, Path: path}

			c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Fatal("dial:", err)
			}
			defer c.Close()

			done := make(chan struct{})

			writeChan := make(chan *[]byte)

			go func() {
				defer close(done)
				for {
					_, message, err := c.ReadMessage()
					if err != nil {
						log.Printf("err: %v", err)
						return
					}
					msg := string(message)
					split := strings.Split(msg, ",")
					id := strings.TrimSpace(split[1])
					id = strings.Trim(id,"\"")
					prevTime := reqMap[id]
					delete(reqMap, id)
					fmt.Printf("%v\n", time.Since(prevTime).Milliseconds())
					if strings.Contains(msg, "ping") {
						response := []byte(strings.ReplaceAll(msg, "ping", "pong"))
						writeChan <- &response
					}
				}
			}()

			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case data := <- writeChan:
					err := c.WriteMessage(websocket.TextMessage, *data)
					if err != nil {
						log.Println("write:", err)
						return
					}
				case <-done:
					return
				case _ = <-ticker.C:
					if len(reqMap) == 0 {
						randId := fmt.Sprintf("%d", rand.Int63n(math.MaxInt64))
						request := []byte(fmt.Sprintf("[2,  \"%v\", \"beat\", {}]", randId))
						//log.Println(string(request))
						reqMap[randId] = time.Now()
						err := c.WriteMessage(websocket.TextMessage,request)
						if err != nil {
							//log.Println("write:", err)
							return
						}
					}
				case <-interrupt:

					// Cleanly close the connection by sending a close message and then
					// waiting (with timeout) for the server to close the connection.
					err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					if err != nil {
						log.Println("write close:", err)
						return
					}
					select {
					case <-done:
					case <-time.After(time.Second):
					}
					return
				}
			}
		}()
	}

	for{
		select {
		case <-interrupt:
			select {
			case <-time.After(time.Second):
			}
			return
		}
	}
}
