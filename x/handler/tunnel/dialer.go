package tunnel

import (
	"context"
	"net"
	"time"

	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/core/sd"
)

type Dialer struct {
	node    string
	pool    *ConnectorPool
	sd      sd.ISD
	retry   int
	timeout time.Duration
	log     logger.ILogger
}

func (d *Dialer) Dial(ctx context.Context, network string, tid string) (conn net.Conn, node string, cid string, err error) {
	retry := d.retry
	if retry <= 0 {
		retry = 1
	}

	for i := 0; i < retry; i++ {
		c := d.pool.Get(network, tid)
		if c == nil {
			break
		}

		conn, err = c.Session().GetConn()
		if err != nil {
			d.log.Error(err)
			continue
		}
		node = d.node
		cid = c.id.String()

		break
	}
	if conn != nil || err != nil {
		return
	}

	if d.sd == nil {
		err = ErrTunnelNotAvailable
		return
	}

	ss, err := d.sd.Get(ctx, tid)
	if err != nil {
		return
	}

	var service *sd.Service
	for _, s := range ss {
		d.log.Debugf("%+v", s)
		if s.Name != d.node && s.Network == network {
			service = s
			break
		}
	}
	if service == nil || service.Address == "" {
		err = ErrTunnelNotAvailable
		return
	}

	node = service.Node
	cid = service.Name

	dialer := net.Dialer{
		Timeout: d.timeout,
	}
	conn, err = dialer.DialContext(ctx, network, service.Address)
	return
}
