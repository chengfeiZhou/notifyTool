package main

import (
	"github.com/namsral/flag"
)

var (
	httpPort = flag.Int("httpPort", 8900, "address to listen for http")
	wsPort   = flag.Int("wsPort", 8901, "address to listen for websocket")
)

func main() {
	flag.Parse()

}
