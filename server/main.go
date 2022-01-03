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
	sh := &service.ServerHandler{
		Logger: lg,
		WsSrv:  nil,
	}
	panic(http.ListenAndServe(*addr, sh))
}
