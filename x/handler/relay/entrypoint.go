package relay

import (
	"context"
	"net"
	"time"

	"github.com/168yy/netx/core/handler"
	"github.com/168yy/netx/core/listener"
	md "github.com/168yy/netx/core/metadata"
	"github.com/168yy/netx/relay"
	admission "github.com/168yy/netx/x/admission/wrapper"
	xnet "github.com/168yy/netx/x/internal/net"
	"github.com/168yy/netx/x/internal/net/proxyproto"
	"github.com/168yy/netx/x/internal/util/mux"
	climiter "github.com/168yy/netx/x/limiter/conn/wrapper"
	limiter "github.com/168yy/netx/x/limiter/traffic/wrapper"
	metrics "github.com/168yy/netx/x/metrics/wrapper"
)

type tcpListener struct {
	ln      net.Listener
	options listener.Options
}

func newTCPListener(ln net.Listener, opts ...listener.Option) listener.IListener {
	options := listener.Options{}
	for _, opt := range opts {
		opt(&options)
	}
	return &tcpListener{
		ln:      ln,
		options: options,
	}
}

func (l *tcpListener) Init(md md.IMetaData) (err error) {
	// l.logger.Debugf("pp: %d", l.options.ProxyProtocol)
	ln := l.ln
	ln = proxyproto.WrapListener(l.options.ProxyProtocol, ln, 10*time.Second)
	ln = metrics.WrapListener(l.options.Service, ln)
	ln = admission.WrapListener(l.options.Admission, ln)
	ln = limiter.WrapListener(l.options.TrafficLimiter, ln)
	ln = climiter.WrapListener(l.options.ConnLimiter, ln)
	l.ln = ln

	return
}

func (l *tcpListener) Accept() (conn net.Conn, err error) {
	return l.ln.Accept()
}

func (l *tcpListener) Addr() net.Addr {
	return l.ln.Addr()
}

func (l *tcpListener) Close() error {
	return l.ln.Close()
}

type tcpHandler struct {
	session *mux.Session
	options handler.Options
}

func newTCPHandler(session *mux.Session, opts ...handler.Option) handler.IHandler {
	options := handler.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	return &tcpHandler{
		session: session,
		options: options,
	}
}

func (h *tcpHandler) Init(md md.IMetaData) (err error) {
	return
}

func (h *tcpHandler) Handle(ctx context.Context, conn net.Conn, opts ...handler.HandleOption) error {
	defer conn.Close()

	start := time.Now()
	log := h.options.Logger.WithFields(map[string]any{
		"remote": conn.RemoteAddr().String(),
		"local":  conn.LocalAddr().String(),
	})

	log.Infof("%s <> %s", conn.RemoteAddr(), conn.LocalAddr())
	defer func() {
		log.WithFields(map[string]any{
			"duration": time.Since(start),
		}).Infof("%s >< %s", conn.RemoteAddr(), conn.LocalAddr())
	}()

	cc, err := h.session.GetConn()
	if err != nil {
		log.Error(err)
		return err
	}
	defer cc.Close()

	af := &relay.AddrFeature{}
	af.ParseFrom(conn.RemoteAddr().String())
	resp := relay.Response{
		Version:  relay.Version1,
		Status:   relay.StatusOK,
		Features: []relay.Feature{af},
	}
	if _, err := resp.WriteTo(cc); err != nil {
		log.Error(err)
		return err
	}

	t := time.Now()
	log.Debugf("%s <-> %s", conn.RemoteAddr(), cc.RemoteAddr())
	xnet.Transport(conn, cc)
	log.WithFields(map[string]any{"duration": time.Since(t)}).
		Debugf("%s >-< %s", conn.RemoteAddr(), cc.RemoteAddr())
	return nil
}
