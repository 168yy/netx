package app

import (
	"github.com/168yy/netx/core/admission"
	"github.com/168yy/netx/core/auth"
	"github.com/168yy/netx/core/bypass"
	"github.com/168yy/netx/core/chain"
	"github.com/168yy/netx/core/connector"
	"github.com/168yy/netx/core/dialer"
	"github.com/168yy/netx/core/handler"
	"github.com/168yy/netx/core/hop"
	"github.com/168yy/netx/core/hosts"
	"github.com/168yy/netx/core/ingress"
	"github.com/168yy/netx/core/limiter/conn"
	"github.com/168yy/netx/core/limiter/rate"
	"github.com/168yy/netx/core/limiter/traffic"
	"github.com/168yy/netx/core/listener"
	"github.com/168yy/netx/core/logger"
	"github.com/168yy/netx/core/observer"
	"github.com/168yy/netx/core/recorder"
	reg "github.com/168yy/netx/core/registry"
	"github.com/168yy/netx/core/resolver"
	"github.com/168yy/netx/core/router"
	"github.com/168yy/netx/core/sd"
	"github.com/168yy/netx/core/service"
)

type IRuntime interface {
	AdmissionRegistry() reg.IRegistry[admission.IAdmission]
	AutherRegistry() reg.IRegistry[auth.IAuthenticator]
	BypassRegistry() reg.IRegistry[bypass.IBypass]
	ChainRegistry() reg.IRegistry[chain.IChainer]
	ConnectorRegistry() reg.IRegistry[connector.NewConnector]
	ConnLimiterRegistry() reg.IRegistry[conn.IConnLimiter]
	DialerRegistry() reg.IRegistry[dialer.NewDialer]
	HandlerRegistry() reg.IRegistry[handler.NewHandler]
	HopRegistry() reg.IRegistry[hop.IHop]
	HostsRegistry() reg.IRegistry[hosts.IHostMapper]
	IngressRegistry() reg.IRegistry[ingress.IIngress]
	ListenerRegistry() reg.IRegistry[listener.NewListener]
	RateLimiterRegistry() reg.IRegistry[rate.IRateLimiter]
	RecorderRegistry() reg.IRegistry[recorder.IRecorder]
	ResolverRegistry() reg.IRegistry[resolver.IResolver]
	RouterRegistry() reg.IRegistry[router.IRouter]
	SDRegistry() reg.IRegistry[sd.ISD]
	ObserverRegistry() reg.IRegistry[observer.IObserver]
	ServiceRegistry() reg.IRegistry[service.IService]
	LoggerRegistry() reg.IRegistry[logger.ILogger]
	TrafficLimiterRegistry() reg.IRegistry[traffic.ITrafficLimiter]
}
