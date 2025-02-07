package chain

import (
	"context"
	"fmt"
	"net"

	"github.com/168yy/netx/core/hosts"
	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/core/resolver"
)

func Resolve(ctx context.Context, network, addr string, r resolver.IResolver, hosts hosts.IHostMapper, log logger.ILogger) (string, error) {
	if addr == "" {
		return addr, nil
	}

	host, port, _ := net.SplitHostPort(addr)
	if host == "" {
		return addr, nil
	}

	if hosts != nil {
		if ips, _ := hosts.Lookup(ctx, network, host); len(ips) > 0 {
			log.Debugf("hit host mapper: %s -> %s", host, ips)
			return net.JoinHostPort(ips[0].String(), port), nil
		}
	}

	if r != nil {
		ips, err := r.Resolve(ctx, network, host)
		if err != nil {
			if err == resolver.ErrInvalid {
				return addr, nil
			}
			log.Error(err)
		}
		if len(ips) == 0 {
			return "", fmt.Errorf("resolver: domain %s does not exist", host)
		}
		return net.JoinHostPort(ips[0].String(), port), nil
	}
	return addr, nil
}
