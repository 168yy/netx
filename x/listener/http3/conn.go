package http3

import (
	"errors"
	"net"
	"net/http"
	"time"

	mdata "github.com/168yy/netx/core/metadata"
)

// a dummy HTTP3 server conn used by HTTP3 handler
type conn struct {
	md     mdata.IMetaData
	r      *http.Request
	w      http.ResponseWriter
	laddr  net.Addr
	raddr  net.Addr
	closed chan struct{}
}

func (c *conn) Read(b []byte) (n int, err error) {
	return 0, &net.OpError{Op: "read", Net: "http3", Source: nil, Addr: nil, Err: errors.New("read not supported")}
}

func (c *conn) Write(b []byte) (n int, err error) {
	return 0, &net.OpError{Op: "write", Net: "http3", Source: nil, Addr: nil, Err: errors.New("write not supported")}
}

func (c *conn) Close() error {
	select {
	case <-c.closed:
	default:
		close(c.closed)
	}
	return nil
}

func (c *conn) LocalAddr() net.Addr {
	return c.laddr
}

func (c *conn) RemoteAddr() net.Addr {
	return c.raddr
}

func (c *conn) SetDeadline(t time.Time) error {
	return &net.OpError{Op: "set", Net: "http3", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return &net.OpError{Op: "set", Net: "http3", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return &net.OpError{Op: "set", Net: "http3", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}

func (c *conn) Done() <-chan struct{} {
	return c.closed
}

// Metadata implements metadata.IMetaDatable interface.
func (c *conn) Metadata() mdata.IMetaData {
	return c.md
}
