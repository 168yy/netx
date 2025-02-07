package bot

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/core/service"
	"github.com/168yy/netx/x/app"
	admission_parser "github.com/168yy/netx/x/config/parsing/admission"
	auth_parser "github.com/168yy/netx/x/config/parsing/auth"
	bypass_parser "github.com/168yy/netx/x/config/parsing/bypass"
	chain_parser "github.com/168yy/netx/x/config/parsing/chain"
	hop_parser "github.com/168yy/netx/x/config/parsing/hop"
	hosts_parser "github.com/168yy/netx/x/config/parsing/hosts"
	ingress_parser "github.com/168yy/netx/x/config/parsing/ingress"
	limiter_parser "github.com/168yy/netx/x/config/parsing/limiter"
	logger_parser "github.com/168yy/netx/x/config/parsing/logger"
	recorder_parser "github.com/168yy/netx/x/config/parsing/recorder"
	resolver_parser "github.com/168yy/netx/x/config/parsing/resolver"
	router_parser "github.com/168yy/netx/x/config/parsing/router"
	sd_parser "github.com/168yy/netx/x/config/parsing/sd"
	service_parser "github.com/168yy/netx/x/config/parsing/service"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	mdutil "github.com/168yy/netx/core/metadata/util"
	"github.com/168yy/netx/x/config"
	"github.com/168yy/netx/x/limiter/conn"
	"github.com/168yy/netx/x/limiter/traffic"
	mdx "github.com/168yy/netx/x/metadata"
)

var (
	ErrInvalidCmd  = errors.New("invalid cmd")
	ErrInvalidNode = errors.New("invalid node")
)

type stringList []string

func (l *stringList) String() string {
	return fmt.Sprintf("%s", *l)
}

func (l *stringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func buildConfigFromCmd(services, nodes stringList) (*config.Config, error) {
	namePrefix := ""
	cfg := &config.Config{}

	var chain *config.ChainConfig
	if len(nodes) > 0 {
		chain = &config.ChainConfig{
			Name: fmt.Sprintf("%schain-0", namePrefix),
		}
		cfg.Chains = append(cfg.Chains, chain)
	}

	for i, node := range nodes {
		url, err := normCmd(node)
		if err != nil {
			return nil, err
		}

		nodeConfig, err := buildNodeConfig(url)
		if err != nil {
			return nil, err
		}
		nodeConfig.Name = fmt.Sprintf("%snode-0", namePrefix)

		var nodes []*config.NodeConfig
		for _, host := range strings.Split(nodeConfig.Addr, ",") {
			if host == "" {
				continue
			}
			nodeCfg := &config.NodeConfig{}
			*nodeCfg = *nodeConfig
			nodeCfg.Name = fmt.Sprintf("%snode-%d", namePrefix, len(nodes))
			nodeCfg.Addr = host
			nodes = append(nodes, nodeCfg)
		}

		mc := nodeConfig.Connector.Metadata
		md := mdx.NewMetadata(mc)

		hopConfig := &config.HopConfig{
			Name:     fmt.Sprintf("%shop-%d", namePrefix, i),
			Selector: parseSelector(mc),
			Nodes:    nodes,
		}

		if v := mdutil.GetString(md, "bypass"); v != "" {
			bypassCfg := &config.BypassConfig{
				Name: fmt.Sprintf("%sbypass-%d", namePrefix, len(cfg.Bypasses)),
			}
			if v[0] == '~' {
				bypassCfg.Whitelist = true
				v = v[1:]
			}
			for _, s := range strings.Split(v, ",") {
				if s == "" {
					continue
				}
				bypassCfg.Matchers = append(bypassCfg.Matchers, s)
			}
			hopConfig.Bypass = bypassCfg.Name
			cfg.Bypasses = append(cfg.Bypasses, bypassCfg)
			delete(mc, "bypass")
		}
		if v := mdutil.GetString(md, "resolver"); v != "" {
			resolverCfg := &config.ResolverConfig{
				Name: fmt.Sprintf("%sresolver-%d", namePrefix, len(cfg.Resolvers)),
			}
			for _, rs := range strings.Split(v, ",") {
				if rs == "" {
					continue
				}
				resolverCfg.Nameservers = append(
					resolverCfg.Nameservers,
					&config.NameserverConfig{
						Addr: rs,
					},
				)
			}
			hopConfig.Resolver = resolverCfg.Name
			cfg.Resolvers = append(cfg.Resolvers, resolverCfg)
			delete(mc, "resolver")
		}
		if v := mdutil.GetString(md, "hosts"); v != "" {
			hostsCfg := &config.HostsConfig{
				Name: fmt.Sprintf("%shosts-%d", namePrefix, len(cfg.Hosts)),
			}
			for _, s := range strings.Split(v, ",") {
				ss := strings.SplitN(s, ":", 2)
				if len(ss) != 2 {
					continue
				}
				hostsCfg.Mappings = append(
					hostsCfg.Mappings,
					&config.HostMappingConfig{
						Hostname: ss[0],
						IP:       ss[1],
					},
				)
			}
			hopConfig.Hosts = hostsCfg.Name
			cfg.Hosts = append(cfg.Hosts, hostsCfg)
			delete(mc, "hosts")
		}

		if v := mdutil.GetString(md, "interface"); v != "" {
			hopConfig.Interface = v
			delete(mc, "interface")
		}
		if v := mdutil.GetInt(md, "so_mark"); v > 0 {
			hopConfig.SockOpts = &config.SockOptsConfig{
				Mark: v,
			}
			delete(mc, "so_mark")
		}

		chain.Hops = append(chain.Hops, hopConfig)
	}

	for i, svc := range services {
		url, err := normCmd(svc)
		if err != nil {
			return nil, err
		}

		service, err := buildServiceConfig(url)
		if err != nil {
			return nil, err
		}
		service.Name = fmt.Sprintf("%sservice-%d", namePrefix, i)
		if chain != nil {
			if service.Listener.Type == "rtcp" || service.Listener.Type == "rudp" {
				service.Listener.Chain = chain.Name
			} else {
				service.Handler.Chain = chain.Name
			}
		}
		cfg.Services = append(cfg.Services, service)

		mh := service.Handler.Metadata
		md := mdx.NewMetadata(mh)
		if v := mdutil.GetInt(md, "retries"); v > 0 {
			service.Handler.Retries = v
			delete(mh, "retries")
		}
		if v := mdutil.GetString(md, "admission"); v != "" {
			admCfg := &config.AdmissionConfig{
				Name: fmt.Sprintf("%sadmission-%d", namePrefix, len(cfg.Admissions)),
			}
			if v[0] == '~' {
				admCfg.Whitelist = true
				v = v[1:]
			}
			for _, s := range strings.Split(v, ",") {
				if s == "" {
					continue
				}
				admCfg.Matchers = append(admCfg.Matchers, s)
			}
			service.Admission = admCfg.Name
			cfg.Admissions = append(cfg.Admissions, admCfg)
			delete(mh, "admission")
		}
		if v := mdutil.GetString(md, "bypass"); v != "" {
			bypassCfg := &config.BypassConfig{
				Name: fmt.Sprintf("%sbypass-%d", namePrefix, len(cfg.Bypasses)),
			}
			if v[0] == '~' {
				bypassCfg.Whitelist = true
				v = v[1:]
			}
			for _, s := range strings.Split(v, ",") {
				if s == "" {
					continue
				}
				bypassCfg.Matchers = append(bypassCfg.Matchers, s)
			}
			service.Bypass = bypassCfg.Name
			cfg.Bypasses = append(cfg.Bypasses, bypassCfg)
			delete(mh, "bypass")
		}
		if v := mdutil.GetString(md, "resolver"); v != "" {
			resolverCfg := &config.ResolverConfig{
				Name: fmt.Sprintf("%sresolver-%d", namePrefix, len(cfg.Resolvers)),
			}
			for _, rs := range strings.Split(v, ",") {
				if rs == "" {
					continue
				}
				resolverCfg.Nameservers = append(
					resolverCfg.Nameservers,
					&config.NameserverConfig{
						Addr:   rs,
						Prefer: mdutil.GetString(md, "prefer"),
					},
				)
			}
			service.Resolver = resolverCfg.Name
			cfg.Resolvers = append(cfg.Resolvers, resolverCfg)
			delete(mh, "resolver")
		}
		if v := mdutil.GetString(md, "hosts"); v != "" {
			hostsCfg := &config.HostsConfig{
				Name: fmt.Sprintf("%shosts-%d", namePrefix, len(cfg.Hosts)),
			}
			for _, s := range strings.Split(v, ",") {
				ss := strings.SplitN(s, ":", 2)
				if len(ss) != 2 {
					continue
				}
				hostsCfg.Mappings = append(
					hostsCfg.Mappings,
					&config.HostMappingConfig{
						Hostname: ss[0],
						IP:       ss[1],
					},
				)
			}
			service.Hosts = hostsCfg.Name
			cfg.Hosts = append(cfg.Hosts, hostsCfg)
			delete(mh, "hosts")
		}

		in := mdutil.GetString(md, "limiter.in")
		out := mdutil.GetString(md, "limiter.out")
		cin := mdutil.GetString(md, "limiter.conn.in")
		cout := mdutil.GetString(md, "limiter.conn.out")
		if in != "" || cin != "" || out != "" || cout != "" {
			limiter := &config.LimiterConfig{
				Name: fmt.Sprintf("%slimiter-%d", namePrefix, len(cfg.Limiters)),
			}
			if in != "" || out != "" {
				limiter.Limits = append(limiter.Limits,
					fmt.Sprintf("%s %s %s", traffic.GlobalLimitKey, in, out))
			}
			if cin != "" || cout != "" {
				limiter.Limits = append(limiter.Limits,
					fmt.Sprintf("%s %s %s", traffic.ConnLimitKey, cin, cout))
			}
			service.Limiter = limiter.Name
			cfg.Limiters = append(cfg.Limiters, limiter)
			delete(mh, "limiter.in")
			delete(mh, "limiter.out")
			delete(mh, "limiter.conn.in")
			delete(mh, "limiter.conn.out")
		}

		if climit := mdutil.GetInt(md, "climiter"); climit > 0 {
			limiter := &config.LimiterConfig{
				Name:   fmt.Sprintf("%sclimiter-%d", namePrefix, len(cfg.CLimiters)),
				Limits: []string{fmt.Sprintf("%s %d", conn.GlobalLimitKey, climit)},
			}
			service.CLimiter = limiter.Name
			cfg.CLimiters = append(cfg.CLimiters, limiter)
			delete(mh, "climiter")
		}

		if rlimit := mdutil.GetFloat(md, "rlimiter"); rlimit > 0 {
			limiter := &config.LimiterConfig{
				Name:   fmt.Sprintf("%srlimiter-%d", namePrefix, len(cfg.RLimiters)),
				Limits: []string{fmt.Sprintf("%s %s", conn.GlobalLimitKey, strconv.FormatFloat(rlimit, 'f', -1, 64))},
			}
			service.RLimiter = limiter.Name
			cfg.RLimiters = append(cfg.RLimiters, limiter)
			delete(mh, "rlimiter")
		}
	}

	return cfg, nil
}

func buildServiceConfig(url *url.URL) (*config.ServiceConfig, error) {
	namePrefix := ""
	if v := os.Getenv("_GOST_ID"); v != "" {
		namePrefix = fmt.Sprintf("go-%s@", v)
	}

	var handler, listener string
	schemes := strings.Split(url.Scheme, "+")
	if len(schemes) == 1 {
		handler = schemes[0]
		listener = schemes[0]
	}
	if len(schemes) == 2 {
		handler = schemes[0]
		listener = schemes[1]
	}

	svc := &config.ServiceConfig{
		Addr: url.Host,
	}
	if h := app.Runtime.HandlerRegistry().Get(handler); h == nil {
		handler = "auto"
	}
	if ln := app.Runtime.ListenerRegistry().Get(listener); ln == nil {
		listener = "tcp"
		if handler == "ssu" {
			listener = "udp"
		}
	}

	// forward mode
	if remotes := strings.Trim(url.EscapedPath(), "/"); remotes != "" {
		svc.Forwarder = &config.ForwarderConfig{
			// Targets: strings.Split(remotes, ","),
		}
		for i, addr := range strings.Split(remotes, ",") {
			svc.Forwarder.Nodes = append(svc.Forwarder.Nodes,
				&config.ForwardNodeConfig{
					Name: fmt.Sprintf("%starget-%d", namePrefix, i),
					Addr: addr,
				})
		}
		if handler != "relay" {
			if listener == "tcp" || listener == "udp" ||
				listener == "rtcp" || listener == "rudp" ||
				listener == "tun" || listener == "tap" ||
				listener == "dns" {
				handler = listener
			} else {
				handler = "forward"
			}
		}
	}

	var auth *config.AuthConfig
	if url.User != nil {
		auth = &config.AuthConfig{
			Username: url.User.Username(),
		}
		auth.Password, _ = url.User.Password()
	}

	m := map[string]any{}
	for k, v := range url.Query() {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	md := mdx.NewMetadata(m)

	if sa := mdutil.GetString(md, "auth"); sa != "" {
		au, err := parseAuthFromCmd(sa)
		if err != nil {
			return nil, err
		}
		auth = au
	}
	delete(m, "auth")

	tlsConfig := &config.TLSConfig{
		CertFile: mdutil.GetString(md, "certFile", "cert"),
		KeyFile:  mdutil.GetString(md, "keyFile", "key"),
		CAFile:   mdutil.GetString(md, "caFile", "ca"),
	}

	delete(m, "certFile")
	delete(m, "cert")
	delete(m, "keyFile")
	delete(m, "key")
	delete(m, "caFile")
	delete(m, "ca")

	if tlsConfig.CertFile == "" {
		tlsConfig = nil
	}

	if v := mdutil.GetString(md, "dns"); v != "" {
		md.Set("dns", strings.Split(v, ","))
	}

	if svc.Forwarder != nil {
		svc.Forwarder.Selector = parseSelector(m)
	}

	svc.Handler = &config.HandlerConfig{
		Type:     handler,
		Auth:     auth,
		Metadata: m,
	}
	svc.Listener = &config.ListenerConfig{
		Type:     listener,
		TLS:      tlsConfig,
		Metadata: m,
	}

	svc.Metadata = m

	if svc.Listener.Type == "ssh" || svc.Listener.Type == "sshd" {
		svc.Handler.Auth = nil
		svc.Listener.Auth = auth
	}

	return svc, nil
}

func buildNodeConfig(url *url.URL) (*config.NodeConfig, error) {
	var connector, dialer string
	schemes := strings.Split(url.Scheme, "+")
	if len(schemes) == 1 {
		connector = schemes[0]
		dialer = schemes[0]
	}
	if len(schemes) == 2 {
		connector = schemes[0]
		dialer = schemes[1]
	}

	node := &config.NodeConfig{
		Addr: url.Host,
	}

	if c := app.Runtime.ConnectorRegistry().Get(connector); c == nil {
		connector = "http"
	}
	if d := app.Runtime.DialerRegistry().Get(dialer); d == nil {
		dialer = "tcp"
		if connector == "ssu" {
			dialer = "udp"
		}
	}

	var auth *config.AuthConfig
	if url.User != nil {
		auth = &config.AuthConfig{
			Username: url.User.Username(),
		}
		auth.Password, _ = url.User.Password()
	}

	m := map[string]any{}
	for k, v := range url.Query() {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	md := mdx.NewMetadata(m)

	if sauth := mdutil.GetString(md, "auth"); sauth != "" && auth == nil {
		au, err := parseAuthFromCmd(sauth)
		if err != nil {
			return nil, err
		}
		auth = au
	}
	delete(m, "auth")

	tlsConfig := &config.TLSConfig{
		CertFile:   mdutil.GetString(md, "certFile", "cert"),
		KeyFile:    mdutil.GetString(md, "keyFile", "key"),
		CAFile:     mdutil.GetString(md, "caFile", "ca"),
		Secure:     mdutil.GetBool(md, "secure"),
		ServerName: mdutil.GetString(md, "serverName"),
	}
	if tlsConfig.ServerName == "" {
		tlsConfig.ServerName = url.Hostname()
	}

	delete(m, "certFile")
	delete(m, "cert")
	delete(m, "keyFile")
	delete(m, "key")
	delete(m, "caFile")
	delete(m, "ca")
	delete(m, "secure")
	delete(m, "serverName")

	if !tlsConfig.Secure && tlsConfig.CertFile == "" && tlsConfig.CAFile == "" {
		tlsConfig = nil
	}

	node.Connector = &config.ConnectorConfig{
		Type:     connector,
		Auth:     auth,
		Metadata: m,
	}
	node.Dialer = &config.DialerConfig{
		Type:     dialer,
		TLS:      tlsConfig,
		Metadata: m,
	}

	if node.Dialer.Type == "ssh" || node.Dialer.Type == "sshd" {
		node.Connector.Auth = nil
		node.Dialer.Auth = auth
	}

	return node, nil
}

func normCmd(s string) (*url.URL, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrInvalidCmd
	}

	if s[0] == ':' || !strings.Contains(s, "://") {
		s = "auto://" + s
	}

	url, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	if url.Scheme == "https" {
		url.Scheme = "http+tls"
	}

	return url, nil
}

func parseAuthFromCmd(sa string) (*config.AuthConfig, error) {
	v, err := base64.StdEncoding.DecodeString(sa)
	if err != nil {
		return nil, err
	}
	cs := string(v)
	n := strings.IndexByte(cs, ':')
	if n < 0 {
		return &config.AuthConfig{
			Username: cs,
		}, nil
	}

	return &config.AuthConfig{
		Username: cs[:n],
		Password: cs[n+1:],
	}, nil
}

func parseSelector(m map[string]any) *config.SelectorConfig {
	md := mdx.NewMetadata(m)
	strategy := mdutil.GetString(md, "strategy")
	maxFails := mdutil.GetInt(md, "maxFails", "max_fails")
	failTimeout := mdutil.GetDuration(md, "failTimeout", "fail_timeout")
	if strategy == "" && maxFails <= 0 && failTimeout <= 0 {
		return nil
	}
	if strategy == "" {
		strategy = "round"
	}
	if maxFails <= 0 {
		maxFails = 1
	}
	if failTimeout <= 0 {
		failTimeout = 30 * time.Second
	}

	delete(m, "strategy")
	delete(m, "maxFails")
	delete(m, "max_fails")
	delete(m, "failTimeout")
	delete(m, "fail_timeout")

	return &config.SelectorConfig{
		Strategy:    strategy,
		MaxFails:    maxFails,
		FailTimeout: failTimeout,
	}
}

func buildService(cfg *config.Config) (services []service.IService) {
	if cfg == nil {
		return
	}

	log := logger.Default()

	for _, loggerCfg := range cfg.Loggers {
		if lg := logger_parser.ParseLogger(loggerCfg); lg != nil {
			if err := app.Runtime.LoggerRegistry().Register(loggerCfg.Name, lg); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, autherCfg := range cfg.Authers {
		if auther := auth_parser.ParseAuther(autherCfg); auther != nil {
			if err := app.Runtime.AutherRegistry().Register(autherCfg.Name, auther); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, admissionCfg := range cfg.Admissions {
		if adm := admission_parser.ParseAdmission(admissionCfg); adm != nil {
			if err := app.Runtime.AdmissionRegistry().Register(admissionCfg.Name, adm); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, bypassCfg := range cfg.Bypasses {
		if bp := bypass_parser.ParseBypass(bypassCfg); bp != nil {
			if err := app.Runtime.BypassRegistry().Register(bypassCfg.Name, bp); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, resolverCfg := range cfg.Resolvers {
		r, err := resolver_parser.ParseResolver(resolverCfg)
		if err != nil {
			log.Fatal(err)
		}
		if r != nil {
			if err := app.Runtime.ResolverRegistry().Register(resolverCfg.Name, r); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, hostsCfg := range cfg.Hosts {
		if h := hosts_parser.ParseHostMapper(hostsCfg); h != nil {
			if err := app.Runtime.HostsRegistry().Register(hostsCfg.Name, h); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, ingressCfg := range cfg.Ingresses {
		if h := ingress_parser.ParseIngress(ingressCfg); h != nil {
			if err := app.Runtime.IngressRegistry().Register(ingressCfg.Name, h); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, routerCfg := range cfg.Routers {
		if h := router_parser.ParseRouter(routerCfg); h != nil {
			if err := app.Runtime.RouterRegistry().Register(routerCfg.Name, h); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, sdCfg := range cfg.SDs {
		if h := sd_parser.ParseSD(sdCfg); h != nil {
			if err := app.Runtime.SDRegistry().Register(sdCfg.Name, h); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, recorderCfg := range cfg.Recorders {
		if h := recorder_parser.ParseRecorder(recorderCfg); h != nil {
			if err := app.Runtime.RecorderRegistry().Register(recorderCfg.Name, h); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, limiterCfg := range cfg.Limiters {
		if h := limiter_parser.ParseTrafficLimiter(limiterCfg); h != nil {
			if err := app.Runtime.TrafficLimiterRegistry().Register(limiterCfg.Name, h); err != nil {
				log.Fatal(err)
			}
		}
	}
	for _, limiterCfg := range cfg.CLimiters {
		if h := limiter_parser.ParseConnLimiter(limiterCfg); h != nil {
			if err := app.Runtime.ConnLimiterRegistry().Register(limiterCfg.Name, h); err != nil {
				log.Fatal(err)
			}
		}
	}
	for _, limiterCfg := range cfg.RLimiters {
		if h := limiter_parser.ParseRateLimiter(limiterCfg); h != nil {
			if err := app.Runtime.RateLimiterRegistry().Register(limiterCfg.Name, h); err != nil {
				log.Fatal(err)
			}
		}
	}
	for _, hopCfg := range cfg.Hops {
		hop, err := hop_parser.ParseHop(hopCfg, log)
		if err != nil {
			log.Fatal(err)
		}
		if hop != nil {
			if err := app.Runtime.HopRegistry().Register(hopCfg.Name, hop); err != nil {
				log.Fatal(err)
			}
		}
	}
	for _, chainCfg := range cfg.Chains {
		c, err := chain_parser.ParseChain(chainCfg, log)
		if err != nil {
			log.Fatal(err)
		}
		if c != nil {
			if err := app.Runtime.ChainRegistry().Register(chainCfg.Name, c); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, svcCfg := range cfg.Services {
		svc, err := service_parser.ParseService(svcCfg)
		if err != nil {
			log.Fatal(err)
		}
		if svc != nil {
			if err := app.Runtime.ServiceRegistry().Register(svcCfg.Name, svc); err != nil {
				log.Fatal(err)
			}
		}
		services = append(services, svc)
	}

	return
}
