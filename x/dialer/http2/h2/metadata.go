package h2

import (
	"net/http"

	mdata "github.com/168yy/netx/core/metadata"
	mdutil "github.com/168yy/netx/core/metadata/util"
)

type metadata struct {
	host   string
	path   string
	header http.Header
}

func (d *h2Dialer) parseMetadata(md mdata.IMetaData) (err error) {
	const (
		host   = "host"
		path   = "path"
		header = "header"
	)

	d.md.host = mdutil.GetString(md, host)
	d.md.path = mdutil.GetString(md, path)
	if m := mdutil.GetStringMapString(md, header); len(m) > 0 {
		h := http.Header{}
		for k, v := range m {
			h.Add(k, v)
		}
		d.md.header = h
	}
	return
}
