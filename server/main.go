package main

import (
	"log"
	"net/http"
	service "notifyTool/http"
	"notifyTool/utils/server"

	"github.com/namsral/flag"
)

var (
	addr  = flag.String("addr", ":8900", "address to listen for http & websocket")
	debug = flag.Bool("debug", false, "open debug logging")
)

func main() {
	flag.Parse()
	lg, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create zap logger: %v", err)
	}
	msgChan := make(chan *service.Msg, 10)
	sh := service.CreateServerHandler(lg, msgChan)
	panic(http.ListenAndServe(*addr, sh))
}
