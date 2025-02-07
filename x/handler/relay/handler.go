package relay

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/168yy/netx/core/chain"
	"github.com/168yy/netx/core/handler"
	"github.com/168yy/netx/core/hop"
	md "github.com/168yy/netx/core/metadata"
	"github.com/168yy/netx/relay"
	ctxvalue "github.com/168yy/netx/x/ctx"
	stats_util "github.com/168yy/netx/x/internal/util/stats"
)

var (
	ErrBadVersion   = errors.New("relay: bad version")
	ErrUnknownCmd   = errors.New("relay: unknown command")
	ErrUnauthorized = errors.New("relay: unauthorized")
	ErrRateLimit    = errors.New("relay: rate limiting exceeded")
)

type relayHandler struct {
	hop     hop.IHop
	router  *chain.Router
	md      metadata
	options handler.Options
	stats   *stats_util.HandlerStats
	cancel  context.CancelFunc
}

func NewHandler(opts ...handler.Option) handler.IHandler {
	options := handler.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	return &relayHandler{
		options: options,
		stats:   stats_util.NewHandlerStats(options.Service),
	}
}

func (h *relayHandler) Init(md md.IMetaData) (err error) {
	if err := h.parseMetadata(md); err != nil {
		return err
	}

	h.router = h.options.Router
	if h.router == nil {
		h.router = chain.NewRouter(chain.LoggerRouterOption(h.options.Logger))
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	if h.options.Observer != nil {
		go h.observeStats(ctx)
	}

	return nil
}

// Forward implements handler.Forwarder.
func (h *relayHandler) Forward(hop hop.IHop) {
	h.hop = hop
}

func (h *relayHandler) Handle(ctx context.Context, conn net.Conn, opts ...handler.HandleOption) (err error) {
	start := time.Now()
	log := h.options.Logger.WithFields(map[string]any{
		"remote": conn.RemoteAddr().String(),
		"local":  conn.LocalAddr().String(),
	})

	log.Infof("%s <> %s", conn.RemoteAddr(), conn.LocalAddr())

	defer func() {
		if err != nil {
			conn.Close()
		}
		log.WithFields(map[string]any{
			"duration": time.Since(start),
		}).Infof("%s >< %s", conn.RemoteAddr(), conn.LocalAddr())
	}()

	if !h.checkRateLimit(conn.RemoteAddr()) {
		return ErrRateLimit
	}

	if h.md.readTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(h.md.readTimeout))
	}

	req := relay.Request{}
	if _, err := req.ReadFrom(conn); err != nil {
		return err
	}

	conn.SetReadDeadline(time.Time{})

	resp := relay.Response{
		Version: relay.Version1,
		Status:  relay.StatusOK,
	}

	if req.Version != relay.Version1 {
		resp.Status = relay.StatusBadRequest
		resp.WriteTo(conn)
		return ErrBadVersion
	}

	var user, pass string
	var address string
	var networkID relay.NetworkID
	for _, f := range req.Features {
		switch f.Type() {
		case relay.FeatureUserAuth:
			if feature, _ := f.(*relay.UserAuthFeature); feature != nil {
				user, pass = feature.Username, feature.Password
			}
		case relay.FeatureAddr:
			if feature, _ := f.(*relay.AddrFeature); feature != nil {
				address = net.JoinHostPort(feature.Host, strconv.Itoa(int(feature.Port)))
			}
		case relay.FeatureNetwork:
			if feature, _ := f.(*relay.NetworkFeature); feature != nil {
				networkID = feature.Network
			}
		}
	}

	if user != "" {
		log = log.WithFields(map[string]any{"user": user})
	}

	if h.options.Auther != nil {
		clientID, ok := h.options.Auther.Authenticate(ctx, user, pass)
		if !ok {
			resp.Status = relay.StatusUnauthorized
			resp.WriteTo(conn)
			return ErrUnauthorized
		}
		ctx = ctxvalue.ContextWithClientID(ctx, ctxvalue.ClientID(clientID))
	}

	network := networkID.String()
	if (req.Cmd & relay.FUDP) == relay.FUDP {
		network = "udp"
	}

	if h.hop != nil {
		defer conn.Close()
		// forward mode
		return h.handleForward(ctx, conn, network, log)
	}

	switch req.Cmd & relay.CmdMask {
	case 0, relay.CmdConnect:
		defer conn.Close()

		return h.handleConnect(ctx, conn, network, address, log)
	case relay.CmdBind:
		defer conn.Close()

		return h.handleBind(ctx, conn, network, address, log)
	default:
		resp.Status = relay.StatusBadRequest
		resp.WriteTo(conn)
		return ErrUnknownCmd
	}
}

// Close implements io.Closer interface.
func (h *relayHandler) Close() error {
	if h.cancel != nil {
		h.cancel()
	}
	return nil
}

func (h *relayHandler) checkRateLimit(addr net.Addr) bool {
	if h.options.RateLimiter == nil {
		return true
	}
	host, _, _ := net.SplitHostPort(addr.String())
	if limiter := h.options.RateLimiter.Limiter(host); limiter != nil {
		return limiter.Allow(1)
	}

	return true
}

func (h *relayHandler) observeStats(ctx context.Context) {
	if h.options.Observer == nil {
		return
	}

	d := h.md.observePeriod
	if d < time.Millisecond {
		d = 5 * time.Second
	}
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.options.Observer.Observe(ctx, h.stats.Events())
		case <-ctx.Done():
			return
		}
	}
}
