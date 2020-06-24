package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

type HTTPHandleFunc func(w http.ResponseWriter, r *http.Request)

type Message struct {
	mt      int
	message []byte
}

func publish(channel chan Message) HTTPHandleFunc {
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

func subscribe(channel chan Message) HTTPHandleFunc {
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
				log.Printf("got %s from publisher", msg.message)
				err = c.WriteMessage(msg.mt, msg.message)
				if err != nil {
					log.Println("write:", err)
					break
				}

			}
		}

	}

}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host)
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	c := make(chan Message)
	Publish := publish(c)
	Subscribe := subscribe(c)
	http.HandleFunc("/publish", Publish)
	http.HandleFunc("/subscribe", Subscribe)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var pubws;
    var subws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
    };

    document.getElementById("open").onclick = function(evt) {
        if (pubws) {
            return false;
        }
        pubws = new WebSocket("{{.}}/publish");
        pubws.onopen = function(evt) {
            print("OPEN PUBLISH");
        }
        pubws.onclose = function(evt) {
            print("CLOSE PUBLISH");
            ws = null;
        }
        pubws.onerror = function(evt) {
            print("PUBLISH ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("subscribe").onclick = function(evt) {
        if (subws) {
            return false;
        }
        subws = new WebSocket("{{.}}/subscribe");
        subws.onopen = function(evt) {
            print("OPEN SUBSCRIBE");
        }
        subws.onclose = function(evt) {
            print("CLOSE SUBSCRIBE");
            ws = null;
        }
        subws.onmessage = function(evt) {
            print("RESPONSE SUBSCRIBE: " + evt.data);
        }
        subws.onerror = function(evt) {
            print("SUBSCRIBE ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!pubws) {
            return false;
        }
        print("SEND PUBLISH: " + input.value);
        pubws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        pubws.close();
        subws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server,
"Send" to send a message to the server and "Close" to close the connection.
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="subscribe">Subscribe</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>

`))
