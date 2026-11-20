// Package svc 是只在网格内暴露的 gRPC 服务。它同样可以调用任意 svc。
package svc

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	svcv1 "mall/service_mesh_example/api/svc/v1"
	"mall/service_mesh_example/internal/callpath"
	"mall/service_mesh_example/internal/peers"
)

// Server 实现内部 Svc。
type Server struct {
	svcv1.UnimplementedSvcServer

	name string
	pool *peers.Pool
	log  *log.Helper
}

// New 构造 gRPC 服务。
func New(name string, pool *peers.Pool, logger log.Logger) *Server {
	return &Server{
		name: name,
		pool: pool,
		log:  log.NewHelper(log.With(logger, "service", name)),
	}
}

// Ping 返回自己的服务名和路径。
func (s *Server) Ping(_ context.Context, req *svcv1.PingRequest) (*svcv1.PingReply, error) {
	return &svcv1.PingReply{
		Service: s.name,
		Trace:   callpath.AppendTrace(req.GetTrace(), s.name),
	}, nil
}

// Call 先记录自己。depth 大于 1 时再调用其他 svc，并把 depth 减一。
func (s *Server) Call(ctx context.Context, req *svcv1.CallRequest) (*svcv1.CallReply, error) {
	targets := s.pool.TargetsOrDefault(req.GetTargets())
	depth := callpath.ClampDepth(req.GetDepth())
	trace := callpath.AppendTrace(req.GetTrace(), s.name)
	s.log.Infof("call trace=%s targets=%v depth=%d", trace, targets, depth)

	reply := &svcv1.CallReply{
		Service: s.name,
		Trace:   trace,
		Hops:    []*svcv1.Hop{},
	}
	if depth <= 1 {
		return reply, nil
	}
	next := callpath.Downstream(s.name, targets)
	reply.Hops = make([]*svcv1.Hop, 0, len(next))
	for _, name := range next {
		reply.Hops = append(reply.Hops, s.hop(ctx, name, trace, targets, depth-1))
	}
	return reply, nil
}

func (s *Server) hop(ctx context.Context, name, trace string, targets []string, depth int32) *svcv1.Hop {
	downstream, err := s.pool.Call(ctx, name, trace, targets, depth)
	if err != nil {
		s.log.Errorf("call target=%s trace=%s err=%v", name, trace, err)
		return &svcv1.Hop{Service: name, Trace: trace, Error: err.Error()}
	}
	hop := &svcv1.Hop{
		Service: downstream.GetService(),
		Trace:   downstream.GetTrace(),
		Hops:    make([]*svcv1.Hop, 0, len(downstream.GetHops())),
	}
	for _, item := range downstream.GetHops() {
		hop.Hops = append(hop.Hops, item)
	}
	return hop
}
