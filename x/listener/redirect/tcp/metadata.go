package tcp

import (
	mdata "github.com/168yy/netx/core/metadata"
	mdutil "github.com/168yy/netx/core/metadata/util"
)

type metadata struct {
	tproxy bool
	mptcp  bool
}

func (l *redirectListener) parseMetadata(md mdata.IMetaData) (err error) {
	const (
		tproxy = "tproxy"
	)
	l.md.tproxy = mdutil.GetBool(md, tproxy)
	l.md.mptcp = mdutil.GetBool(md, "mptcp")
	return
}
