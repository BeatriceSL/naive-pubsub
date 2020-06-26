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
	station    *PubSubStation // used to send registration and unregistration
}

type Subscriber struct {
	connection *websocket.Conn
	reciever   chan Message   //recieves message from broadcaster
	station    *PubSubStation // used to send registration and unregistration
}

type PubSubStation struct {
	broadcast chan Message

	publishers          map[*Publisher]bool
	registerPublisher   chan *Publisher
	unregisterPublisher chan *Publisher

	subscribers          map[*Subscriber]bool
	registerSubscriber   chan *Subscriber
	unregisterSubscriber chan *Subscriber
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
				// launch a go routine for every subscriber, instead of of just looping through
				select {
				case subscriber.reciever <- message:
				}

			}
		}
	}
}

func stationFactory() *PubSubStation {
	return &PubSubStation{
		broadcast: make(chan Message),

		publishers:          make(map[*Publisher]bool),
		registerPublisher:   make(chan *Publisher),
		unregisterPublisher: make(chan *Publisher),

		subscribers:          make(map[*Subscriber]bool),
		registerSubscriber:   make(chan *Subscriber),
		unregisterSubscriber: make(chan *Subscriber),
	}
}

// making use of closure here so that publish and subscribe can share a channel,
// but I can provide the correct type of function to http.HandleFunc
func publish(station *PubSubStation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		publisher := &Publisher{
			connection: c,
			station:    station,
		}
		publisher.station.registerPublisher <- publisher

		// TODO prob should be a named function but for protoyping leaving here for now
		go func() {
			defer func() {
				publisher.station.unregisterPublisher <- publisher
				publisher.connection.Close()
			}()
			for {
				mt, message, err := publisher.connection.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("error: %v", err)
					}
					break
				}
				msg := &Message{
					mt:      mt,
					message: message,
				}
				publisher.station.broadcast <- *msg
			}
		}()
	}
}

func subscribe(station *PubSubStation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		subscriber := &Subscriber{
			connection: c,
			station:    station,
			reciever:   make(chan Message, 1),
		}

		subscriber.station.registerSubscriber <- subscriber
		go func() {
			defer func() {
				subscriber.station.unregisterSubscriber <- subscriber
				subscriber.connection.Close()

			}()
			for {
				select {
				case message, ok := <-subscriber.reciever:
					if !ok {
						// The hub closed the channel.
						subscriber.connection.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}
					subscriber.connection.WriteMessage(message.mt, message.message)

				}
			}
		}()

	}

}

func main() {
	flag.Parse()
	log.SetFlags(0)
	pubSubStation := stationFactory()
	Publish := publish(pubSubStation)
	Subscribe := subscribe(pubSubStation)
	http.HandleFunc("/publish", Publish)
	http.HandleFunc("/subscribe", Subscribe)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
