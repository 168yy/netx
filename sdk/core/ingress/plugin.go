package ingress

import (
	"context"

	xlogger "github.com/jxo-me/netx/sdk/core/logger"
	"github.com/jxo-me/netx/sdk/plugin/ingress/proto"
)

type pluginIngress struct {
	client  proto.IngressClient
	options options
}

// NewPluginIngress creates a plugin ingress.
func NewPluginIngress(opts ...Option) IIngress {
	var options options
	for _, opt := range opts {
		opt(&options)
	}
	if options.logger == nil {
		options.logger = xlogger.Nop()
	}

	p := &pluginIngress{
		options: options,
	}
	if options.client != nil {
		p.client = proto.NewIngressClient(options.client)
	}
	return p
}

func (p *pluginIngress) Get(ctx context.Context, host string) string {
	if p.client == nil {
		return ""
	}

	r, err := p.client.Get(ctx,
		&proto.GetRequest{
			Host: host,
		})
	if err != nil {
		p.options.logger.Error(err)
		return ""
	}
	return r.GetEndpoint()
}

func (p *pluginIngress) Close() error {
	if p.options.client != nil {
		return p.options.client.Close()
	}
	return nil
}
