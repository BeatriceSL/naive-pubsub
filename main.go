package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

// HTTPHandleFunc used as a return value from either publish or subscribe,
// will contain a channel to be used to pass messages between publisher and subscriber
// type HTTPHandleFunc func(w http.ResponseWriter, r *http.Request)

// Message to be passed between publish websocket and subscribe websocket
type Message struct {
	mt      int
	message []byte
}

// making use of closure here so that publish and subscribe can share a channel,
// but I can provide the correct type of function to http.HandleFunc
func publish(channel chan Message) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)
			msg := new(Message)
			msg.mt = mt
			msg.message = message
			channel <- *msg // publish message to channel
		}
	}
}

func subscribe(channel chan Message) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()
		for {
			select {
			case msg := <-channel:
				log.Printf("got %s %d from publisher", msg.message, msg.mt)
				err = c.WriteMessage(msg.mt, msg.message)
				if err != nil {
					log.Println("write:", err)
					break
				}

			}
		}

	}

}

func main() {
	flag.Parse()
	log.SetFlags(0)
	c := make(chan Message)
	Publish := publish(c)
	Subscribe := subscribe(c)
	http.HandleFunc("/publish", Publish)
	http.HandleFunc("/subscribe", Subscribe)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
