package boot

import (
	"github.com/168yy/netx/core/dialer"
	"github.com/168yy/netx/x/consts"
	dialerDirect "github.com/168yy/netx/x/dialer/direct"
	"github.com/168yy/netx/x/dialer/dtls"
	"github.com/168yy/netx/x/dialer/ftcp"
	"github.com/168yy/netx/x/dialer/grpc"
	dialerHttp2 "github.com/168yy/netx/x/dialer/http2"
	"github.com/168yy/netx/x/dialer/http2/h2"
	"github.com/168yy/netx/x/dialer/http3"
	"github.com/168yy/netx/x/dialer/http3/wt"
	dialerIcmp "github.com/168yy/netx/x/dialer/icmp"
	"github.com/168yy/netx/x/dialer/kcp"
	"github.com/168yy/netx/x/dialer/mtcp"
	"github.com/168yy/netx/x/dialer/mtls"
	"github.com/168yy/netx/x/dialer/mws"
	dialerObfsHttp "github.com/168yy/netx/x/dialer/obfs/http"
	dialerObfsTls "github.com/168yy/netx/x/dialer/obfs/tls"
	"github.com/168yy/netx/x/dialer/pht"
	dialerQuic "github.com/168yy/netx/x/dialer/quic"
	dialerSerial "github.com/168yy/netx/x/dialer/serial"
	"github.com/168yy/netx/x/dialer/ssh"
	dialerSshd "github.com/168yy/netx/x/dialer/sshd"
	dialerTcp "github.com/168yy/netx/x/dialer/tcp"
	dialerTls "github.com/168yy/netx/x/dialer/tls"
	dialerUdp "github.com/168yy/netx/x/dialer/udp"
	dialerUnix "github.com/168yy/netx/x/dialer/unix"
	"github.com/168yy/netx/x/dialer/wg"
	"github.com/168yy/netx/x/dialer/ws"
)

var Dialers = map[string]dialer.NewDialer{
	consts.Direct:  dialerDirect.NewDialer,
	consts.Virtual: dialerDirect.NewDialer,
	consts.Dtls:    dtls.NewDialer,
	consts.Ftcp:    ftcp.NewDialer,
	consts.Grpc:    grpc.NewDialer,
	consts.Http2:   dialerHttp2.NewDialer,
	consts.H2:      h2.NewTLSDialer,
	consts.H2c:     h2.NewDialer,
	consts.Http3:   http3.NewDialer,
	consts.H3:      http3.NewDialer,
	consts.Wt:      wt.NewDialer,
	consts.Icmp:    dialerIcmp.NewDialer,
	consts.Kcp:     kcp.NewDialer,
	consts.Mtcp:    mtcp.NewDialer,
	consts.Mtls:    mtls.NewDialer,
	consts.Mws:     mws.NewDialer,
	consts.Mwss:    mws.NewTLSDialer,
	consts.Ohttp:   dialerObfsHttp.NewDialer,
	consts.Otls:    dialerObfsTls.NewDialer,
	consts.Pht:     pht.NewDialer,
	consts.Phts:    pht.NewTLSDialer,
	consts.Quic:    dialerQuic.NewDialer,
	consts.Serial:  dialerSerial.NewDialer,
	consts.Ssh:     ssh.NewDialer,
	consts.Sshd:    dialerSshd.NewDialer,
	consts.Tcp:     dialerTcp.NewDialer,
	consts.Tls:     dialerTls.NewDialer,
	consts.Udp:     dialerUdp.NewDialer,
	consts.Unix:    dialerUnix.NewDialer,
	consts.Wg:      wg.NewDialer,
	consts.Ws:      ws.NewDialer,
	consts.Wss:     ws.NewTLSDialer,
}
