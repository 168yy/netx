package http

import (
	"context"
	"github.com/168yy/netx/x/internal/net/proxyproto"
	"net"
	"time"

	"github.com/168yy/netx/core/listener"
	"github.com/168yy/netx/core/logger"
	md "github.com/168yy/netx/core/metadata"
	admission "github.com/168yy/netx/x/admission/wrapper"
	xnet "github.com/168yy/netx/x/internal/net"
	climiter "github.com/168yy/netx/x/limiter/conn/wrapper"
	limiter "github.com/168yy/netx/x/limiter/traffic/wrapper"
	metrics "github.com/168yy/netx/x/metrics/wrapper"

	stats "github.com/168yy/netx/x/stats/wrapper"
)

type obfsListener struct {
	net.Listener
	logger  logger.ILogger
	md      metadata
	options listener.Options
}

func NewListener(opts ...listener.Option) listener.IListener {
	options := listener.Options{}
	for _, opt := range opts {
		opt(&options)
	}
	return &obfsListener{
		logger:  options.Logger,
		options: options,
	}
}

func (l *obfsListener) Init(md md.IMetaData) (err error) {
	if err = l.parseMetadata(md); err != nil {
		return
	}

	network := "tcp"
	if xnet.IsIPv4(l.options.Addr) {
		network = "tcp4"
	}

	lc := net.ListenConfig{}
	if l.md.mptcp {
		lc.SetMultipathTCP(true)
		l.logger.Debugf("mptcp enabled: %v", lc.MultipathTCP())
	}
	ln, err := lc.Listen(context.Background(), network, l.options.Addr)
	if err != nil {
		return
	}
	ln = proxyproto.WrapListener(l.options.ProxyProtocol, ln, 10*time.Second)
	ln = metrics.WrapListener(l.options.Service, ln)
	ln = stats.WrapListener(ln, l.options.Stats)
	ln = admission.WrapListener(l.options.Admission, ln)
	ln = limiter.WrapListener(l.options.TrafficLimiter, ln)
	ln = climiter.WrapListener(l.options.ConnLimiter, ln)

	l.Listener = ln
	return
}

func (l *obfsListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &obfsHTTPConn{
		Conn:   c,
		header: l.md.header,
		logger: l.logger,
	}, nil
}
