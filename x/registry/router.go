package registry

import (
	"context"
	"net"

	"github.com/jxo-me/netx/core/router"
)

type routerRegistry struct {
	registry[router.IRouter]
}

func (r *routerRegistry) Register(name string, v router.IRouter) error {
	return r.registry.Register(name, v)
}

func (r *routerRegistry) Get(name string) router.IRouter {
	if name != "" {
		return &routerWrapper{name: name, r: r}
	}
	return nil
}

func (r *routerRegistry) get(name string) router.IRouter {
	return r.registry.Get(name)
}

type routerWrapper struct {
	name string
	r    *routerRegistry
}

func (w *routerWrapper) GetRoute(ctx context.Context, dst net.IP, opts ...router.Option) *router.Route {
	v := w.r.get(w.name)
	if v == nil {
		return nil
	}
	return v.GetRoute(ctx, dst, opts...)
}
