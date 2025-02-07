package wrapper

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"syscall"
	"time"

	limiter "github.com/168yy/netx/core/limiter/traffic"
	"github.com/168yy/netx/core/metadata"
	xnet "github.com/168yy/netx/x/internal/net"
	"github.com/168yy/netx/x/internal/net/udp"
	"github.com/patrickmn/go-cache"
)

var (
	errUnsupport = errors.New("unsupported operation")
)

// serverConn is a server side Conn with traffic limiter supported.
type serverConn struct {
	net.Conn
	rbuf       bytes.Buffer
	limiter    limiter.ITrafficLimiter
	limiterIn  limiter.ILimiter
	limiterOut limiter.ILimiter
	expIn      int64
	expOut     int64
	opts       []limiter.Option
}

func WrapConn(tlimiter limiter.ITrafficLimiter, c net.Conn) net.Conn {
	if tlimiter == nil {
		return c
	}
	return &serverConn{
		Conn:    c,
		limiter: tlimiter,
		opts: []limiter.Option{
			limiter.NetworkOption(c.LocalAddr().Network()),
			limiter.SrcOption(c.RemoteAddr().String()),
			limiter.AddrOption(c.LocalAddr().String()),
		},
	}
}

func (c *serverConn) getInLimiter() limiter.ILimiter {
	now := time.Now().UnixNano()
	// cache the limiter for 60s
	if c.limiter != nil && time.Duration(now-c.expIn) > 60*time.Second {
		if lim := c.limiter.In(context.Background(), c.RemoteAddr().String()); lim != nil {
			c.limiterIn = lim
		}
		c.expIn = now
	}
	return c.limiterIn
}

func (c *serverConn) getOutLimiter() limiter.ILimiter {
	now := time.Now().UnixNano()
	// cache the limiter for 60s
	if c.limiter != nil && time.Duration(now-c.expOut) > 60*time.Second {
		if lim := c.limiter.Out(context.Background(), c.RemoteAddr().String()); lim != nil {
			c.limiterOut = lim
		}
		c.expOut = now
	}
	return c.limiterOut
}

func (c *serverConn) Read(b []byte) (n int, err error) {
	limiter := c.getInLimiter()
	if limiter == nil {
		return c.Conn.Read(b)
	}

	if c.rbuf.Len() > 0 {
		burst := len(b)
		if c.rbuf.Len() < burst {
			burst = c.rbuf.Len()
		}
		lim := limiter.Wait(context.Background(), burst)
		return c.rbuf.Read(b[:lim])
	}

	nn, err := c.Conn.Read(b)
	if err != nil {
		return nn, err
	}

	n = limiter.Wait(context.Background(), nn)
	if n < nn {
		if _, err = c.rbuf.Write(b[n:nn]); err != nil {
			return 0, err
		}
	}

	return
}

func (c *serverConn) Write(b []byte) (n int, err error) {
	limiter := c.getOutLimiter()
	if limiter == nil {
		return c.Conn.Write(b)
	}

	nn := 0
	for len(b) > 0 {
		nn, err = c.Conn.Write(b[:limiter.Wait(context.Background(), len(b))])
		n += nn
		if err != nil {
			return
		}
		b = b[nn:]
	}

	return
}

func (c *serverConn) SyscallConn() (rc syscall.RawConn, err error) {
	if sc, ok := c.Conn.(syscall.Conn); ok {
		rc, err = sc.SyscallConn()
		return
	}
	err = errUnsupport
	return
}

func (c *serverConn) Metadata() metadata.IMetaData {
	if md, ok := c.Conn.(metadata.IMetaDatable); ok {
		return md.Metadata()
	}
	return nil
}

type packetConn struct {
	net.PacketConn
	limiter   limiter.ITrafficLimiter
	inLimits  *cache.Cache
	outLimits *cache.Cache
}

func WrapPacketConn(lim limiter.ITrafficLimiter, pc net.PacketConn) net.PacketConn {
	if lim == nil {
		return pc
	}
	return &packetConn{
		PacketConn: pc,
		limiter:    lim,
		inLimits:   cache.New(time.Second, 10*time.Second),
		outLimits:  cache.New(time.Second, 10*time.Second),
	}
}

func (c *packetConn) getInLimiter(addr net.Addr) limiter.ILimiter {
	if c.limiter == nil {
		return nil
	}

	lim, ok := func() (lim limiter.ILimiter, ok bool) {
		v, ok := c.inLimits.Get(addr.String())
		if ok {
			if v != nil {
				lim = v.(limiter.ILimiter)
			}
		}
		return
	}()
	if ok {
		return lim
	}

	lim = c.limiter.In(context.Background(), addr.String())
	c.inLimits.Set(addr.String(), lim, 0)

	return lim
}

func (c *packetConn) getOutLimiter(addr net.Addr) limiter.ILimiter {
	if c.limiter == nil {
		return nil
	}

	lim, ok := func() (lim limiter.ILimiter, ok bool) {
		v, ok := c.outLimits.Get(addr.String())
		if ok {
			if v != nil {
				lim = v.(limiter.ILimiter)
			}
		}
		return
	}()
	if ok {
		return lim
	}

	lim = c.limiter.Out(context.Background(), addr.String())
	c.outLimits.Set(addr.String(), lim, 0)

	return lim
}

func (c *packetConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	for {
		n, addr, err = c.PacketConn.ReadFrom(p)
		if err != nil {
			return
		}

		limiter := c.getInLimiter(addr)
		if limiter == nil {
			return
		}

		// discard when exceed the limit size.
		if limiter.Wait(context.Background(), n) < n {
			continue
		}

		return
	}
}

func (c *packetConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	// discard when exceed the limit size.
	if limiter := c.getOutLimiter(addr); limiter != nil &&
		limiter.Wait(context.Background(), len(p)) < len(p) {
		n = len(p)
		return
	}

	return c.PacketConn.WriteTo(p, addr)
}

func (c *packetConn) Metadata() metadata.IMetaData {
	if md, ok := c.PacketConn.(metadata.IMetaDatable); ok {
		return md.Metadata()
	}
	return nil
}

type udpConn struct {
	net.PacketConn
	limiter   limiter.ITrafficLimiter
	inLimits  *cache.Cache
	outLimits *cache.Cache
}

func WrapUDPConn(limiter limiter.ITrafficLimiter, pc net.PacketConn) udp.Conn {
	return &udpConn{
		PacketConn: pc,
		limiter:    limiter,
		inLimits:   cache.New(time.Second, 10*time.Second),
		outLimits:  cache.New(time.Second, 10*time.Second),
	}
}

func (c *udpConn) getInLimiter(addr net.Addr) limiter.ILimiter {
	if c.limiter == nil {
		return nil
	}

	lim, ok := func() (lim limiter.ILimiter, ok bool) {
		v, ok := c.inLimits.Get(addr.String())
		if ok {
			if v != nil {
				lim = v.(limiter.ILimiter)
			}
		}
		return
	}()
	if ok {
		return lim
	}

	lim = c.limiter.In(context.Background(), addr.String())
	c.inLimits.Set(addr.String(), lim, 0)

	return lim
}

func (c *udpConn) getOutLimiter(addr net.Addr) limiter.ILimiter {
	if c.limiter == nil {
		return nil
	}

	lim, ok := func() (lim limiter.ILimiter, ok bool) {
		v, ok := c.outLimits.Get(addr.String())
		if ok {
			if v != nil {
				lim = v.(limiter.ILimiter)
			}
		}
		return
	}()
	if ok {
		return lim
	}

	lim = c.limiter.Out(context.Background(), addr.String())
	c.outLimits.Set(addr.String(), lim, 0)

	return lim
}

func (c *udpConn) RemoteAddr() net.Addr {
	if nc, ok := c.PacketConn.(xnet.RemoteAddr); ok {
		return nc.RemoteAddr()
	}
	return nil
}

func (c *udpConn) SetReadBuffer(n int) error {
	if nc, ok := c.PacketConn.(xnet.SetBuffer); ok {
		return nc.SetReadBuffer(n)
	}
	return errUnsupport
}

func (c *udpConn) SetWriteBuffer(n int) error {
	if nc, ok := c.PacketConn.(xnet.SetBuffer); ok {
		return nc.SetWriteBuffer(n)
	}
	return errUnsupport
}

func (c *udpConn) Read(b []byte) (n int, err error) {
	if nc, ok := c.PacketConn.(io.Reader); ok {
		n, err = nc.Read(b)
		return
	}
	err = errUnsupport
	return
}

func (c *udpConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	for {
		n, addr, err = c.PacketConn.ReadFrom(p)
		if err != nil {
			return
		}

		// discard when exceed the limit size.
		if limiter := c.getInLimiter(addr); limiter != nil &&
			limiter.Wait(context.Background(), n) < n {
			continue
		}
		return
	}
}

func (c *udpConn) ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error) {
	if nc, ok := c.PacketConn.(udp.ReadUDP); ok {
		for {
			n, addr, err = nc.ReadFromUDP(b)
			if err != nil {
				return
			}

			// discard when exceed the limit size.
			if limiter := c.getInLimiter(addr); limiter != nil &&
				limiter.Wait(context.Background(), n) < n {
				continue
			}
			return
		}
	}
	err = errUnsupport
	return
}

func (c *udpConn) ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *net.UDPAddr, err error) {
	if nc, ok := c.PacketConn.(udp.ReadUDP); ok {
		for {
			n, oobn, flags, addr, err = nc.ReadMsgUDP(b, oob)
			if err != nil {
				return
			}

			// discard when exceed the limit size.
			if limiter := c.getInLimiter(addr); limiter != nil &&
				limiter.Wait(context.Background(), n) < n {
				continue
			}
			return
		}
	}
	err = errUnsupport
	return
}

func (c *udpConn) Write(b []byte) (n int, err error) {
	if nc, ok := c.PacketConn.(io.Writer); ok {
		n, err = nc.Write(b)
		return
	}
	err = errUnsupport
	return
}

func (c *udpConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	// discard when exceed the limit size.
	if limiter := c.getOutLimiter(addr); limiter != nil &&
		limiter.Wait(context.Background(), len(p)) < len(p) {
		n = len(p)
		return
	}

	n, err = c.PacketConn.WriteTo(p, addr)
	return
}

func (c *udpConn) WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error) {
	// discard when exceed the limit size.
	if limiter := c.getOutLimiter(addr); limiter != nil &&
		limiter.Wait(context.Background(), len(b)) < len(b) {
		n = len(b)
		return
	}

	if nc, ok := c.PacketConn.(udp.WriteUDP); ok {
		n, err = nc.WriteToUDP(b, addr)
		return
	}
	err = errUnsupport
	return
}

func (c *udpConn) WriteMsgUDP(b, oob []byte, addr *net.UDPAddr) (n, oobn int, err error) {
	// discard when exceed the limit size.
	if limiter := c.getOutLimiter(addr); limiter != nil &&
		limiter.Wait(context.Background(), len(b)) < len(b) {
		n = len(b)
		return
	}

	if nc, ok := c.PacketConn.(udp.WriteUDP); ok {
		n, oobn, err = nc.WriteMsgUDP(b, oob, addr)
		return
	}
	err = errUnsupport
	return
}

func (c *udpConn) SyscallConn() (rc syscall.RawConn, err error) {
	if nc, ok := c.PacketConn.(xnet.SyscallConn); ok {
		return nc.SyscallConn()
	}
	err = errUnsupport
	return
}

func (c *udpConn) SetDSCP(n int) error {
	if nc, ok := c.PacketConn.(xnet.SetDSCP); ok {
		return nc.SetDSCP(n)
	}
	return nil
}

func (c *udpConn) Metadata() metadata.IMetaData {
	if md, ok := c.PacketConn.(metadata.IMetaDatable); ok {
		return md.Metadata()
	}
	return nil
}
