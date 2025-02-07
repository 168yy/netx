package router

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/core/router"
	"github.com/168yy/netx/x/internal/plugin"
	xrouter "github.com/168yy/netx/x/router"
)

type httpPluginGetRouteRequest struct {
	Dst string `json:"dst"`
}

type httpPluginGetRouteResponse struct {
	Net     string `json:"net"`
	Gateway string `json:"gateway"`
}

type httpPlugin struct {
	url    string
	client *http.Client
	header http.Header
	log    logger.ILogger
}

// NewHTTPPlugin creates an Router plugin based on HTTP.
func NewHTTPPlugin(name string, url string, opts ...plugin.Option) router.IRouter {
	var options plugin.Options
	for _, opt := range opts {
		opt(&options)
	}

	return &httpPlugin{
		url:    url,
		client: plugin.NewHTTPClient(&options),
		header: options.Header,
		log: logger.Default().WithFields(map[string]any{
			"kind":   "router",
			"router": name,
		}),
	}
}

func (p *httpPlugin) GetRoute(ctx context.Context, dst net.IP, opts ...router.Option) *router.Route {
	if p.client == nil {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return nil
	}
	if p.header != nil {
		req.Header = p.header.Clone()
	}
	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Set("dst", dst.String())
	req.URL.RawQuery = q.Encode()

	resp, err := p.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	res := httpPluginGetRouteResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil
	}

	return xrouter.ParseRoute(res.Net, res.Gateway)
}
