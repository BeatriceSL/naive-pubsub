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

~~an obvious improvement would be more test coverage, but I was running out of time.~~

there could also be rate limiting for the end points.

there is also no authentication for the routes at all, and it's possible to send messages to the subscriber websocket, which should be prevented.

### race conditions
As far as race conditions go, I don't expect any to come from the implementation that I have supplied. In the event that the publisher gets two messages sent to it at the same time, the channel will pick between the two pseudo-randomly. I don't know where to source that claim, but Rob Pike mentions it in this video https://www.youtube.com/watch?v=f6kdp27TYZs .

another note on race conditions, [websocket connections](https://godoc.org/github.com/gorilla/websocket#hdr-Concurrency) support one concurrent reader and writer. Since each connection has only one go routine that either reads or writes I don't think that this is a problem.

there is a race condition in my tests casued by the following snippet.

```
	// this creates a rece condition, I'm not really sure how to perform this test without one
	regSubs := len(pubSubStation.subscribers)
	if regSubs != 2 {
		t.Fatalf("subscriber was not unregisterd")
	}
```

If this is commented, then `go test -race .` will pass. I also commented that there has to be a better way to do this, but I'm unaware of one. I didn't really want to spend more than a day on this in the spirit of the original requirements.


## best practices

As far as go best practices, I tried to adhere to them the best I could. The only way I think that I may be able to improve is to perhaps use buffered channels so that sending and receiving messages is non blocking? I don't know the size of messages beforehand though, so that may be a non issue.


## additonal notes

I originally started this out registering and unregistering publishers, but after writing a few tests I realized that I really don't need to worry about doing that, since I'm not broadcasting any information to the publishers. this resulted in my test running one millisecond faster (yay).

I spent most of my day Friday working through this on and off. I took a bit more time on this than I did the first implementation because there was a lot that I didn't understand in terms of the example I was following, so I wanted to take my time so I could make sure I knew what I was talking about when we meet again.

I kept the additions in the sam file, but I think they could be easily seperated into `station.go`, `publisher.go` and `subscriber.go`. I left them in the same file as a matter of personal preference, It was a bit easier to follow the logic of what I was trying to do with everything in front of me. If there was a style guide that defined something else I would follow that.


The tests are extremely braindead and could vastly be improved. This was my first time writing tests for go and I'm not sure of some of the resources available to me, and I had to spend some time with the examples of how to actually create the pubsub, I didn't really have time to create tests up to the standard that I normally would. I would like to add Unit tests, and maybe have publisher write and reads be more randomized.


I also violate DRY all over the place, somewhat intentionally. I would prefer to make some DRY violations than to make a poor abstraction too early.


I see your point about how waiting on the timer to ping the websocket connection is a bit clunky, and would like to have a better way to do that.
