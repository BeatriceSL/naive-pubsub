package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func Test_subscribe(t *testing.T) {
	c := make(chan Message)
	Subscribe := subscribe(c)

	s := httptest.NewServer(http.HandlerFunc(Subscribe))
	defer s.Close()

	msg := new(Message)
	msg.mt = 1
	msg.message = []byte("hello")

	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	defer ws.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {
		c <- *msg

		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p) != "hello" {
			t.Fatalf("bad message")
		}
	}
}

// Integration test for pubsub using a single publisher and a single subscriber
func Test_publishSubscribeIntegration(t *testing.T) {
	c := make(chan Message)
	Subscribe := subscribe(c)
	Publish := publish(c)

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
