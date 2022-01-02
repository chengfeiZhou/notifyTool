package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	CHANNEL_SPLIT_CHAR = "|"
)

type WsHub struct {
	Hub     map[*websocket.Conn][]string
	Locks   map[*websocket.Conn]*sync.Mutex
	MsgChan chan *Msg
}

// Command defines notify struct
type Command struct {
	CommandType string `json:"command_type"`
	Message     string `json:"message"`
}

// Msg defines message body
type Msg struct {
	channels []string
	content  []byte // json content
}

type Response struct {
	Channel string      `json:"channel"`
	Content interface{} `json:"content"`
}

// ServerHandler 定义http服务处理实例
type ServerHandler struct {
	Logger     *zap.Logger
	wsUpgrader *websocket.Upgrader
	WsHub      *WsHub
}

// CreateServerHandler
func CreateServerHandler(logger *zap.Logger, msgChan chan *Msg) *ServerHandler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return &ServerHandler{
		Logger:     logger,
		wsUpgrader: &upgrader,
		WsHub: &WsHub{
			Hub:     make(map[*websocket.Conn][]string, 100),
			Locks:   make(map[*websocket.Conn]*sync.Mutex, 100),
			MsgChan: msgChan,
		},
	}
}

// ServeHTTP 实现Handler的ServeHTTP
// type Handler interface {
// 	ServeHTTP(ResponseWriter, *Request)
// }
func (h *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 负载到websocket handle
	if upgrade, exist := r.Header["Upgrade"]; exist {
		for _, v := range upgrade {
			if v == "websocket" {
				h.WSHandler(w, r)
				return
			}
		}
	}
	// 负载到http handle
	h.HTTPHandler(w, r)
}

// WSHandler websocket处理函数
// 优先获取路径参数, host/{cid1}|{cid2}
// 获取query, cid={cid1}&cid={cid1}
func (h *ServerHandler) WSHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := h.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		h.Logger.Sugar().Error("websocket connect error: ", zap.Error(err))
		return
	}

	cids := getChannels(r)
	if len(cids) == 0 {
		// TODO: 返回失败
		h.Logger.Sugar().Error("websocket connect error: channel's length is 0")
		return
	}
	h.Logger.Sugar().Info("open websocket for key: ", zap.Strings("channels", cids))

	// save connection
	h.WsHub.Hub[ws] = cids
	h.WsHub.Locks[ws] = &sync.Mutex{}
	defer delete(h.WsHub.Hub, ws)
	defer delete(h.WsHub.Locks, ws)

	for {
		var command Command
		_, message, err := ws.ReadMessage()
		if err := json.Unmarshal(message, &command); err == nil {
			if command.CommandType == "update_channel" {
				channels := strings.Split(command.Message, CHANNEL_SPLIT_CHAR)
				h.Logger.Sugar().Info(zap.String("update channel", command.Message))
				h.WsHub.Hub[ws] = channels
			}
		}
		if err != nil {
			break
		}
	}
}

// HTTPHandler 接收http端的消息
func (h *ServerHandler) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	cids := getChannels(r)
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "error: %s", err)
		return
	}
	h.WsHub.MsgChan <- &Msg{
		channels: cids,
		content:  content,
	}
}

func getChannels(r *http.Request) []string {
	cids := strings.Split(strings.Trim(r.URL.Path, "/"), CHANNEL_SPLIT_CHAR)
	if cqs, exist := r.URL.Query()["cid"]; exist {
		cids = append(cids, cqs...)
	}
	return cids
}

// ////////////////////////////////////////////////////////////////////////
func (h *WsHub) MonitorMsgChan() {
	for msg := range h.MsgChan {
		for _, channel := range msg.channels {
			for ws, chs := range h.Hub {
				if hasChannel(chs, channel) {
					go h.sendMsg(ws, channel, msg.content)
				}
			}
		}
	}
}

func hasChannel(channels []string, channel string) bool {
	for _, c := range channels {
		if c == channel {
			return true
		}
	}
	return false
}

func (h *WsHub) sendMsg(ws *websocket.Conn, channel string, message []byte) error {
	sockMutex, exist := h.Locks[ws]
	if exist {
		sockMutex.Lock()
		defer sockMutex.Unlock()
	}
	var content interface{}
	json.Unmarshal(message, &content)
	response := &Response{
		Channel: channel,
		Content: content,
	}

	if err := ws.WriteJSON(response); err != nil {
		// log.Infof("Failed sending to client of channel %s", channel)
		return err
	}
	return nil
}
