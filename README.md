# notifyTool
this is a message forwarding service for http to websocket

## task
* websocket server for `Security` web client


## example
* post alarm and capture content

```
post("http://xxx:8900/{cid}", content)
```

* websocket 

```
var websocket = new WebSocket("ws://xxx:8901/{cid}")
or
var websocket = new WebSocket("ws://xxx:8901/?cid={cid}")
websocket.onMessage = function( con ){
	// TODO ... 
}
```