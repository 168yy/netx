package boot

import (
	"github.com/168yy/netx/core/handler"
	"github.com/168yy/netx/x/consts"
	"github.com/168yy/netx/x/handler/auto"
	"github.com/168yy/netx/x/handler/dns"
	"github.com/168yy/netx/x/handler/file"
	"github.com/168yy/netx/x/handler/forward/local"
	"github.com/168yy/netx/x/handler/forward/remote"
	handlerHttp "github.com/168yy/netx/x/handler/http"
	handlerHttp2 "github.com/168yy/netx/x/handler/http2"
	handlerHttp3 "github.com/168yy/netx/x/handler/http3"
	"github.com/168yy/netx/x/handler/metrics"
	redirect "github.com/168yy/netx/x/handler/redirect/tcp"
	redirectUdp "github.com/168yy/netx/x/handler/redirect/udp"
	handlerRelay "github.com/168yy/netx/x/handler/relay"
	handlerSerial "github.com/168yy/netx/x/handler/serial"
	handlerSni "github.com/168yy/netx/x/handler/sni"
	handlerSocksV4 "github.com/168yy/netx/x/handler/socks/v4"
	handlerSocksV5 "github.com/168yy/netx/x/handler/socks/v5"
	handlerSs "github.com/168yy/netx/x/handler/ss"
	handlerSsUdp "github.com/168yy/netx/x/handler/ss/udp"
	handlerSshd "github.com/168yy/netx/x/handler/sshd"
	"github.com/168yy/netx/x/handler/tap"
	"github.com/168yy/netx/x/handler/tun"
	"github.com/168yy/netx/x/handler/tunnel"
	"github.com/168yy/netx/x/handler/unix"
)

var Handlers = map[string]handler.NewHandler{
	consts.Auto:     auto.NewHandler,
	consts.Dns:      dns.NewHandler,
	consts.File:     file.NewHandler,
	consts.Tcp:      local.NewHandler,
	consts.Udp:      local.NewHandler,
	consts.Forward:  local.NewHandler,
	consts.Rtcp:     remote.NewHandler,
	consts.Rudp:     remote.NewHandler,
	consts.Http:     handlerHttp.NewHandler,
	consts.Http2:    handlerHttp2.NewHandler,
	consts.Http3:    handlerHttp3.NewHandler,
	consts.Metrics:  metrics.NewHandler,
	consts.Red:      redirect.NewHandler,
	consts.Redir:    redirect.NewHandler,
	consts.Redirect: redirect.NewHandler,
	consts.Redu:     redirectUdp.NewHandler,
	consts.Relay:    handlerRelay.NewHandler,
	consts.Serial:   handlerSerial.NewHandler,
	consts.Sni:      handlerSni.NewHandler,
	consts.Socks4:   handlerSocksV4.NewHandler,
	consts.Socks4a:  handlerSocksV4.NewHandler,
	consts.Socks5:   handlerSocksV5.NewHandler,
	consts.Socks:    handlerSocksV5.NewHandler,
	consts.Ss:       handlerSs.NewHandler,
	consts.Ssu:      handlerSsUdp.NewHandler,
	consts.Sshd:     handlerSshd.NewHandler,
	consts.Tap:      tap.NewHandler,
	consts.Tun:      tun.NewHandler,
	consts.Tunnel:   tunnel.NewHandler,
	consts.Unix:     unix.NewHandler,
}
