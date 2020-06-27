package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Integration test for pubsub using a single publisher and a single subscriber
func Test_publishSubscribeIntegration(t *testing.T) {
	pubSubStation := stationFactory()
	Publish := publish(pubSubStation)
	Subscribe := subscribe(pubSubStation)
	go pubSubStation.run()
	mux := http.NewServeMux()

	mux.HandleFunc("/publish/", Publish)
	mux.HandleFunc("/subscribe/", Subscribe)

	s := httptest.NewServer(mux)
	defer s.Close()

	// all caps url because vscode whines if i don't
	subURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/subscribe/"
	pubURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/publish/"

	// Connect to the server
	subWS, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	pubWS, _, err := websocket.DefaultDialer.Dial(pubURL, nil)
	defer pubWS.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {

		pubWS.WriteMessage(1, []byte("hello"))
		_, p, err := subWS.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p) != "hello" {
			t.Fatalf("bad message")
		}
	}
}

// Integration test for pubsub using a single publisher andd multiple subscribers
func Test_publishSubscribeMultipleSubscribersIntegration(t *testing.T) {
	pubSubStation := stationFactory()
	Publish := publish(pubSubStation)
	Subscribe := subscribe(pubSubStation)
	go pubSubStation.run()

	mux := http.NewServeMux()

	mux.HandleFunc("/publish/", Publish)
	mux.HandleFunc("/subscribe/", Subscribe)

	s := httptest.NewServer(mux)
	defer s.Close()

	// all caps url because vscode whines if i don't
	subURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/subscribe/"
	pubURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/publish/"

	// Connect to the server
	subWS0, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS0.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	subWS1, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS1.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	subWS2, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS2.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	pubWS, _, err := websocket.DefaultDialer.Dial(pubURL, nil)
	defer pubWS.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {

		pubWS.WriteMessage(1, []byte("hello"))

		_, p0, err := subWS0.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p0) != "hello" {
			t.Fatalf("bad message")
		}

		_, p1, err := subWS1.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p1) != "hello" {
			t.Fatalf("bad message")
		}
		_, p2, err := subWS2.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p2) != "hello" {
			t.Fatalf("bad message")
		}
	}
}

// Integration test for pubsub using a single publisher andd multiple subscribers
func Test_multiplePublishersMultipleSubscribersIntegration(t *testing.T) {
	pubSubStation := stationFactory()
	Publish := publish(pubSubStation)
	Subscribe := subscribe(pubSubStation)
	go pubSubStation.run()

	mux := http.NewServeMux()

	mux.HandleFunc("/publish/", Publish)
	mux.HandleFunc("/subscribe/", Subscribe)

	s := httptest.NewServer(mux)
	defer s.Close()

	// all caps url because vscode whines if i don't
	subURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/subscribe/"
	pubURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/publish/"

	// Connect to the server
	subWS0, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS0.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	subWS1, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS1.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	subWS2, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS2.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	pubWS0, _, err := websocket.DefaultDialer.Dial(pubURL, nil)
	defer pubWS0.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	pubWS1, _, err := websocket.DefaultDialer.Dial(pubURL, nil)
	defer pubWS1.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {

		// ossicilate websockets publishing
		if i%2 == 0 {

			pubWS0.WriteMessage(1, []byte("hello"))
		} else {

			pubWS1.WriteMessage(1, []byte("hello"))
		}

		_, p0, err := subWS0.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p0) != "hello" {
			t.Fatalf("bad message")
		}

		_, p1, err := subWS1.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p1) != "hello" {
			t.Fatalf("bad message")
		}
		_, p2, err := subWS2.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p2) != "hello" {
			t.Fatalf("bad message")
		}
	}
}

func Test_disconnectPublisherDoesntBlock(t *testing.T) {
	pubSubStation := stationFactory()
	Publish := publish(pubSubStation)
	Subscribe := subscribe(pubSubStation)
	go pubSubStation.run()

	mux := http.NewServeMux()

	mux.HandleFunc("/publish/", Publish)
	mux.HandleFunc("/subscribe/", Subscribe)

	s := httptest.NewServer(mux)
	defer s.Close()

	// all caps url because vscode whines if i don't
	subURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/subscribe/"
	pubURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/publish/"

	// Connect to the server
	subWS0, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS0.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	subWS1, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS1.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	subWS2, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS2.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	pubWS0, _, err := websocket.DefaultDialer.Dial(pubURL, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	pubWS1, _, err := websocket.DefaultDialer.Dial(pubURL, nil)
	defer pubWS1.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	pubWS0.WriteMessage(1, []byte("hello"))
	_, p0, err := subWS0.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p0) != "hello" {
		t.Fatalf("bad message")
	}

	_, p1, err := subWS1.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p1) != "hello" {
		t.Fatalf("bad message")
	}
	_, p2, err := subWS2.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p2) != "hello" {
		t.Fatalf("bad message")
	}

	pubWS0.Close()

	pubWS1.WriteMessage(1, []byte("hello"))
	_, p0, err = subWS0.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p0) != "hello" {
		t.Fatalf("bad message")
	}

	_, p1, err = subWS1.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p1) != "hello" {
		t.Fatalf("bad message")
	}
	_, p2, err = subWS2.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p2) != "hello" {
		t.Fatalf("bad message")
	}

}

func Test_disconnectSubscriberDoesntBlock(t *testing.T) {
	pubSubStation := stationFactory()
	Publish := publish(pubSubStation)
	Subscribe := subscribe(pubSubStation)
	go pubSubStation.run()

	mux := http.NewServeMux()

	mux.HandleFunc("/publish/", Publish)
	mux.HandleFunc("/subscribe/", Subscribe)

	s := httptest.NewServer(mux)
	defer s.Close()

	// all caps url because vscode whines if i don't
	subURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/subscribe/"
	pubURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/publish/"

	// Connect to the server
	subWS0, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS0.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	subWS1, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	defer subWS1.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	subWS2, _, err := websocket.DefaultDialer.Dial(subURL, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	pubWS0, _, err := websocket.DefaultDialer.Dial(pubURL, nil)
	defer pubWS0.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	pubWS1, _, err := websocket.DefaultDialer.Dial(pubURL, nil)
	defer pubWS1.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	pubWS0.WriteMessage(1, []byte("hello"))
	_, p0, err := subWS0.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p0) != "hello" {
		t.Fatalf("bad message")
	}

	_, p1, err := subWS1.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p1) != "hello" {
		t.Fatalf("bad message")
	}
	_, p2, err := subWS2.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p2) != "hello" {
		t.Fatalf("bad message")
	}

	subWS2.Close()
	time.Sleep(11 * time.Second) // has to be a better way to do this
	// this creates a rece condition, I'm not really sure how to perform this test without one
	regSubs := len(pubSubStation.subscribers)
	if regSubs != 2 {
		t.Fatalf("subscriber was not unregisterd")
	}

	pubWS1.WriteMessage(1, []byte("hello"))
	_, p0, err = subWS0.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p0) != "hello" {
		t.Fatalf("bad message")
	}

	_, p1, err = subWS1.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(p1) != "hello" {
		t.Fatalf("bad message")
	}

}
