package tls

import (
	md "github.com/168yy/netx/core/metadata"
	mdutil "github.com/168yy/netx/core/metadata/util"
)

type metadata struct {
	mptcp bool
}

func (l *obfsListener) parseMetadata(md md.IMetaData) (err error) {
	l.md.mptcp = mdutil.GetBool(md, "mptcp")
	return
}
