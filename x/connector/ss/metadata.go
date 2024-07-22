package ss

import (
	"time"

	mdata "github.com/168yy/netx/core/metadata"
	mdutil "github.com/168yy/netx/core/metadata/util"
)

type metadata struct {
	key            string
	connectTimeout time.Duration
	noDelay        bool
}

func (c *ssConnector) parseMetadata(md mdata.IMetaData) (err error) {
	const (
		key            = "key"
		connectTimeout = "timeout"
		noDelay        = "nodelay"
	)

	c.md.key = mdutil.GetString(md, key)
	c.md.connectTimeout = mdutil.GetDuration(md, connectTimeout)
	c.md.noDelay = mdutil.GetBool(md, noDelay)

	return
}
