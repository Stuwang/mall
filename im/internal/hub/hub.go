package hub

import "sync"

// Conn 一条长连接。读循环和推送尚未实现。
type Conn struct {
	ID     string
	UserID int64
}

// Hub 保存本进程内的在线连接。
type Hub struct {
	mu    sync.RWMutex
	users map[int64]map[string]*Conn
}

func New() *Hub {
	return &Hub{users: make(map[int64]map[string]*Conn)}
}

func (h *Hub) Add(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	bucket := h.users[c.UserID]
	if bucket == nil {
		bucket = make(map[string]*Conn)
		h.users[c.UserID] = bucket
	}
	bucket[c.ID] = c
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
