// Package biz 是对外 HTTP 服务。它通过 gRPC 调用任意内部 svc。
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	bizv1 "mall/service_mesh_example/api/biz/v1"
	svcv1 "mall/service_mesh_example/api/svc/v1"
	"mall/service_mesh_example/internal/callpath"
	"mall/service_mesh_example/internal/peers"
)

// Server 实现 Biz 的 HTTP 接口。
type Server struct {
	name string
	pool *peers.Pool
	log  *log.Helper
}

// New 构造 HTTP 业务服务。
func New(name string, pool *peers.Pool, logger log.Logger) *Server {
	return &Server{
		name: name,
		pool: pool,
		log:  log.NewHelper(log.With(logger, "service", name)),
	}
}

// Ping 只返回自己的服务名，不进入网格。
func (s *Server) Ping(context.Context, *bizv1.PingRequest) (*bizv1.PingReply, error) {
	return &bizv1.PingReply{Service: s.name}, nil
}

// Call 按目标扇出到 svc。单个下游失败会记在 hop 里，其余目标继续。
func (s *Server) Call(ctx context.Context, req *bizv1.CallRequest) (*bizv1.CallReply, error) {
	targets := s.pool.TargetsOrDefault(req.GetTargets())
	depth := callpath.ClampDepth(req.GetDepth())
	trace := callpath.AppendTrace(req.GetTrace(), s.name)
	s.log.Infof("call trace=%s targets=%v depth=%d", trace, targets, depth)

	reply := &bizv1.CallReply{
		Service: s.name,
		Trace:   trace,
		Hops:    make([]*bizv1.Hop, 0, len(targets)),
	}
	for _, name := range targets {
		reply.Hops = append(reply.Hops, s.hop(ctx, name, trace, targets, depth))
	}
	return reply, nil
}

func (s *Server) hop(ctx context.Context, name, trace string, targets []string, depth int32) *bizv1.Hop {
	downstream, err := s.pool.Call(ctx, name, trace, targets, depth)
	if err != nil {
		s.log.Errorf("call target=%s trace=%s err=%v", name, trace, err)
		return &bizv1.Hop{Service: name, Trace: trace, Error: err.Error()}
	}
	return toHop(downstream)
}

func toHop(reply *svcv1.CallReply) *bizv1.Hop {
	hop := &bizv1.Hop{
		Service: reply.GetService(),
		Trace:   reply.GetTrace(),
		Hops:    make([]*bizv1.Hop, 0, len(reply.GetHops())),
	}
	for _, item := range reply.GetHops() {
		hop.Hops = append(hop.Hops, mapHop(item))
	}
	return hop
}

func mapHop(item *svcv1.Hop) *bizv1.Hop {
	hop := &bizv1.Hop{
		Service: item.GetService(),
		Trace:   item.GetTrace(),
		Error:   item.GetError(),
		Hops:    make([]*bizv1.Hop, 0, len(item.GetHops())),
	}
	for _, child := range item.GetHops() {
		hop.Hops = append(hop.Hops, mapHop(child))
	}
	return hop
}
