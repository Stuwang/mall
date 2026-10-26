package ws

import (
	stdhttp "net/http"
	"strconv"
	"strings"

	"github.com/coder/websocket"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/uuid"

	"mall/im/internal/hub"
	"mall/kit/auth"
	"mall/kit/codeerror"
	"mall/kit/logger"
)

const Path = "/im/ws"

type Config struct {
	IsTest          bool
	HeartbeatSec    int
	IdleSec         int
	MaxConnsPerUser int
}

var (
	online *hub.Hub
	cfg    Config
)

// Register 挂上长连接路径。
func Register(srv *khttp.Server, h *hub.Hub, c Config) {
	online = h
	cfg = c
	if cfg.IdleSec <= 0 {
		cfg.IdleSec = 70
	}
	srv.HandleFunc(Path, handle)
}

func handle(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	info, err := authenticate(r)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	if err := checkActive(info.UserID); err != nil {
		writeAuthError(w, err)
		return
	}

	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		logger.Errorf("ws accept user=%d err=%v", info.UserID, err)
		return
	}

	connID := uuid.NewString()
	c := hub.NewConn(connID, info.UserID, wsConn, cfg.IdleSec, func() {
		online.Remove(info.UserID, connID)
	})
	if evicted := online.Add(c); evicted != nil {
		logger.Infof("ws eviction user=%d oldConn=%s", info.UserID, evicted.ID)
		evicted.Close()
	}
	c.Start()
	// 读写循环在 Conn 内。handle 返回后连接仍由 Start 的 goroutine 持有。
}

func authenticate(r *stdhttp.Request) (*auth.AuthInfo, error) {
	if cfg.IsTest {
		raw := strings.TrimSpace(r.URL.Query().Get("user_id"))
		if raw == "" {
			logger.Errorf("ws test missing user_id")
			return nil, codeerror.ErrorUnauthorized
		}
		uid, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || uid <= 0 {
			logger.Errorf("ws test bad user_id")
			return nil, codeerror.ErrorUnauthorized
		}
		return &auth.AuthInfo{UserID: uid}, nil
	}
	token := strings.TrimSpace(r.URL.Query().Get("access_token"))
	if token == "" {
		authz := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if strings.HasPrefix(authz, prefix) {
			token = strings.TrimSpace(strings.TrimPrefix(authz, prefix))
		}
	}
	if token == "" {
		logger.Errorf("ws missing access_token")
		return nil, codeerror.ErrorUnauthorized
	}
	info, err := auth.ValidateToken(token)
	if err != nil {
		logger.Errorf("ws token invalid")
		return nil, err
	}
	return info, nil
}

var activeChecker func(int64) error

// SetActiveChecker 与 HTTP 共用封禁检查。
func SetActiveChecker(fn func(int64) error) {
	activeChecker = fn
}

func checkActive(userID int64) error {
	if activeChecker == nil {
		return nil
	}
	if err := activeChecker(userID); err != nil {
		logger.Errorf("ws active check user=%d err=%v", userID, err)
		return err
	}
	return nil
}

func writeAuthError(w stdhttp.ResponseWriter, err error) {
	status := stdhttp.StatusUnauthorized
	if ce, ok := codeerror.As(err); ok && ce.Is(codeerror.ErrorUserBanned) {
		status = stdhttp.StatusForbidden
	}
	if ce, ok := codeerror.As(err); ok && ce.Is(codeerror.ErrorForbidden) {
		status = stdhttp.StatusForbidden
	}
	stdhttp.Error(w, "unauthorized", status)
}
