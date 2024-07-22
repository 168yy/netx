package registry

import (
	"github.com/168yy/netx/core/service"
)

type ServiceRegistry struct {
	registry[service.IService]
}
