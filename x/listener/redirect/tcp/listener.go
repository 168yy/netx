package tcp

import (
	"context"
	"net"
	"time"

	"github.com/168yy/netx/core/listener"
	"github.com/168yy/netx/core/logger"
	md "github.com/168yy/netx/core/metadata"
	admission "github.com/168yy/netx/x/admission/wrapper"
	xnet "github.com/168yy/netx/x/internal/net"
	"github.com/168yy/netx/x/internal/net/proxyproto"
	climiter "github.com/168yy/netx/x/limiter/conn/wrapper"
	limiter "github.com/168yy/netx/x/limiter/traffic/wrapper"
	metrics "github.com/168yy/netx/x/metrics/wrapper"
	stats "github.com/168yy/netx/x/stats/wrapper"
)

type redirectListener struct {
	ln      net.Listener
	logger  logger.ILogger
	md      metadata
	options listener.Options
}

func NewListener(opts ...listener.Option) listener.IListener {
	options := listener.Options{}
	for _, opt := range opts {
		opt(&options)
	}
	return &redirectListener{
		logger:  options.Logger,
		options: options,
	}
}

func (l *redirectListener) Init(md md.IMetaData) (err error) {
	if err = l.parseMetadata(md); err != nil {
		return
	}

	network := "tcp"
	if xnet.IsIPv4(l.options.Addr) {
		network = "tcp4"
	}
	lc := net.ListenConfig{}
	if l.md.tproxy {
		lc.Control = l.control
	}
	if l.md.mptcp {
		lc.SetMultipathTCP(true)
		l.logger.Debugf("mptcp enabled: %v", lc.MultipathTCP())
	}
	ln, err := lc.Listen(context.Background(), network, l.options.Addr)
	if err != nil {
		return err
	}

	ln = proxyproto.WrapListener(l.options.ProxyProtocol, ln, 10*time.Second)
	ln = metrics.WrapListener(l.options.Service, ln)
	ln = stats.WrapListener(ln, l.options.Stats)
	ln = admission.WrapListener(l.options.Admission, ln)
	ln = limiter.WrapListener(l.options.TrafficLimiter, ln)
	ln = climiter.WrapListener(l.options.ConnLimiter, ln)
	l.ln = ln
	return
}

func (l *redirectListener) Accept() (conn net.Conn, err error) {
	return l.ln.Accept()
}

func (l *redirectListener) Addr() net.Addr {
	return l.ln.Addr()
}

func (l *redirectListener) Close() error {
	return l.ln.Close()
}
