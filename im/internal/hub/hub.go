package hub

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/coder/websocket"

	api "mall/im/api/im"
	"mall/kit/logger"
)

const (
	sendQueueSize = 64
	writeTimeout  = 5 * time.Second
	maxFrameBytes = 64 * 1024
)

// Frame 长连接下行帧。
type Frame struct {
	Cmd   string      `json:"cmd"`
	ReqID string      `json:"req_id,omitempty"`
	Data  interface{} `json:"data"`
}

type inbound struct {
	Cmd   string          `json:"cmd"`
	ReqID string          `json:"req_id"`
	Data  json.RawMessage `json:"data"`
}

// Conn 一条长连接。所有写出都经 sendCh，由单一 writerLoop 写到底层。
type Conn struct {
	ID        string
	UserID    int64
	CreatedAt time.Time

	ws        *websocket.Conn
	idleSec   time.Duration
	sendCh    chan []byte
	done      chan struct{}
	closeOnce sync.Once

	onClose func()
}

// NewConn 创建连接对象。调用 Start 后才开始读写循环。
func NewConn(id string, userID int64, ws *websocket.Conn, idleSec int, onClose func()) *Conn {
	if idleSec <= 0 {
		idleSec = 70
	}
	ws.SetReadLimit(maxFrameBytes)
	return &Conn{
		ID:        id,
		UserID:    userID,
		CreatedAt: time.Now(),
		ws:        ws,
		idleSec:   time.Duration(idleSec) * time.Second,
		sendCh:    make(chan []byte, sendQueueSize),
		done:      make(chan struct{}),
		onClose:   onClose,
	}
}

// Start 启动 writerLoop 与 readerLoop。
func (c *Conn) Start() {
	go c.writerLoop()
	go c.readerLoop()
}

// Enqueue 把已编码帧放入发送队列。队列满或已关闭时关掉连接。
func (c *Conn) Enqueue(payload []byte) {
	select {
	case <-c.done:
		return
	default:
	}
	select {
	case c.sendCh <- payload:
	case <-c.done:
	default:
		logger.Errorf("hub Conn enqueue full user=%d conn=%s", c.UserID, c.ID)
		c.Close()
	}
}

// EnqueueFrame 编码帧后入队。
func (c *Conn) EnqueueFrame(f Frame) {
	bs, err := json.Marshal(f)
	if err != nil {
		logger.Errorf("hub Conn marshal user=%d conn=%s err=%v", c.UserID, c.ID, err)
		return
	}
	c.Enqueue(bs)
}

// Close 关闭发送循环和底层连接。可重复调用。
func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		if c.ws != nil {
			_ = c.ws.Close(websocket.StatusNormalClosure, "")
		}
		if c.onClose != nil {
			c.onClose()
		}
	})
}

func (c *Conn) writerLoop() {
	defer c.Close()
	for {
		select {
		case <-c.done:
			return
		case payload, ok := <-c.sendCh:
			if !ok {
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
			err := c.ws.Write(ctx, websocket.MessageText, payload)
			cancel()
			if err != nil {
				logger.Errorf("hub writerLoop user=%d conn=%s err=%v", c.UserID, c.ID, err)
				return
			}
		}
	}
}

func (c *Conn) readerLoop() {
	defer c.Close()
	for {
		select {
		case <-c.done:
			return
		default:
		}
		ctx, cancel := context.WithTimeout(context.Background(), c.idleSec)
		_, data, err := c.ws.Read(ctx)
		cancel()
		if err != nil {
			return
		}
		var in inbound
		if err := json.Unmarshal(data, &in); err != nil {
			logger.Errorf("hub readerLoop bad frame user=%d conn=%s err=%v", c.UserID, c.ID, err)
			continue
		}
		if in.Cmd != "ping" {
			logger.Errorf("hub readerLoop unknown cmd user=%d conn=%s cmd=%s", c.UserID, c.ID, in.Cmd)
			continue
		}
		c.EnqueueFrame(Frame{Cmd: "pong", ReqID: in.ReqID, Data: map[string]interface{}{}})
	}
}

// Hub 保存本进程内的在线连接。
type Hub struct {
	mu              sync.RWMutex
	users           map[int64]map[string]*Conn
	maxConnsPerUser int
}

func New(maxConnsPerUser int) *Hub {
	if maxConnsPerUser <= 0 {
		maxConnsPerUser = 5
	}
	return &Hub{
		users:           make(map[int64]map[string]*Conn),
		maxConnsPerUser: maxConnsPerUser,
	}
}

// Add 登记连接。若超过上限，关掉该用户最早的一条并返回它。
func (h *Hub) Add(c *Conn) (evicted *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	bucket := h.users[c.UserID]
	if bucket == nil {
		bucket = make(map[string]*Conn)
		h.users[c.UserID] = bucket
	}
	for len(bucket) >= h.maxConnsPerUser {
		var oldest *Conn
		for _, existing := range bucket {
			if oldest == nil || existing.CreatedAt.Before(oldest.CreatedAt) {
				oldest = existing
			}
		}
		if oldest == nil {
			break
		}
		delete(bucket, oldest.ID)
		evicted = oldest
	}
	bucket[c.ID] = c
	return evicted
}

func (h *Hub) Remove(userID int64, connID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	bucket := h.users[userID]
	if bucket == nil {
		return
	}
	delete(bucket, connID)
	if len(bucket) == 0 {
		delete(h.users, userID)
	}
}

// ConnCount 返回该用户当前在线连接数。测试与排障用。
func (h *Hub) ConnCount(userID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.users[userID])
}

// OnlineEntry 一个在线用户及其连接数。
type OnlineEntry struct {
	UserID    int64
	ConnCount int
}

// ListOnline 返回当前在线用户。excludeUserID > 0 时排除该用户。
func (h *Hub) ListOnline(excludeUserID int64) []OnlineEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]OnlineEntry, 0, len(h.users))
	for uid, bucket := range h.users {
		if uid == excludeUserID || len(bucket) == 0 {
			continue
		}
		out = append(out, OnlineEntry{UserID: uid, ConnCount: len(bucket)})
	}
	return out
}

// PushMsg 向该用户所有在线连接推 msg.push。经各连接的发送队列，不并发写同一条 socket。
func (h *Hub) PushMsg(userID int64, msg *api.Msg) error {
	frame := Frame{Cmd: "msg.push", Data: msg}
	bs, err := json.Marshal(frame)
	if err != nil {
		logger.Errorf("hub PushMsg marshal user=%d err=%v", userID, err)
		return err
	}
	h.mu.RLock()
	bucket := h.users[userID]
	conns := make([]*Conn, 0, len(bucket))
	for _, c := range bucket {
		conns = append(conns, c)
	}
	h.mu.RUnlock()
	for _, c := range conns {
		c.Enqueue(bs)
	}
	return nil
}
