package service

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"notifyTool/utils/server"
	"testing"
)

func TestReceivedNotify(t *testing.T) {
	lg, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create zap logger: %v", err)
	}
	sh := CreateServerHandler(lg)
	url, _ := url.Parse("http://127.0.0.1:8900/path/ooo?cid=bar&cid=abc")
	data := map[string]string{
		"abc": "123",
	}
	jsonData, _ := json.Marshal(data)
	resp := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, url.String(), bytes.NewReader(jsonData))
	sh.HTTPHandler(resp, request)

}

func TestSendWS(t *testing.T) {
	lg, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create zap logger: %v", err)
	}
	sh := CreateServerHandler(lg)
	u := url.URL{
		Scheme:     "ws",
		Host:       "localhost:8900",
		Path:       "/path|ooo/",
		ForceQuery: false,
		RawQuery:   "cid=bar&cid=abc",
	}
	resp := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, u.String(), nil)
	sh.WSHandler(resp, request)
}
