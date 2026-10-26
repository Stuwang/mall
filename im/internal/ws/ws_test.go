package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"

	api "mall/im/api/im"
	"mall/im/internal/hub"
)

func TestPingPongAndPush(t *testing.T) {
	h := hub.New(5)
	online = h
	cfg = Config{IsTest: true, IdleSec: 5}

	srv := httptest.NewServer(http.HandlerFunc(handle))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+srv.URL[len("http"):]+"?user_id=42", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// 等服务端把连接登记进 Hub。
	deadline := time.Now().Add(time.Second)
	for h.ConnCount(42) == 0 {
		if time.Now().After(deadline) {
			t.Fatalf("conn not registered")
		}
		time.Sleep(10 * time.Millisecond)
	}

	ping, _ := json.Marshal(map[string]interface{}{"cmd": "ping", "req_id": "r1", "data": map[string]interface{}{}})
	if err := conn.Write(ctx, websocket.MessageText, ping); err != nil {
		t.Fatalf("write ping: %v", err)
	}
	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read pong: %v", err)
	}
	var pong hub.Frame
	if err := json.Unmarshal(data, &pong); err != nil {
		t.Fatalf("unmarshal pong: %v", err)
	}
	if pong.Cmd != "pong" || pong.ReqID != "r1" {
		t.Fatalf("pong %#v", pong)
	}

	if err := h.PushMsg(42, &api.Msg{MsgId: 7, Seq: 1, ConvId: "1_42"}); err != nil {
		t.Fatalf("push: %v", err)
	}
	_, data, err = conn.Read(ctx)
	if err != nil {
		t.Fatalf("read push: %v", err)
	}
	var push hub.Frame
	if err := json.Unmarshal(data, &push); err != nil {
		t.Fatalf("unmarshal push: %v", err)
	}
	if push.Cmd != "msg.push" {
		t.Fatalf("push cmd=%s", push.Cmd)
	}
}

func TestRejectUnauthorized(t *testing.T) {
	online = hub.New(5)
	cfg = Config{IsTest: true, IdleSec: 5}
	srv := httptest.NewServer(http.HandlerFunc(handle))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, resp, err := websocket.Dial(ctx, "ws"+srv.URL[len("http"):], nil)
	if err == nil {
		t.Fatalf("want auth failure")
	}
	if resp != nil && resp.StatusCode != 401 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}
