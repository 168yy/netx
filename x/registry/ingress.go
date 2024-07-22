package registry

import (
	"context"

	"github.com/168yy/netx/core/ingress"
)

type IngressRegistry struct {
	registry[ingress.IIngress]
}

func (r *IngressRegistry) Register(name string, v ingress.IIngress) error {
	return r.registry.Register(name, v)
}

func (r *IngressRegistry) Get(name string) ingress.IIngress {
	if name != "" {
		return &ingressWrapper{name: name, r: r}
	}
	return nil
}

func (r *IngressRegistry) get(name string) ingress.IIngress {
	return r.registry.Get(name)
}

type ingressWrapper struct {
	name string
	r    *IngressRegistry
}

func (w *ingressWrapper) GetRule(ctx context.Context, host string, opts ...ingress.Option) *ingress.Rule {
	v := w.r.get(w.name)
	if v == nil {
		return nil
	}
	return v.GetRule(ctx, host, opts...)
}

func (w *ingressWrapper) SetRule(ctx context.Context, rule *ingress.Rule, opts ...ingress.Option) bool {
	v := w.r.get(w.name)
	if v == nil {
		return false
	}

	return v.SetRule(ctx, rule, opts...)
}
