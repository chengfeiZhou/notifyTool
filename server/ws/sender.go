package ws

import (
	"encoding/json"
	"notifyTool/utils/msg"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Response struct {
	Channel string      `json:"channel"`
	Content interface{} `json:"content"`
}

type WsHub struct {
	Hub map[*websocket.Conn][]string
	*sync.Mutex
}

type WsSrv struct {
	*WsHub
	Logger  *zap.Logger
	MsgChan <-chan *msg.Msg
}

func (ws *WsSrv) MonitorMsgChan() {
	for msg := range ws.MsgChan {
		for _, channel := range msg.Channels {
			for wsc, chs := range ws.Hub {
				if hasChannel(chs, channel) {
					go ws.sendMsg(wsc, channel, msg.Content)
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

func (ws *WsSrv) sendMsg(wsc *websocket.Conn, channel string, message []byte) error {
	ws.Lock()
	defer ws.Unlock()
	var content interface{}
	if err := json.Unmarshal(message, &content); err != nil {
		ws.Logger.Sugar().Error("Failed Unmarshal message", zap.Error(err))
		return err
	}
	response := &Response{
		Channel: channel,
		Content: content,
	}
	if err := wsc.WriteJSON(response); err != nil {
		ws.Logger.Sugar().Error("Failed sending to client of channel",
			zap.String("channel", channel), zap.Error(err))
		return err
	}
	return nil
}
