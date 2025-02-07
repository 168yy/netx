package quic

import (
	"time"

	mdata "github.com/168yy/netx/core/metadata"
	mdutil "github.com/168yy/netx/core/metadata/util"
)

const (
	defaultBacklog = 128
)

type metadata struct {
	// keepAlive        bool
	keepAlivePeriod  time.Duration
	handshakeTimeout time.Duration
	maxIdleTimeout   time.Duration
	maxStreams       int

	cipherKey []byte
	backlog   int
}

func (l *quicListener) parseMetadata(md mdata.IMetaData) (err error) {
	const (
		keepAlive        = "keepAlive"
		keepAlivePeriod  = "ttl"
		handshakeTimeout = "handshakeTimeout"
		maxIdleTimeout   = "maxIdleTimeout"
		maxStreams       = "maxStreams"

		backlog   = "backlog"
		cipherKey = "cipherKey"
	)

	l.md.backlog = mdutil.GetInt(md, backlog)
	if l.md.backlog <= 0 {
		l.md.backlog = defaultBacklog
	}

	if key := mdutil.GetString(md, cipherKey); key != "" {
		l.md.cipherKey = []byte(key)
	}

	if mdutil.GetBool(md, keepAlive) {
		l.md.keepAlivePeriod = mdutil.GetDuration(md, keepAlivePeriod)
		if l.md.keepAlivePeriod <= 0 {
			l.md.keepAlivePeriod = 10 * time.Second
		}
	}
	l.md.handshakeTimeout = mdutil.GetDuration(md, handshakeTimeout)
	l.md.maxIdleTimeout = mdutil.GetDuration(md, maxIdleTimeout)
	l.md.maxStreams = mdutil.GetInt(md, maxStreams)

	return
}
