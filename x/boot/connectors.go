package boot

import (
	"github.com/168yy/netx/core/connector"
	direct "github.com/168yy/netx/x/connector/direct"
	"github.com/168yy/netx/x/connector/forward"
	"github.com/168yy/netx/x/connector/http"
	"github.com/168yy/netx/x/connector/http2"
	"github.com/168yy/netx/x/connector/relay"
	"github.com/168yy/netx/x/connector/serial"
	"github.com/168yy/netx/x/connector/sni"
	v4 "github.com/168yy/netx/x/connector/socks/v4"
	v5 "github.com/168yy/netx/x/connector/socks/v5"
	"github.com/168yy/netx/x/connector/ss"
	ssu "github.com/168yy/netx/x/connector/ss/udp"
	"github.com/168yy/netx/x/connector/sshd"
	"github.com/168yy/netx/x/connector/tcp"
	"github.com/168yy/netx/x/connector/tunnel"
	"github.com/168yy/netx/x/connector/unix"
	"github.com/168yy/netx/x/consts"
)

var Connectors = map[string]connector.NewConnector{
	consts.Direct:  direct.NewConnector,
	consts.Virtual: direct.NewConnector,
	consts.Forward: forward.NewConnector,
	consts.Http:    http.NewConnector,
	consts.Http2:   http2.NewConnector,
	consts.Relay:   relay.NewConnector,
	consts.Serial:  serial.NewConnector,
	consts.Sni:     sni.NewConnector,
	consts.Socks4:  v4.NewConnector,
	consts.Socks4a: v4.NewConnector,
	consts.Socks5:  v5.NewConnector,
	consts.Socks:   v5.NewConnector,
	consts.Ss:      ss.NewConnector,
	consts.Ssu:     ssu.NewConnector,
	consts.Sshd:    sshd.NewConnector,
	consts.Tcp:     tcp.NewConnector,
	consts.Tunnel:  tunnel.NewConnector,
	consts.Unix:    unix.NewConnector,
}
