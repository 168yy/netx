package chain

import (
	"context"
	"net"
	"time"

	"github.com/jxo-me/netx/sdk/core/connector"
	"github.com/jxo-me/netx/sdk/core/logger"
	"github.com/jxo-me/netx/sdk/core/metrics"
	xmetrics "github.com/jxo-me/netx/sdk/core/metrics"
	"github.com/jxo-me/netx/sdk/core/selector"
)

type RouteOptions struct {
	Chain IChainer
}

type RouteOption func(*RouteOptions)

func ChainRouteOption(c IChainer) RouteOption {
	return func(o *RouteOptions) {
		o.Chain = c
	}
}

type route struct {
	nodes   []*Node
	options RouteOptions
}

func NewRoute(opts ...RouteOption) *route {
	var options RouteOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return &route{
		options: options,
	}
}

func (r *route) addNode(nodes ...*Node) {
	r.nodes = append(r.nodes, nodes...)
}

func (r *route) Dial(ctx context.Context, network, address string, opts ...DialOption) (net.Conn, error) {
	if len(r.Nodes()) == 0 {
		return DefaultRoute.Dial(ctx, network, address, opts...)
	}

	var options DialOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	conn, err := r.connect(ctx, options.Logger)
	if err != nil {
		return nil, err
	}

	cc, err := r.getNode(len(r.Nodes())-1).Options().Transport.Connect(ctx, conn, network, address)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, err
	}
	return cc, nil
}

func (r *route) Bind(ctx context.Context, network, address string, opts ...BindOption) (net.Listener, error) {
	if len(r.Nodes()) == 0 {
		return DefaultRoute.Bind(ctx, network, address, opts...)
	}

	var options BindOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	conn, err := r.connect(ctx, options.Logger)
	if err != nil {
		return nil, err
	}

	ln, err := r.getNode(len(r.Nodes())-1).Options().Transport.Bind(ctx,
		conn, network, address,
		connector.BacklogBindOption(options.Backlog),
		connector.MuxBindOption(options.Mux),
		connector.UDPConnTTLBindOption(options.UDPConnTTL),
		connector.UDPDataBufferSizeBindOption(options.UDPDataBufferSize),
		connector.UDPDataQueueSizeBindOption(options.UDPDataQueueSize),
	)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return ln, nil
}

func (r *route) connect(ctx context.Context, logger logger.ILogger) (conn net.Conn, err error) {
	network := "ip"
	node := r.nodes[0]

	defer func() {
		if r.options.Chain != nil {
			var marker selector.Marker
			if m, ok := r.options.Chain.(selector.Markable); ok && m != nil {
				marker = m.Marker()
			}
			var name string
			if cn, _ := r.options.Chain.(chainNamer); cn != nil {
				name = cn.Name()
			}
			// chain error
			if err != nil {
				if marker != nil {
					marker.Mark()
				}
				if v := xmetrics.GetCounter(xmetrics.MetricChainErrorsCounter,
					metrics.Labels{"chain": name, "node": node.Name}); v != nil {
					v.Inc()
				}
			} else {
				if marker != nil {
					marker.Reset()
				}
			}
		}
	}()

	addr, err := Resolve(ctx, network, node.Addr, node.Options().Resolver, node.Options().HostMapper, logger)
	marker := node.Marker()
	if err != nil {
		if marker != nil {
			marker.Mark()
		}
		return
	}

	start := time.Now()
	cc, err := node.Options().Transport.Dial(ctx, addr)
	if err != nil {
		if marker != nil {
			marker.Mark()
		}
		return
	}

	cn, err := node.Options().Transport.Handshake(ctx, cc)
	if err != nil {
		cc.Close()
		if marker != nil {
			marker.Mark()
		}
		return
	}
	if marker != nil {
		marker.Reset()
	}

	if r.options.Chain != nil {
		var name string
		if cn, _ := r.options.Chain.(chainNamer); cn != nil {
			name = cn.Name()
		}
		if v := xmetrics.GetObserver(xmetrics.MetricNodeConnectDurationObserver,
			metrics.Labels{"chain": name, "node": node.Name}); v != nil {
			v.Observe(time.Since(start).Seconds())
		}
	}

	preNode := node
	for _, node := range r.nodes[1:] {
		marker := node.Marker()
		addr, err = Resolve(ctx, network, node.Addr, node.Options().Resolver, node.Options().HostMapper, logger)
		if err != nil {
			cn.Close()
			if marker != nil {
				marker.Mark()
			}
			return
		}
		cc, err = preNode.Options().Transport.Connect(ctx, cn, "tcp", addr)
		if err != nil {
			cn.Close()
			if marker != nil {
				marker.Mark()
			}
			return
		}
		cc, err = node.Options().Transport.Handshake(ctx, cc)
		if err != nil {
			cn.Close()
			if marker != nil {
				marker.Mark()
			}
			return
		}
		if marker != nil {
			marker.Reset()
		}

		cn = cc
		preNode = node
	}

	conn = cn
	return
}

func (r *route) getNode(index int) *Node {
	if r == nil || len(r.Nodes()) == 0 || index < 0 || index >= len(r.Nodes()) {
		return nil
	}
	return r.nodes[index]
}

func (r *route) Nodes() []*Node {
	if r != nil {
		return r.nodes
	}
	return nil
}
