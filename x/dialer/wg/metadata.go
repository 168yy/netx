package wg

import (
	"time"

	md "github.com/168yy/netx/core/metadata"
)

const (
	dialTimeout = "dialTimeout"
)

const (
	defaultDialTimeout = 5 * time.Second
)

type metadata struct {
	dialTimeout time.Duration
}

func (d *wgDialer) parseMetadata(md md.IMetaData) (err error) {
	return
}
