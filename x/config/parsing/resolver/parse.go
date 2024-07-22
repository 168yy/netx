package resolver

import (
	"crypto/tls"
	"github.com/168yy/netx/x/app"
	"net"
	"strings"

	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/core/resolver"
	"github.com/168yy/netx/x/config"
	"github.com/168yy/netx/x/internal/plugin"
	xresolver "github.com/168yy/netx/x/resolver"
	resolverplugin "github.com/168yy/netx/x/resolver/plugin"
)

func ParseResolver(cfg *config.ResolverConfig) (resolver.IResolver, error) {
	if cfg == nil {
		return nil, nil
	}

	if cfg.Plugin != nil {
		var tlsCfg *tls.Config
		if cfg.Plugin.TLS != nil {
			tlsCfg = &tls.Config{
				ServerName:         cfg.Plugin.TLS.ServerName,
				InsecureSkipVerify: !cfg.Plugin.TLS.Secure,
			}
		}
		switch strings.ToLower(cfg.Plugin.Type) {
		case "http":
			return resolverplugin.NewHTTPPlugin(
				cfg.Name, cfg.Plugin.Addr,
				plugin.TLSConfigOption(tlsCfg),
				plugin.TimeoutOption(cfg.Plugin.Timeout),
			), nil
		default:
			return resolverplugin.NewGRPCPlugin(
				cfg.Name, cfg.Plugin.Addr,
				plugin.TokenOption(cfg.Plugin.Token),
				plugin.TLSConfigOption(tlsCfg),
			)
		}
	}

	var nameservers []xresolver.NameServer
	for _, server := range cfg.Nameservers {
		nameservers = append(nameservers, xresolver.NameServer{
			Addr:     server.Addr,
			Chain:    app.Runtime.ChainRegistry().Get(server.Chain),
			TTL:      server.TTL,
			Timeout:  server.Timeout,
			ClientIP: net.ParseIP(server.ClientIP),
			Prefer:   server.Prefer,
			Hostname: server.Hostname,
			Async:    server.Async,
			Only:     server.Only,
		})
	}

	return xresolver.NewResolver(
		nameservers,
		xresolver.LoggerOption(
			logger.Default().WithFields(map[string]any{
				"kind":     "resolver",
				"resolver": cfg.Name,
			}),
		),
	)
}
