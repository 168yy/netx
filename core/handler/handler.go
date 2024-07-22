package handler

import (
	"context"
	"net"

	"github.com/168yy/netx/core/hop"
	"github.com/168yy/netx/core/metadata"
)

type IHandler interface {
	Init(metadata.IMetaData) error
	Handle(context.Context, net.Conn, ...HandleOption) error
}

type IForwarder interface {
	Forward(hop.IHop)
}

type NewHandler func(opts ...Option) IHandler
