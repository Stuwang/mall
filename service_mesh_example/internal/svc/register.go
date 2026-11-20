package svc

import (
	"fmt"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	svcv1 "mall/service_mesh_example/api/svc/v1"
)

// Register 只注册当前进程的 proto service。三个服务的方法路径因此不会叠在同一个进程上。
func Register(name string, srv *kgrpc.Server, impl *Server) error {
	switch name {
	case "svc-a":
		svcv1.RegisterSvcAServer(srv, impl)
	case "svc-b":
		svcv1.RegisterSvcBServer(srv, impl)
	case "svc-c":
		svcv1.RegisterSvcCServer(srv, impl)
	default:
		return fmt.Errorf("unknown svc %q", name)
	}
	return nil
}
