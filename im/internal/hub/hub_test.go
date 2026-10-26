package hub

import (
	"encoding/json"
	"testing"
	"time"

	api "mall/im/api/im"
)

func testConn(id string, userID int64, at time.Time) *Conn {
	return &Conn{
		ID:        id,
		UserID:    userID,
		CreatedAt: at,
		sendCh:    make(chan []byte, sendQueueSize),
		done:      make(chan struct{}),
	}
}

func TestPushMsgAndEvict(t *testing.T) {
	h := New(2)
	base := time.Now()
	c1 := testConn("c1", 1, base)
	c2 := testConn("c2", 1, base.Add(time.Second))
	c3 := testConn("c3", 1, base.Add(2*time.Second))

	if ev := h.Add(c1); ev != nil {
		t.Fatalf("unexpected eviction")
	}
	if ev := h.Add(c2); ev != nil {
		t.Fatalf("unexpected eviction")
	}
	ev := h.Add(c3)
	if ev == nil || ev.ID != "c1" {
		t.Fatalf("want evict c1, got %#v", ev)
	}

	if err := h.PushMsg(1, &api.Msg{MsgId: 9, Seq: 3}); err != nil {
		t.Fatalf("push: %v", err)
	}
	select {
	case <-c1.sendCh:
		t.Fatalf("evicted conn should not receive")
	default:
	}
	for _, c := range []*Conn{c2, c3} {
		select {
		case bs := <-c.sendCh:
			var f Frame
			if err := json.Unmarshal(bs, &f); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if f.Cmd != "msg.push" {
				t.Fatalf("cmd=%s", f.Cmd)
			}
		default:
			t.Fatalf("conn %s missing push", c.ID)
		}
	}
}

func TestEnqueueAfterCloseDropped(t *testing.T) {
	c := testConn("c1", 1, time.Now())
	c.Close()
	c.Enqueue([]byte(`{"cmd":"msg.push"}`))
	select {
	case <-c.sendCh:
		t.Fatalf("closed conn should not enqueue")
	default:
	}
}
