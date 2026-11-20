// Package health 提供探针端口。这个端口排除在 Consul 透明代理之外，kubelet 才能直接访问。
package health

import (
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewServer 在 addr 上提供 /healthz。
func NewServer(addr string) *khttp.Server {
	srv := khttp.NewServer(khttp.Address(addr))
	srv.Route("/").GET("/healthz", func(ctx khttp.Context) error {
		return ctx.Result(200, map[string]string{"status": "ok"})
	})
	return srv
}
