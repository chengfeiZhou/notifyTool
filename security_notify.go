package main

import (
	"context"
	"log"
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
	logs, err := logger.NewLogger()
	if err != nil {
		log.Fatal("创建日志实例错误")
	}
	logs.Info("notify service running", logger.MakeField("addr", addr))
	srv := &http.Server{
		Addr:         addr,
		Handler:      server.NewServerHandler(ctx, server.WithLogger(logs)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	panic(srv.ListenAndServe())
}
