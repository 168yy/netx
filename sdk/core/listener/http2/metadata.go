package http2

import (
	mdata "github.com/jxo-me/netx/sdk/core/metadata"
	mdutil "github.com/jxo-me/netx/sdk/core/metadata/util"
)

const (
	defaultBacklog = 128
)

type metadata struct {
	backlog int
}

func (l *http2Listener) parseMetadata(md mdata.IMetaData) (err error) {
	const (
		backlog = "backlog"
	)

	l.md.backlog = mdutil.GetInt(md, backlog)
	if l.md.backlog <= 0 {
		l.md.backlog = defaultBacklog
	}
	return
}
