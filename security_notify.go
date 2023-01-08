package main

import (
	"context"
	"net/http"
	"time"

	"github.com/chengfeiZhou/notifyTool/pkg/logger"
	"github.com/chengfeiZhou/notifyTool/server"
	"github.com/namsral/flag"
)

var (
	addr = flag.String("addr", ":8900", "address to listen for http & websocket")
)

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	addr := *addr
	logs := logger.DefaultLogger()
	logs.Info("notify service running", logger.MakeField("addr", addr))
	srv := &http.Server{
		Addr:         addr,
		Handler:      server.NewServerHandler(ctx),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	panic(srv.ListenAndServe())
}
