package rudp

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
	limiter "github.com/168yy/netx/x/limiter/traffic/wrapper"
	metrics "github.com/168yy/netx/x/metrics/wrapper"
	stats "github.com/168yy/netx/x/stats/wrapper"
)

type rudpListener struct {
	laddr   net.Addr
	ln      net.Listener
	router  *chain.Router
	closed  chan struct{}
	logger  logger.ILogger
	md      metadata
	options listener.Options
	mu      sync.Mutex
}

func NewListener(opts ...listener.Option) listener.IListener {
	options := listener.Options{}
	for _, opt := range opts {
		opt(&options)
	}
	return &rudpListener{
		closed:  make(chan struct{}),
		logger:  options.Logger,
		options: options,
	}
}

func (l *rudpListener) Init(md md.IMetaData) (err error) {
	if err = l.parseMetadata(md); err != nil {
		return
	}

	network := "udp"
	if xnet.IsIPv4(l.options.Addr) {
		network = "udp4"
	}
	if laddr, _ := net.ResolveUDPAddr(network, l.options.Addr); laddr != nil {
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

func (l *rudpListener) Accept() (conn net.Conn, err error) {
	select {
	case <-l.closed:
		return nil, net.ErrClosed
	default:
	}

	ln := l.getListener()
	if ln == nil {
		ln, err = l.router.Bind(
			context.Background(), "udp", l.laddr.String(),
			chain.BacklogBindOption(l.md.backlog),
			chain.UDPConnTTLBindOption(l.md.ttl),
			chain.UDPDataBufferSizeBindOption(l.md.readBufferSize),
			chain.UDPDataQueueSizeBindOption(l.md.readQueueSize),
		)
		if err != nil {
			return nil, listener.NewAcceptError(err)
		}
		l.setListener(ln)
	}

	select {
	case <-l.closed:
		ln.Close()
		return nil, net.ErrClosed
	default:
	}

	conn, err = l.ln.Accept()
	if err != nil {
		l.ln.Close()
		l.setListener(nil)
		return nil, listener.NewAcceptError(err)
	}

	if pc, ok := conn.(net.PacketConn); ok {
		uc := metrics.WrapUDPConn(l.options.Service, pc)
		uc = stats.WrapUDPConn(uc, l.options.Stats)
		uc = admission.WrapUDPConn(l.options.Admission, uc)
		conn = limiter.WrapUDPConn(l.options.TrafficLimiter, uc)
	}

	return
}

func (l *rudpListener) Addr() net.Addr {
	return l.laddr
}

func (l *rudpListener) Close() error {
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

func (l *rudpListener) setListener(ln net.Listener) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ln = ln
}

func (l *rudpListener) getListener() net.Listener {
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
