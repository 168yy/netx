package boot

import (
	"github.com/168yy/netx/core/listener"
	"github.com/168yy/netx/x/consts"
	listenerDns "github.com/168yy/netx/x/listener/dns"
	listenerDtls "github.com/168yy/netx/x/listener/dtls"
	listenerFtcp "github.com/168yy/netx/x/listener/ftcp"
	listenerGrpc "github.com/168yy/netx/x/listener/grpc"
	listenerHttp2 "github.com/168yy/netx/x/listener/http2"
	listenerHttpH2 "github.com/168yy/netx/x/listener/http2/h2"
	listenerHttp3 "github.com/168yy/netx/x/listener/http3"
	listenerHttpH3 "github.com/168yy/netx/x/listener/http3/h3"
	listenerHttpWt "github.com/168yy/netx/x/listener/http3/wt"
	listenerIcmp "github.com/168yy/netx/x/listener/icmp"
	listenerKcp "github.com/168yy/netx/x/listener/kcp"
	listenerMtcp "github.com/168yy/netx/x/listener/mtcp"
	listenerMtls "github.com/168yy/netx/x/listener/mtls"
	listenerMws "github.com/168yy/netx/x/listener/mws"
	listenerObfsHttp "github.com/168yy/netx/x/listener/obfs/http"
	listenerObfsTls "github.com/168yy/netx/x/listener/obfs/tls"
	listenerPht "github.com/168yy/netx/x/listener/pht"
	listenerQuic "github.com/168yy/netx/x/listener/quic"
	listenerRedirectTcp "github.com/168yy/netx/x/listener/redirect/tcp"
	listenerRedirectUdp "github.com/168yy/netx/x/listener/redirect/udp"
	listenerRtcp "github.com/168yy/netx/x/listener/rtcp"
	listenerRudp "github.com/168yy/netx/x/listener/rudp"
	listenerSerial "github.com/168yy/netx/x/listener/serial"
	listenerSsh "github.com/168yy/netx/x/listener/ssh"
	listenerSshd "github.com/168yy/netx/x/listener/sshd"
	listenerTap "github.com/168yy/netx/x/listener/tap"
	listenerTcp "github.com/168yy/netx/x/listener/tcp"
	listenerTls "github.com/168yy/netx/x/listener/tls"
	listenerTun "github.com/168yy/netx/x/listener/tun"
	listenerUdp "github.com/168yy/netx/x/listener/udp"
	listenerUnix "github.com/168yy/netx/x/listener/unix"
	listenerWs "github.com/168yy/netx/x/listener/ws"
)

var Listeners = map[string]listener.NewListener{
	consts.Dns:      listenerDns.NewListener,
	consts.Dtls:     listenerDtls.NewListener,
	consts.Ftcp:     listenerFtcp.NewListener,
	consts.Grpc:     listenerGrpc.NewListener,
	consts.Http2:    listenerHttp2.NewListener,
	consts.H2c:      listenerHttpH2.NewListener,
	consts.H2:       listenerHttpH2.NewTLSListener,
	consts.Http3:    listenerHttp3.NewListener,
	consts.H3:       listenerHttpH3.NewListener,
	consts.Wt:       listenerHttpWt.NewListener,
	consts.Icmp:     listenerIcmp.NewListener,
	consts.Kcp:      listenerKcp.NewListener,
	consts.Mtcp:     listenerMtcp.NewListener,
	consts.Mtls:     listenerMtls.NewListener,
	consts.Mws:      listenerMws.NewListener,
	consts.Mwss:     listenerMws.NewTLSListener,
	consts.Ohttp:    listenerObfsHttp.NewListener,
	consts.Otls:     listenerObfsTls.NewListener,
	consts.Pht:      listenerPht.NewListener,
	consts.Phts:     listenerPht.NewTLSListener,
	consts.Quic:     listenerQuic.NewListener,
	consts.Red:      listenerRedirectTcp.NewListener,
	consts.Redir:    listenerRedirectTcp.NewListener,
	consts.Redirect: listenerRedirectTcp.NewListener,
	consts.Redu:     listenerRedirectUdp.NewListener,
	consts.Rtcp:     listenerRtcp.NewListener,
	consts.Rudp:     listenerRudp.NewListener,
	consts.Serial:   listenerSerial.NewListener,
	consts.Ssh:      listenerSsh.NewListener,
	consts.Sshd:     listenerSshd.NewListener,
	consts.Tap:      listenerTap.NewListener,
	consts.Tcp:      listenerTcp.NewListener,
	consts.Tls:      listenerTls.NewListener,
	consts.Tun:      listenerTun.NewListener,
	consts.Udp:      listenerUdp.NewListener,
	consts.Unix:     listenerUnix.NewListener,
	consts.Ws:       listenerWs.NewListener,
	consts.Wss:      listenerWs.NewTLSListener,
}
