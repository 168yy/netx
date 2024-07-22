package registry

import (
	"github.com/168yy/netx/core/listener"
	"github.com/168yy/netx/core/logger"
)

type ListenerRegistry struct {
	registry[listener.NewListener]
}

func (r *ListenerRegistry) Register(name string, v listener.NewListener) error {
	if err := r.registry.Register(name, v); err != nil {
		logger.Default().Fatal(err)
	}
	return nil
}
