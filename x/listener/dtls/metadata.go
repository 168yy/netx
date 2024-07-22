package dtls

import (
	"time"

	mdata "github.com/168yy/netx/core/metadata"
	mdutil "github.com/168yy/netx/core/metadata/util"
)

const (
	defaultBufferSize = 1200
)

type metadata struct {
	mtu            int
	bufferSize     int
	flightInterval time.Duration
}

func (l *dtlsListener) parseMetadata(md mdata.IMetaData) (err error) {
	l.md.mtu = mdutil.GetInt(md, "dtls.mtu", "mtu")
	l.md.bufferSize = mdutil.GetInt(md, "dtls.bufferSize", "bufferSize")
	if l.md.bufferSize <= 0 {
		l.md.bufferSize = defaultBufferSize
	}

	l.md.flightInterval = mdutil.GetDuration(md, "dtls.flightInterval", "flightInterval")

	return nil
}
