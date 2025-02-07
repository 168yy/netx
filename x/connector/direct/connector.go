package forward

import (
	"context"
	"net"

	"github.com/168yy/netx/core/connector"
	md "github.com/168yy/netx/core/metadata"
)

type directConnector struct {
	options connector.Options
}

func NewConnector(opts ...connector.Option) connector.IConnector {
	options := connector.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	return &directConnector{
		options: options,
	}
}

func (c *directConnector) Init(md md.IMetaData) (err error) {
	return nil
}

func (c *directConnector) Connect(ctx context.Context, _ net.Conn, network, address string, opts ...connector.ConnectOption) (net.Conn, error) {
	var cOpts connector.ConnectOptions
	for _, opt := range opts {
		opt(&cOpts)
	}

	conn, err := cOpts.NetDialer.Dial(ctx, network, address)
	if err != nil {
		return nil, err
	}

	var localAddr, remoteAddr string
	if addr := conn.LocalAddr(); addr != nil {
		localAddr = addr.String()
	}
	if addr := conn.RemoteAddr(); addr != nil {
		remoteAddr = addr.String()
	}

	log := c.options.Logger.WithFields(map[string]any{
		"remote":  remoteAddr,
		"local":   localAddr,
		"network": network,
		"address": address,
	})
	log.Debugf("connect %s/%s", address, network)

	return conn, nil
}
