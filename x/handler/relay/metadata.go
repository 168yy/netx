package relay

import (
	"math"
	"time"

	mdata "github.com/168yy/netx/core/metadata"
	mdutil "github.com/168yy/netx/core/metadata/util"
	"github.com/168yy/netx/x/internal/util/mux"
)

type metadata struct {
	readTimeout   time.Duration
	enableBind    bool
	udpBufferSize int
	noDelay       bool
	hash          string
	muxCfg        *mux.Config
	observePeriod time.Duration
}

func (h *relayHandler) parseMetadata(md mdata.IMetaData) (err error) {
	h.md.readTimeout = mdutil.GetDuration(md, "readTimeout")
	h.md.enableBind = mdutil.GetBool(md, "bind")
	h.md.noDelay = mdutil.GetBool(md, "nodelay")

	if bs := mdutil.GetInt(md, "udpBufferSize"); bs > 0 {
		h.md.udpBufferSize = int(math.Min(math.Max(float64(bs), 512), 64*1024))
	} else {
		h.md.udpBufferSize = 4096
	}

	h.md.hash = mdutil.GetString(md, "hash")

	h.md.muxCfg = &mux.Config{
		Version:           mdutil.GetInt(md, "mux.version"),
		KeepAliveInterval: mdutil.GetDuration(md, "mux.keepaliveInterval"),
		KeepAliveDisabled: mdutil.GetBool(md, "mux.keepaliveDisabled"),
		KeepAliveTimeout:  mdutil.GetDuration(md, "mux.keepaliveTimeout"),
		MaxFrameSize:      mdutil.GetInt(md, "mux.maxFrameSize"),
		MaxReceiveBuffer:  mdutil.GetInt(md, "mux.maxReceiveBuffer"),
		MaxStreamBuffer:   mdutil.GetInt(md, "mux.maxStreamBuffer"),
	}

	h.md.observePeriod = mdutil.GetDuration(md, "observePeriod")

	return
}
