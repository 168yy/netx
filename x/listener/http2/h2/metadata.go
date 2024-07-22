package h2

import (
	mdata "github.com/168yy/netx/core/metadata"
	mdutil "github.com/168yy/netx/core/metadata/util"
)

const (
	defaultBacklog = 128
)

type metadata struct {
	path    string
	backlog int
	mptcp   bool
}

func (l *h2Listener) parseMetadata(md mdata.IMetaData) (err error) {
	const (
		path    = "path"
		backlog = "backlog"
	)

	l.md.backlog = mdutil.GetInt(md, backlog)
	if l.md.backlog <= 0 {
		l.md.backlog = defaultBacklog
	}

	l.md.path = mdutil.GetString(md, path)
	l.md.mptcp = mdutil.GetBool(md, "mptcp")

	return
}
