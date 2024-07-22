package registry

import (
	"github.com/168yy/netx/core/dialer"
	"github.com/168yy/netx/core/logger"
)

type DialerRegistry struct {
	registry[dialer.NewDialer]
}

func (r *DialerRegistry) Register(name string, v dialer.NewDialer) error {
	if err := r.registry.Register(name, v); err != nil {
		logger.Default().Fatal(err)
	}
	return nil
}
