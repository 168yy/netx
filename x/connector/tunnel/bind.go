package tunnel

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/jxo-me/netx/core/connector"
	"github.com/jxo-me/netx/relay"
	"github.com/jxo-me/netx/x/internal/util/mux"
)

// Bind implements connector.Binder.
func (c *tunnelConnector) Bind(ctx context.Context, conn net.Conn, network, address string, opts ...connector.BindOption) (net.Listener, error) {

	addr, cid, err := c.initTunnel(conn, network, address)
	if err != nil {
		return nil, err
	}

	endpoint := addr.String()
	if v, _, _ := net.SplitHostPort(addr.String()); v != "" {
		endpoint = v
	}
	log := c.options.Logger.WithFields(map[string]any{
		"endpoint": endpoint,
		"tunnel":   c.md.tunnelID.String(),
	})
	log.Infof("create tunnel on %s/%s OK, tunnel=%s, connector=%s", addr, network, c.md.tunnelID.String(), cid)

	session, err := mux.ServerSession(conn, c.md.muxCfg)
	if err != nil {
		return nil, err
	}

	return &bindListener{
		network: network,
		addr:    addr,
		session: session,
		logger:  log,
	}, nil
}

func (c *tunnelConnector) initTunnel(conn net.Conn, network, address string) (addr net.Addr, cid relay.ConnectorID, err error) {
	req := relay.Request{
		Version: relay.Version1,
		Cmd:     relay.CmdBind,
	}

	if network == "udp" {
		req.Cmd |= relay.FUDP
		req.Features = append(req.Features, &relay.NetworkFeature{
			Network: relay.NetworkUDP,
		})
	}

	if c.options.Auth != nil {
		pwd, _ := c.options.Auth.Password()
		req.Features = append(req.Features, &relay.UserAuthFeature{
			Username: c.options.Auth.Username(),
			Password: pwd,
		})
	}

	af := &relay.AddrFeature{}
	af.ParseFrom(conn.LocalAddr().String()) // src address
	req.Features = append(req.Features, af)

	af = &relay.AddrFeature{}
	af.ParseFrom(address)
	req.Features = append(req.Features, af) // dst address

	req.Features = append(req.Features, &relay.TunnelFeature{
		ID: c.md.tunnelID.ID(),
	})
	if _, err = req.WriteTo(conn); err != nil {
		return
	}

	// first reply, bind status
	resp := relay.Response{}
	if _, err = resp.ReadFrom(conn); err != nil {
		return
	}

	if resp.Status != relay.StatusOK {
		err = fmt.Errorf("%d: create tunnel %s failed", resp.Status, c.md.tunnelID.String())
		return
	}

	for _, f := range resp.Features {
		switch f.Type() {
		case relay.FeatureAddr:
			if feature, _ := f.(*relay.AddrFeature); feature != nil {
				addr = &bindAddr{
					network: network,
					addr:    net.JoinHostPort(feature.Host, strconv.Itoa(int(feature.Port))),
				}
			}
		case relay.FeatureTunnel:
			if feature, _ := f.(*relay.TunnelFeature); feature != nil {
				cid = relay.NewConnectorID(feature.ID[:])
			}
		}
	}

	return
}
