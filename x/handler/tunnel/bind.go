package tunnel

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net"

	"github.com/168yy/netx/core/ingress"
	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/core/sd"
	"github.com/168yy/netx/relay"
	"github.com/168yy/netx/x/internal/util/mux"
	"github.com/google/uuid"
)

func (h *tunnelHandler) handleBind(ctx context.Context, conn net.Conn, network, address string, tunnelID relay.TunnelID, log logger.ILogger) (err error) {
	resp := relay.Response{
		Version: relay.Version1,
		Status:  relay.StatusOK,
	}

	uuid, err := uuid.NewRandom()
	if err != nil {
		resp.Status = relay.StatusInternalServerError
		resp.WriteTo(conn)
		return
	}
	connectorID := relay.NewConnectorID(uuid[:])
	if network == "udp" {
		connectorID = relay.NewUDPConnectorID(uuid[:])
	}
	// copy weight from tunnelID
	connectorID = connectorID.SetWeight(tunnelID.Weight())

	v := md5.Sum([]byte(tunnelID.String()))
	endpoint := hex.EncodeToString(v[:8])

	addr := address
	host, port, _ := net.SplitHostPort(addr)
	if host == "" {
		addr = net.JoinHostPort(endpoint, port)
	}

	af := &relay.AddrFeature{}
	err = af.ParseFrom(addr)
	if err != nil {
		log.Warn(err)
	}
	resp.Features = append(resp.Features, af,
		&relay.TunnelFeature{
			ID: connectorID,
		},
	)
	resp.WriteTo(conn)

	// Upgrade connection to multiplex session.
	session, err := mux.ClientSession(conn, h.md.muxCfg)
	if err != nil {
		return
	}

	h.pool.Add(tunnelID, NewConnector(connectorID, tunnelID, h.id, session, h.md.sd), h.md.tunnelTTL)
	if h.md.ingress != nil {
		h.md.ingress.SetRule(ctx, &ingress.Rule{
			Hostname: endpoint,
			Endpoint: tunnelID.String(),
		})
		if host != "" {
			h.md.ingress.SetRule(ctx, &ingress.Rule{
				Hostname: host,
				Endpoint: tunnelID.String(),
			})
		}
	}
	if h.md.sd != nil {
		err := h.md.sd.Register(ctx, &sd.Service{
			ID:      connectorID.String(),
			Name:    tunnelID.String(),
			Node:    h.id,
			Network: network,
			Address: h.md.entryPoint,
		})
		if err != nil {
			h.log.Error(err)
		}
	}

	log.Debugf("%s/%s: tunnel=%s, connector=%s, weight=%d established", addr, network, tunnelID, connectorID, connectorID.Weight())

	return
}
