package local

import (
	"time"

	mdata "github.com/168yy/netx/core/metadata"
	mdutil "github.com/168yy/netx/core/metadata/util"
)

type metadata struct {
	readTimeout     time.Duration
	sniffing        bool
	sniffingTimeout time.Duration
}

func (h *forwardHandler) parseMetadata(md mdata.IMetaData) (err error) {
	const (
		readTimeout = "readTimeout"
		sniffing    = "sniffing"
	)

	h.md.readTimeout = mdutil.GetDuration(md, readTimeout)
	h.md.sniffing = mdutil.GetBool(md, sniffing)
	h.md.sniffingTimeout = mdutil.GetDuration(md, "sniffing.timeout")
	return
}
