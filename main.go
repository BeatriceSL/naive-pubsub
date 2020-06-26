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

type Publisher struct {
	connection *websocket.Conn
	broadcast  chan Message   // send to a broadcast to be picked up by a reciever
	station    *PubSubStation // used to send registration and unregistration
}

type Subscriber struct {
	connection *websocket.Conn
	reciever   chan Message   //recieves message from broadcaster
	station    *PubSubStation // used to send registration and unregistration
}

type PubSubStation struct {
	broadcast chan Message

	publishers          map[Publisher]bool
	registerPublisher   chan Publisher
	unregisterPublisher chan Publisher

	subscribers          map[Subscriber]bool
	registerSubscriber   chan Subscriber
	unregisterSubscriber chan Subscriber
}

func (PSS *PubSubStation) run() {
	for {
		select {
		case subscriber := <-PSS.registerSubscriber:
			PSS.subscribers[subscriber] = true // true is kind of unimportant, just want a subscriber
		case subscriber := <-PSS.unregisterSubscriber:
			if _, ok := PSS.subscribers[subscriber]; ok {
				delete(PSS.subscribers, subscriber)
			}

		case publisher := <-PSS.registerPublisher:
			PSS.publishers[publisher] = true // true is kind of unimportant, just want a publisher
		case publisher := <-PSS.unregisterPublisher:
			if _, ok := PSS.publishers[publisher]; ok {
				delete(PSS.publishers, publisher)

			}
		case message := <-PSS.broadcast:
			for subscriber := range PSS.subscribers {
				select {
				case subscriber.reciever <- message:
				default:
					close(subscriber.reciever)
					delete(PSS.subscribers, subscriber) //?
				}

			}
		}
	}
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
