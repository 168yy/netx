package direct

import (
	"context"
	"net"

	"github.com/168yy/netx/core/dialer"
	"github.com/168yy/netx/core/logger"
	md "github.com/168yy/netx/core/metadata"
)

type directDialer struct {
	logger logger.ILogger
}

func NewDialer(opts ...dialer.Option) dialer.IDialer {
	options := &dialer.Options{}
	for _, opt := range opts {
		opt(options)
	}

	return &directDialer{
		logger: options.Logger,
	}
}

func (d *directDialer) Init(md md.IMetaData) (err error) {
	return
}

func (d *directDialer) Dial(ctx context.Context, addr string, opts ...dialer.DialOption) (net.Conn, error) {
	return &conn{}, nil
}
