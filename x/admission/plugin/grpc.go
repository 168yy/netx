package admission

import (
	"context"
	"io"

	"github.com/168yy/netx/core/admission"
	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/plugin/admission/proto"
	"github.com/168yy/netx/x/internal/plugin"
	"google.golang.org/grpc"
)

type grpcPlugin struct {
	conn   grpc.ClientConnInterface
	client proto.AdmissionClient
	log    logger.ILogger
}

// NewGRPCPlugin creates an Admission plugin based on gRPC.
func NewGRPCPlugin(name string, addr string, opts ...plugin.Option) admission.IAdmission {
	var options plugin.Options
	for _, opt := range opts {
		opt(&options)
	}

	log := logger.Default().WithFields(map[string]any{
		"kind":      "admission",
		"admission": name,
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
		p.client = proto.NewAdmissionClient(conn)
	}
	return p
}

func (p *grpcPlugin) Admit(ctx context.Context, addr string, opts ...admission.Option) bool {
	if p.client == nil {
		return false
	}

	r, err := p.client.Admit(ctx,
		&proto.AdmissionRequest{
			Addr: addr,
		})
	if err != nil {
		p.log.Error(err)
		return false
	}
	return r.Ok
}

func (p *grpcPlugin) Close() error {
	if closer, ok := p.conn.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
