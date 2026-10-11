package hash_balancer

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"sync/atomic"

	"github.com/go-kratos/kratos/v2/selector"
)

const BalancerKey = "balancer_key"

var (
	_ selector.Selector = (*Default)(nil)
	_ selector.Builder  = (*DefaultBuilder)(nil)
)

// Default is composite selector.
type Default struct {
	nodes    atomic.Pointer[CHash[selector.Node]]
	fallback selector.Selector
}

func (d *Default) Select(ctx context.Context, opts ...selector.SelectOption) (selected selector.Node, done selector.DoneFunc, err error) {
	c := d.nodes.Load()
	s, ok := ctx.Value(BalancerKey).(string)
	if !ok {
		return d.fallback.Select(ctx, opts...)
	}
	return c.Get(s), func(ctx context.Context, di selector.DoneInfo) {}, nil
}

func (d *Default) Apply(nodes []selector.Node) {
	log.Infof("apply")
	c := NewHash[selector.Node]()
	for _, node := range nodes {
		c.Add(node.Address(), node, 32)
	}
	c.Build()
	d.nodes.Store(c)

	d.fallback.Apply(nodes)
}

func NewBuilder(fb func() selector.Selector) selector.Builder {
	d := &DefaultBuilder{
		fallbackBuilder: fb,
	}
	return d
}

// DefaultBuilder is de
type DefaultBuilder struct {
	fallbackBuilder func() selector.Selector
}

// Build create builder
func (db *DefaultBuilder) Build() selector.Selector {
	return &Default{
		nodes:    atomic.Pointer[CHash[selector.Node]]{},
		fallback: db.fallbackBuilder(),
	}
}
