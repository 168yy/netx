module github.com/168yy/netx/x

go 1.22

toolchain go1.22.4

replace (
	github.com/168yy/netx/core => ../core
	github.com/168yy/netx/gosocks4 => ../gosocks4
	github.com/168yy/netx/gosocks5 => ../gosocks5
	github.com/168yy/netx/plugin => ../plugin
	github.com/168yy/netx/relay => ../relay
	github.com/168yy/netx/tls-dissector => ../tls-dissector
)

require (
	github.com/alecthomas/units v0.0.0-20231202071711-9a357b53e9c9
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/gin-contrib/cors v1.7.0
	github.com/gin-gonic/gin v1.9.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gobwas/glob v0.2.3
	github.com/golang/snappy v0.0.4
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.1
	github.com/168yy/netx/core v0.0.10
	github.com/168yy/netx/gosocks4 v0.0.1
	github.com/168yy/netx/gosocks5 v0.0.2
	github.com/168yy/netx/plugin v0.0.6
	github.com/168yy/netx/relay v0.0.2
	github.com/168yy/netx/tls-dissector v0.0.1
	github.com/miekg/dns v1.1.61
	github.com/mitchellh/go-homedir v1.1.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pion/dtls/v2 v2.2.6
	github.com/pires/go-proxyproto v0.7.0
	github.com/prometheus/client_golang v1.19.1
	github.com/quic-go/quic-go v0.45.0
	github.com/quic-go/webtransport-go v0.8.0
	github.com/rs/xid v1.3.0
	github.com/shadowsocks/go-shadowsocks2 v0.1.5
	github.com/shadowsocks/shadowsocks-go v0.0.0-20200409064450-3e585ff90601
	github.com/sirupsen/logrus v1.9.3
	github.com/songgao/water v0.0.0-20200317203138-2b4b6d7c09d8
	github.com/spf13/viper v1.19.0
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.4
	github.com/xtaci/kcp-go/v5 v5.6.8
	github.com/xtaci/smux v1.5.24
	github.com/xtaci/tcpraw v1.2.25
	github.com/yl2chen/cidranger v1.0.2
	github.com/zalando/go-keyring v0.2.4
	golang.org/x/crypto v0.24.0
	golang.org/x/net v0.26.0
	golang.org/x/sys v0.21.0
	golang.org/x/time v0.5.0
	golang.zx2c4.com/wireguard v0.0.0-20220703234212-c31a7b1ab478
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.2
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/aead/chacha20 v0.0.0-20180709150244-8b13a72661da // indirect
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.11.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.1 // indirect
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/coreos/go-iptables v0.7.0 // indirect
	github.com/danieljoos/wincred v1.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.19.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogf/gf/v2 v2.7.2 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/pprof v0.0.0-20230821062121-407c9e7a662f // indirect
	github.com/grokify/html-strip-tags-go v0.1.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/168yy/gfbot v0.1.18 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/klauspost/reedsolomon v1.12.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/onsi/ginkgo/v2 v2.12.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/transport/v2 v2.2.4 // indirect
	github.com/pion/udp/v2 v2.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/riobard/go-bloom v0.0.0-20200614022211-cdc8013cb5b3 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/templexxx/cpu v0.1.0 // indirect
	github.com/templexxx/xorsimd v0.4.2 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/sdk v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/mock v0.4.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/arch v0.7.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20211104114900-415007cec224 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240610135401-a8a62080eff3 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)
