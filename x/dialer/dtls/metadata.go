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

func (d *dtlsDialer) parseMetadata(md mdata.IMetaData) (err error) {
	d.md.mtu = mdutil.GetInt(md, "dtls.mtu", "mtu")
	d.md.bufferSize = mdutil.GetInt(md, "dtls.bufferSize", "bufferSize")
	if d.md.bufferSize <= 0 {
		d.md.bufferSize = defaultBufferSize
	}
	d.md.flightInterval = mdutil.GetDuration(md, "dtls.flightInterval", "flightInterval")
	return
}
