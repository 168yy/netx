package observer

import (
	"context"
	"io"

	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/core/observer"
	"github.com/168yy/netx/plugin/observer/proto"
	"github.com/168yy/netx/x/internal/plugin"
	"github.com/168yy/netx/x/service"
	"github.com/168yy/netx/x/stats"
	"google.golang.org/grpc"
)

type grpcPlugin struct {
	conn   grpc.ClientConnInterface
	client proto.ObserverClient
	log    logger.ILogger
}

// NewGRPCPlugin creates an Observer plugin based on gRPC.
func NewGRPCPlugin(name string, addr string, opts ...plugin.Option) observer.IObserver {
	var options plugin.Options
	for _, opt := range opts {
		opt(&options)
	}

	log := logger.Default().WithFields(map[string]any{
		"kind":     "observer",
		"observer": name,
	})
	conn, err := plugin.NewGRPCConn(addr, &options)
	if err != nil {
		log.Error(err)
		return nil
	}

	p := &grpcPlugin{
		conn: conn,
		log:  log,
	}
	if conn != nil {
		p.client = proto.NewObserverClient(conn)
	}
	return p
}

func (p *grpcPlugin) Observe(ctx context.Context, events []observer.Event, opts ...observer.Option) error {
	if p.client == nil || len(events) == 0 {
		return nil
	}

	var req proto.ObserveRequest

	for _, event := range events {
		switch event.Type() {
		case observer.EventStatus:
			ev := event.(service.ServiceEvent)
			req.Events = append(req.Events, &proto.Event{
				Kind:    ev.Kind,
				Service: ev.Service,
				Type:    string(event.Type()),
				Status: &proto.ServiceStatus{
					State: string(ev.State),
					Msg:   ev.Msg,
				},
			})
		case observer.EventStats:
			ev := event.(stats.StatsEvent)
			req.Events = append(req.Events, &proto.Event{
				Kind:    ev.Kind,
				Service: ev.Service,
				Client:  ev.Client,
				Type:    string(event.Type()),
				Stats: &proto.Stats{
					TotalConns:   ev.TotalConns,
					CurrentConns: ev.CurrentConns,
					InputBytes:   ev.InputBytes,
					OutputBytes:  ev.OutputBytes,
					TotalErrs:    ev.TotalErrs,
				},
			})
		}
	}
	_, err := p.client.Observe(ctx, &req)
	if err != nil {
		p.log.Error(err)
		return err
	}
	return nil
}

func (p *grpcPlugin) Close() error {
	if closer, ok := p.conn.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
