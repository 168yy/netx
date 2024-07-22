package udp

import (
	"context"
	"net"

	"github.com/168yy/netx/core/dialer"
	"github.com/168yy/netx/core/logger"
	md "github.com/168yy/netx/core/metadata"
)

type udpDialer struct {
	md     metadata
	logger logger.ILogger
}

func NewDialer(opts ...dialer.Option) dialer.IDialer {
	options := &dialer.Options{}
	for _, opt := range opts {
		opt(options)
	}

	return &udpDialer{
		logger: options.Logger,
	}
}

func (d *udpDialer) Init(md md.IMetaData) (err error) {
	return d.parseMetadata(md)
}

func (d *udpDialer) Dial(ctx context.Context, addr string, opts ...dialer.DialOption) (net.Conn, error) {
	var options dialer.DialOptions
	for _, opt := range opts {
		opt(&options)
	}

	c, err := options.NetDialer.Dial(ctx, "udp", addr)
	if err != nil {
		return nil, err
	}
	return &conn{
		UDPConn: c.(*net.UDPConn),
	}, nil
}
