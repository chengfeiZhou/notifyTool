package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/chengfeiZhou/notifyTool/pkg/logger"
	"github.com/chengfeiZhou/notifyTool/pkg/utils"
	"github.com/gorilla/websocket"
)

const (
	channelSplitChar = "|"
)

type wsHub struct {
	ws    *websocket.Conn
	lock  sync.Locker
	chans *utils.StringSet // [ch1,ch2,ch3]
}

type Command struct {
	CommandType string `json:"command_type"`
	Message     string `json:"message"`
}

type Msg struct {
	Channel  string
	Channels []string
	Content  []byte // json content
}

// nolint
type Response struct {
	Channel string      `json:"channel"`
	Content interface{} `json:"content"`
}

// nolint
type ServerHandler struct {
	upgrader *websocket.Upgrader
	logger   logger.Logger
	wsHubs   []*wsHub
	msgChan  chan *Msg
}

type OptionFunc func(*ServerHandler)

func WithLogger(lg logger.Logger) OptionFunc {
	return func(sh *ServerHandler) {
		sh.logger = lg
	}
}

func NewServerHandler(ctx context.Context, ops ...OptionFunc) *ServerHandler {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	sh := &ServerHandler{
		logger:   logger.DefaultLogger(),
		upgrader: upgrader,
		wsHubs:   make([]*wsHub, 0, 20),
		msgChan:  make(chan *Msg, 10),
	}

	for _, op := range ops {
		op(sh)
	}
	go sh.monitorMsgChan(ctx)
	return sh
}

func (sh *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// DOC: https://zh.m.wikipedia.org/zh-cn/WebSocket
	if upgrade := r.Header.Get("Upgrade"); upgrade == "websocket" {
		sh.WSHandler(w, r)
		return
	}
	sh.HTTPHandler(w, r)
}

func (sh *ServerHandler) WSHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := sh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		sh.logger.Error(logger.ErrorHandleError, "[ws]websocket connect error", logger.ErrorField(err))
		return
	}
	// save connection
	key := r.URL.Path[1:]
	chans := strings.Split(key, channelSplitChar)
	sh.logger.Info("[ws]open websocket for key", logger.MakeField("chans", chans))
	wh := &wsHub{
		ws:    ws,
		lock:  &sync.Mutex{},
		chans: utils.NewStringSet(chans...),
	}
	sh.wsHubs = append(sh.wsHubs, wh)

	for {
		_, ms, err := ws.ReadMessage()
		if err != nil {
			sh.logger.Error(logger.ErrorEmptyData, "[ws] read message", logger.MakeField("remote", wh.ws.RemoteAddr().String()),
				logger.ErrorField(err))
			break
		}
		command := new(Command)
		if err := json.Unmarshal(ms, command); err != nil {
			sh.logger.Error(logger.ErrorMethodNotSupport, "[ws] read message", logger.ErrorField(err))
			continue
		}
		switch command.CommandType {
		case "update_channel":
			newKey := command.Message
			wh.lock.Lock()
			newChans := strings.Split(newKey, channelSplitChar)
			sh.logger.Info("update channel", logger.MakeField("new chans", newChans))
			wh.chans = utils.NewStringSet(newChans...)
			wh.lock.Unlock()
		default:
			sh.logger.Error(logger.ErrorMethodNotSupport, "No commands defined", logger.MakeField("command type", command.CommandType))
		}
	}
}

func (sh *ServerHandler) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		sh.logger.Error(logger.ErrorEmptyData, "[http]read body", logger.ErrorField(err))
		fmt.Fprintf(w, "error")
		return
	}
	key := r.URL.Path[1:]
	sh.msgChan <- &Msg{
		Channel:  key,
		Channels: strings.Split(key, channelSplitChar),
		Content:  content,
	}
	fmt.Fprintf(w, "ok")
}

func (sh *ServerHandler) monitorMsgChan(ctx context.Context) {
	for {
		select {
		case msg, ok := <-sh.msgChan:
			if !ok {
				continue
			}
			sh.logger.Info("[http] collect msg from http",
				logger.MakeField("channels ", msg.Channels))
			for _, channel := range msg.Channels {
				sh.sendMsgToChannel(channel, msg.Content)
			}
		case <-ctx.Done():
			sh.logger.Info("notify monitor is close")
			return
		}
	}
}

func (sh *ServerHandler) sendMsgToChannel(channel string, message []byte) {
	for _, hub := range sh.wsHubs {
		if !hub.chans.Has(channel) {
			continue
		}
		var content interface{}
		_ = json.Unmarshal(message, &content)
		resp := &Response{
			Channel: channel,
			Content: content,
		}
		go sh.sendMsg(hub, resp)
	}
}

func (sh *ServerHandler) sendMsg(wh *wsHub, msg *Response) {
	wh.lock.Lock()
	defer wh.lock.Unlock()
	if err := wh.ws.WriteJSON(msg); err != nil {
		sh.logger.Error(logger.ErrorHandleError, "failed sending to client of channel",
			logger.ErrorField(err), logger.MakeField("channel", msg.Channel))
	}

	sh.logger.Info("[ws]send notify", logger.MakeField("remote", wh.ws.RemoteAddr().String()),
		logger.MakeField("channel", msg.Channel), logger.MakeField("data", msg))
}
