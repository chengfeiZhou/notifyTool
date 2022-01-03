package service

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

const (
	CHANNEL_SPLIT_CHAR = "|"
	QUERY_KEY          = "cid"
)

type WsServer interface {
	AddConn(ctx context.Context, cids []string, w http.ResponseWriter, r *http.Request) (WsCli, error)
	DelConn(ctx context.Context, cli WsCli) error
}

type WsCli interface {
	SendData(ctx context.Context, content io.Reader) error
	Listen(ctx context.Context, cb func(context.Context, Command) error)
	UpdateChannel(ctx context.Context, cids []string) error
}

// Command defines notify struct
type Command struct {
	CommandType string `json:"command_type"`
	Message     string `json:"message"`
}

// ServerHandler 定义http服务处理实例
type ServerHandler struct {
	Logger *zap.Logger
	WsSrv  WsServer
}

// ServeHTTP 实现Handler的ServeHTTP
// type Handler interface {
// 	ServeHTTP(ResponseWriter, *Request)
// }
func (h *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 负载到websocket handle
	for _, v := range r.Header.Values("Upgrade") {
		if v == "websocket" {
			h.WSHandler(w, r)
			return
		}
	}
	// 负载到http handle
	h.HTTPHandler(w, r)
}

// WSHandler websocket处理函数
// 优先获取路径参数, host/{cid1}|{cid2}
// 获取query, cid={cid1}&cid={cid1}
func (h *ServerHandler) WSHandler(w http.ResponseWriter, r *http.Request) {
	cids := getChannels(r)
	if len(cids) == 0 {
		h.Logger.Sugar().Error("websocket connect error: channel's length is 0")
		// TODO: 返回失败
		return
	}
	// 增加一个ws连接
	ws, err := h.WsSrv.AddConn(context.Background(), cids, w, r)
	if err != nil {
		h.Logger.Sugar().Error("websocket connect error: ", zap.Error(err))
		// TODO: 返回一个错误响应
		return
	}
	fmt.Println(len(cids), cids)
	h.Logger.Sugar().Info("open websocket for key: ", zap.Strings("channels", cids))
	go ws.Listen(context.Background(), func(ctx context.Context, command Command) error {
		if command.CommandType == "update_channel" {
			channels := strings.Split(command.Message, CHANNEL_SPLIT_CHAR)
			h.Logger.Sugar().Info(zap.String("update channel", command.Message))
			return ws.UpdateChannel(ctx, channels)
		}
		return nil
	})
	fmt.Fprint(w, "ok!")
}

// HTTPHandler 接收http端的消息
func (h *ServerHandler) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	cids := getChannels(r)
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "error: %s", err)
		return
	}
	fmt.Println(cids, content)
	// h.WsHub.MsgChan <- &Msg{
	// 	channels: cids,
	// 	content:  content,
	// }
}

func getChannels(r *http.Request) (cids []string) {
	if pathStr := strings.Trim(r.URL.Path, "/"); pathStr != "" {
		cids = strings.Split(pathStr, CHANNEL_SPLIT_CHAR)
	}
	if cqs, exist := r.URL.Query()[QUERY_KEY]; exist {
		cids = append(cids, cqs...)
	}
	return cids
}
