package wrapper

import (
	"net"

	limiter "github.com/168yy/netx/core/limiter/conn"
)

type listener struct {
	net.Listener
	limiter limiter.IConnLimiter
}

func WrapListener(limiter limiter.IConnLimiter, ln net.Listener) net.Listener {
	if limiter == nil {
		return ln
	}

	return &listener{
		limiter:  limiter,
		Listener: ln,
	}
}

func (ln *listener) Accept() (net.Conn, error) {
	c, err := ln.Listener.Accept()
	if err != nil {
		return nil, err
	}

	host, _, _ := net.SplitHostPort(c.RemoteAddr().String())
	if lim := ln.limiter.Limiter(host); lim != nil {
		if lim.Allow(1) {
			return WrapConn(lim, c), nil
		}
		c.Close()
	}

	return c, nil
}
