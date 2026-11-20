package peers

import (
	"slices"

	"google.golang.org/grpc"

	svcv1 "mall/service_mesh_example/api/svc/v1"
)

// svcFactories 把服务名映射到生成出来的客户端。路径由各自的 proto service 决定。
var svcFactories = map[string]func(grpc.ClientConnInterface) svcClient{
	"svc-a": func(conn grpc.ClientConnInterface) svcClient { return svcv1.NewSvcAClient(conn) },
	"svc-b": func(conn grpc.ClientConnInterface) svcClient { return svcv1.NewSvcBClient(conn) },
	"svc-c": func(conn grpc.ClientConnInterface) svcClient { return svcv1.NewSvcCClient(conn) },
}

// Names 返回 svc 名字，按字典序，默认扇出时响应稳定。
func (p *Pool) Names() []string {
	names := make([]string, 0, len(svcFactories))
	for name := range svcFactories {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
