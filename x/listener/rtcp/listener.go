package rtcp

import (
	"context"
	"net"
	"sync"

	"github.com/168yy/netx/core/chain"
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

type rtcpListener struct {
	laddr   net.Addr
	ln      net.Listener
	router  *chain.Router
	logger  logger.ILogger
	closed  chan struct{}
	options listener.Options
	mu      sync.Mutex
}

func NewListener(opts ...listener.Option) listener.IListener {
	options := listener.Options{}
	for _, opt := range opts {
		opt(&options)
	}
	return &rtcpListener{
		closed:  make(chan struct{}),
		logger:  options.Logger,
		options: options,
	}
}

func (l *rtcpListener) Init(md md.IMetaData) (err error) {
	if err = l.parseMetadata(md); err != nil {
		return
	}

	network := "tcp"
	if xnet.IsIPv4(l.options.Addr) {
		network = "tcp4"
	}
	if laddr, _ := net.ResolveTCPAddr(network, l.options.Addr); laddr != nil {
		l.laddr = laddr
	}
	if l.laddr == nil {
		l.laddr = &bindAddr{addr: l.options.Addr}
	}

	l.router = l.options.Router
	if l.router == nil {
		l.router = chain.NewRouter(chain.LoggerRouterOption(l.logger))
	}

	return
}

func (l *rtcpListener) Accept() (conn net.Conn, err error) {
	select {
	case <-l.closed:
		return nil, net.ErrClosed
	default:
	}

	ln := l.getListener()
	if ln == nil {
		ln, err = l.router.Bind(
			context.Background(), "tcp", l.laddr.String(),
			chain.MuxBindOption(true),
		)
		if err != nil {
			return nil, listener.NewAcceptError(err)
		}
		ln = metrics.WrapListener(l.options.Service, ln)
		ln = stats.WrapListener(ln, l.options.Stats)
		ln = admission.WrapListener(l.options.Admission, ln)
		ln = limiter.WrapListener(l.options.TrafficLimiter, ln)
		ln = climiter.WrapListener(l.options.ConnLimiter, ln)
		l.setListener(ln)
	}

	select {
	case <-l.closed:
		ln.Close()
		return nil, net.ErrClosed
	default:
	}

	conn, err = ln.Accept()
	if err != nil {
		ln.Close()
		l.setListener(nil)
		return nil, listener.NewAcceptError(err)
	}
	return
}

func (l *rtcpListener) Addr() net.Addr {
	return l.laddr
}

func (l *rtcpListener) Close() error {
	select {
	case <-l.closed:
	default:
		close(l.closed)
		if ln := l.getListener(); ln != nil {
			ln.Close()
		}
	}

	return nil
}

func (l *rtcpListener) setListener(ln net.Listener) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ln = ln
}

func (l *rtcpListener) getListener() net.Listener {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.ln
}

type bindAddr struct {
	addr string
}

func (p *bindAddr) Network() string {
	return "tcp"
}

func (p *bindAddr) String() string {
	return p.addr
}
