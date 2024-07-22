package tls

import (
	mdata "github.com/168yy/netx/core/metadata"
	mdutil "github.com/168yy/netx/core/metadata/util"
)

type metadata struct {
	host string
}

func (d *obfsTLSDialer) parseMetadata(md mdata.IMetaData) (err error) {
	const (
		host = "host"
	)

	d.md.host = mdutil.GetString(md, host)
	return
}
