package ws

import (
	stdhttp "net/http"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"mall/im/internal/hub"
	"mall/kit/logger"
)

const Path = "/im/ws"

var online *hub.Hub

// Register 挂上长连接路径。升级、心跳和推送尚未实现。
func Register(srv *khttp.Server, h *hub.Hub) {
	online = h
	srv.HandleFunc(Path, func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		logger.Errorf("websocket not implemented path=%s hub=%v", r.URL.Path, online != nil)
		stdhttp.Error(w, "not implemented", stdhttp.StatusNotImplemented)
	})
}
