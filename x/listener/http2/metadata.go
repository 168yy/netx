package http2

import (
	mdata "github.com/168yy/netx/core/metadata"
	mdutil "github.com/168yy/netx/core/metadata/util"
)

const (
	defaultBacklog = 128
)

type metadata struct {
	backlog int
	mptcp   bool
}

func (l *http2Listener) parseMetadata(md mdata.IMetaData) (err error) {
	const (
		backlog = "backlog"
	)

	l.md.backlog = mdutil.GetInt(md, backlog)
	if l.md.backlog <= 0 {
		l.md.backlog = defaultBacklog
	}
	l.md.mptcp = mdutil.GetBool(md, "mptcp")

	return
}
