// Package peers 持有到 API Gateway 的一条 gRPC 连接。
// svc-a、svc-b、svc-c 走同一条连接，靠 proto 方法路径区分。
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

// svcClient 是三个生成客户端共有的 Call 方法。
type svcClient interface {
	Call(ctx context.Context, in *svcv1.CallRequest, opts ...grpc.CallOption) (*svcv1.CallReply, error)
}

// Pool 复用一条到网关的连接。第一次调用时才拨号，避免启动顺序互相卡死。
type Pool struct {
	addr string
	log  *log.Helper

	mu      sync.Mutex
	conn    *grpc.ClientConn
	clients map[string]svcClient
}

// New 用网关地址构造连接池。三个 svc 都从这条连接出去。
func New(addr string, logger log.Logger) *Pool {
	return &Pool{
		addr: addr,
		log:  log.NewHelper(logger),
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

func (p *Pool) client(name string) (svcClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if client, ok := p.clients[name]; ok {
		return client, nil
	}
	if _, ok := svcFactories[name]; !ok && p.conn != nil {
		return nil, fmt.Errorf("unknown svc %q", name)
	}
	if p.conn == nil {
		if err := p.dial(); err != nil {
			return nil, err
		}
	}
	client, ok := p.clients[name]
	if !ok {
		return nil, fmt.Errorf("unknown svc %q", name)
	}
	return client, nil
}

func (p *Pool) dial() error {
	endpoint := p.addr
	if !strings.Contains(endpoint, "://") {
		endpoint = "direct:///" + endpoint
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 网关本身不是某个 svc 的健康检查端点，不能按单个服务做 grpc health。
	conn, err := kgrpc.DialInsecure(
		ctx,
		kgrpc.WithEndpoint(endpoint),
		kgrpc.WithTimeout(3*time.Second),
		kgrpc.WithHealthCheck(false),
	)
	if err != nil {
		return fmt.Errorf("dial gateway %s: %w", p.addr, err)
	}
	clients := make(map[string]svcClient, len(svcFactories))
	for name, newClient := range svcFactories {
		clients[name] = newClient(conn)
	}
	p.conn = conn
	p.clients = clients
	p.log.Infof("dialed gateway=%s", p.addr)
	return nil
}

// TargetsOrDefault 在请求没写目标时使用全部 svc。
func (p *Pool) TargetsOrDefault(targets []string) []string {
	normalized := callpath.NormalizeTargets(targets)
	if len(normalized) == 0 {
		return p.Names()
	}
	return normalized
}
