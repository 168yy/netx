package http

import (
	"context"
	"net"

	"github.com/jxo-me/netx/core/dialer"
	"github.com/jxo-me/netx/core/logger"
	md "github.com/jxo-me/netx/core/metadata"
	"github.com/jxo-me/netx/x/registry"
)

func init() {
	registry.DialerRegistry().Register("ohttp", NewDialer)
}

type obfsHTTPDialer struct {
	md     metadata
	logger logger.ILogger
}

func NewDialer(opts ...dialer.Option) dialer.IDialer {
	options := &dialer.Options{}
	for _, opt := range opts {
		opt(options)
	}

	return &obfsHTTPDialer{
		logger: options.Logger,
	}
}

func (d *obfsHTTPDialer) Init(md md.IMetaData) (err error) {
	return d.parseMetadata(md)
}

func (d *obfsHTTPDialer) Dial(ctx context.Context, addr string, opts ...dialer.DialOption) (net.Conn, error) {
	options := &dialer.DialOptions{}
	for _, opt := range opts {
		opt(options)
	}

	conn, err := options.NetDialer.Dial(ctx, "tcp", addr)
	if err != nil {
		d.logger.Error(err)
	}
	return conn, err
}

// Handshake implements dialer.IHandshaker
func (d *obfsHTTPDialer) Handshake(ctx context.Context, conn net.Conn, options ...dialer.HandshakeOption) (net.Conn, error) {
	opts := &dialer.HandshakeOptions{}
	for _, option := range options {
		option(opts)
	}

	host := d.md.host
	if host == "" {
		host = opts.Addr
	}

	return &obfsHTTPConn{
		Conn:   conn,
		host:   host,
		header: d.md.header,
		logger: d.logger,
	}, nil
}
