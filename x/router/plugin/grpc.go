package router

import (
	"context"
	"io"
	"net"

	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/core/router"
	"github.com/168yy/netx/plugin/router/proto"
	"github.com/168yy/netx/x/internal/plugin"
	xrouter "github.com/168yy/netx/x/router"
	"google.golang.org/grpc"
)

type grpcPlugin struct {
	conn   grpc.ClientConnInterface
	client proto.RouterClient
	log    logger.ILogger
}

// NewGRPCPlugin creates an Router plugin based on gRPC.
func NewGRPCPlugin(name string, addr string, opts ...plugin.Option) router.IRouter {
	var options plugin.Options
	for _, opt := range opts {
		opt(&options)
	}

	log := logger.Default().WithFields(map[string]any{
		"kind":   "router",
		"router": name,
	})
	conn, err := plugin.NewGRPCConn(addr, &options)
	if err != nil {
		log.Error(err)
	}

	p := &grpcPlugin{
		conn: conn,
		log:  log,
	}
	if conn != nil {
		p.client = proto.NewRouterClient(conn)
	}
	return p
}

func (p *grpcPlugin) GetRoute(ctx context.Context, dst net.IP, opts ...router.Option) *router.Route {
	if p.client == nil {
		return nil
	}

	r, err := p.client.GetRoute(ctx,
		&proto.GetRouteRequest{
			Dst: dst.String(),
		})
	if err != nil {
		p.log.Error(err)
		return nil
	}

	return xrouter.ParseRoute(r.Net, r.Gateway)
}

func (p *grpcPlugin) Close() error {
	if p.conn == nil {
		return nil
	}

	if closer, ok := p.conn.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
