package biz

import (
	"fmt"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	bizv1 "mall/service_mesh_example/api/biz/v1"
)

// RegisterHTTP 只挂上当前进程对应的路径。biz-a 不会同时暴露 /biz_b。
func RegisterHTTP(name string, srv *khttp.Server, impl *Server) error {
	switch name {
	case "biz-a":
		bizv1.RegisterBizAHTTPServer(srv, impl)
	case "biz-b":
		bizv1.RegisterBizBHTTPServer(srv, impl)
	case "biz-c":
		bizv1.RegisterBizCHTTPServer(srv, impl)
	default:
		return fmt.Errorf("unknown biz %q", name)
	}
	return nil
}
