//go:build !linux

package router

import (
	"github.com/168yy/netx/core/router"
)

func (*localRouter) setSysRoutes(routes ...*router.Route) error {
	return nil
}
