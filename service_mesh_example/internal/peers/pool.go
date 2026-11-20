// Package peers 持有到内部 svc 的 gRPC 连接。地址使用 Kubernetes DNS，
// 流量由 Consul 透明代理劫持并做 mTLS。
package peers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc"

	svcv1 "mall/service_mesh_example/api/svc/v1"
	"mall/service_mesh_example/internal/callpath"
)

// Pool 按服务名复用连接。第一次调用时才拨号，避免启动顺序互相卡死。
type Pool struct {
	addrs map[string]string
	log   *log.Helper

	mu    sync.Mutex
	conns map[string]*grpc.ClientConn
}

// New 用服务名到地址的映射构造连接池。
func New(addrs map[string]string, logger log.Logger) *Pool {
	copied := make(map[string]string, len(addrs))
	for name, addr := range addrs {
		copied[name] = addr
	}
	return &Pool{
		addrs: copied,
		log:   log.NewHelper(logger),
		conns: map[string]*grpc.ClientConn{},
	}
}

// Call 调用指定 svc。未知服务名直接返回错误。
func (p *Pool) Call(ctx context.Context, name, trace string, targets []string, depth int32) (*svcv1.CallReply, error) {
	client, err := p.client(name)
	if err != nil {
		p.log.Errorf("dial peer=%s err=%v", name, err)
		return nil, err
	}
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	reply, err := client.Call(cctx, &svcv1.CallRequest{
		Trace:   trace,
		Targets: targets,
		Depth:   depth,
	})
	if err != nil {
		p.log.Errorf("call peer=%s trace=%s err=%v", name, trace, err)
		return nil, err
	}
	return reply, nil
}

func (p *Pool) client(name string) (svcv1.SvcClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if conn, ok := p.conns[name]; ok {
		return svcv1.NewSvcClient(conn), nil
	}
	addr, ok := p.addrs[name]
	if !ok {
		return nil, fmt.Errorf("unknown svc %q", name)
	}
	endpoint := addr
	if !strings.Contains(addr, "://") {
		// Kratos 的 direct 解析器把 host:port 交给 gRPC 拨号，DNS 由透明代理之前的内核完成。
		endpoint = "direct:///" + addr
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := kgrpc.DialInsecure(
		ctx,
		kgrpc.WithEndpoint(endpoint),
		kgrpc.WithTimeout(3*time.Second),
		kgrpc.WithHealthCheck(true),
	)
	if err != nil {
		return nil, fmt.Errorf("dial %s %s: %w", name, addr, err)
	}
	p.conns[name] = conn
	p.log.Infof("dialed peer=%s addr=%s", name, addr)
	return svcv1.NewSvcClient(conn), nil
}

// TargetsOrDefault 在请求没写目标时使用全部 svc。
func (p *Pool) TargetsOrDefault(targets []string) []string {
	normalized := callpath.NormalizeTargets(targets)
	if len(normalized) == 0 {
		return p.Names()
	}
	return normalized
}
