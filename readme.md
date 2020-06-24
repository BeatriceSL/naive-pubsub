# Naive PubSub with websockets

makes use of channels to send a message from the `publish` endpoint to the `subscribe` endpoint.

## quickstart

install dependencies with

```
go get github.com/gorilla/websocket
```

run the server with

```
go run main.go
```

navigate to [localhost:8080](http://localhost:8080) and click the `open` button to open a connection to the publish websocket, `subscribe` to open a connection to the subscribe websocket, and `send` to send a message to the publisher, and to receive it through the subscriber.

run tests with

```
got test -run .
```


## Improvement Suggestions / notes

an obvious improvement would be more test coverage, but I was running out of time.

there could also be rate limiting for the end points.

there is also no authentication for the routes at all, and it's possible to send messages to the subscriber websocket, which should be prevented.

As far as race conditions go, I don't expect any to come from the implementation that I have supplied. In the event that the publisher gets two messages sent to it at the same time, the channel will pick between the two pseudo-randomly. I don't know where to source that claim, but Rob Pike mentions it in this video https://www.youtube.com/watch?v=f6kdp27TYZs .

As far as go best practices, I tried to adhere to them the best I could. The only way I think that I may be able to improve is to perhaps use buffered channels so that sending and receiving messages is non blocking? I don't know the size of messages beforehand though, so that may be a non issue.
