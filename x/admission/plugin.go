package admission

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/jxo-me/netx/core/admission"
	"github.com/jxo-me/netx/core/logger"
	"github.com/jxo-me/netx/plugin/admission/proto"
	"github.com/jxo-me/netx/x/internal/util/plugin"
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

type httpPluginRequest struct {
	Addr string `json:"addr"`
}

type httpPluginResponse struct {
	OK bool `json:"ok"`
}

type httpPlugin struct {
	url    string
	client *http.Client
	header http.Header
	log    logger.ILogger
}

// NewHTTPPlugin creates an Admission plugin based on HTTP.
func NewHTTPPlugin(name string, url string, opts ...plugin.Option) admission.IAdmission {
	var options plugin.Options
	for _, opt := range opts {
		opt(&options)
	}

	return &httpPlugin{
		url:    url,
		client: plugin.NewHTTPClient(&options),
		header: options.Header,
		log: logger.Default().WithFields(map[string]any{
			"kind":      "admission",
			"admission": name,
		}),
	}
}

func (p *httpPlugin) Admit(ctx context.Context, addr string, opts ...admission.Option) (ok bool) {
	if p.client == nil {
		return
	}

	rb := httpPluginRequest{
		Addr: addr,
	}
	v, err := json.Marshal(&rb)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(v))
	if err != nil {
		return
	}

	if p.header != nil {
		req.Header = p.header.Clone()
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	res := httpPluginResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return
	}
	return res.OK
}
