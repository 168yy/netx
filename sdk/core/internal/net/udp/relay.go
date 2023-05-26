package udp

import (
	"context"
	"net"

	"github.com/jxo-me/netx/sdk/core/bypass"
	"github.com/jxo-me/netx/sdk/core/common/bufpool"
	"github.com/jxo-me/netx/sdk/core/logger"
)

type Relay struct {
	pc1 net.PacketConn
	pc2 net.PacketConn

	bypass     bypass.IBypass
	bufferSize int
	logger     logger.ILogger
}

func NewRelay(pc1, pc2 net.PacketConn) *Relay {
	return &Relay{
		pc1: pc1,
		pc2: pc2,
	}
}

func (r *Relay) WithBypass(bp bypass.IBypass) *Relay {
	r.bypass = bp
	return r
}

func (r *Relay) WithLogger(logger logger.ILogger) *Relay {
	r.logger = logger
	return r
}

func (r *Relay) SetBufferSize(n int) {
	r.bufferSize = n
}

func (r *Relay) Run() (err error) {
	bufSize := r.bufferSize
	if bufSize <= 0 {
		bufSize = 4096
	}

	errc := make(chan error, 2)

	go func() {
		for {
			err := func() error {
				b := bufpool.Get(bufSize)
				defer bufpool.Put(b)

				n, raddr, err := r.pc1.ReadFrom(*b)
				if err != nil {
					return err
				}

				if r.bypass != nil && r.bypass.Contains(context.Background(), raddr.String()) {
					if r.logger != nil {
						r.logger.Warn("bypass: ", raddr)
					}
					return nil
				}

				if _, err := r.pc2.WriteTo((*b)[:n], raddr); err != nil {
					return err
				}

				if r.logger != nil {
					r.logger.Tracef("%s >>> %s data: %d",
						r.pc2.LocalAddr(), raddr, n)

				}

				return nil
			}()

			if err != nil {
				errc <- err
				return
			}
		}
	}()

	go func() {
		for {
			err := func() error {
				b := bufpool.Get(bufSize)
				defer bufpool.Put(b)

				n, raddr, err := r.pc2.ReadFrom(*b)
				if err != nil {
					return err
				}

				if r.bypass != nil && r.bypass.Contains(context.Background(), raddr.String()) {
					if r.logger != nil {
						r.logger.Warn("bypass: ", raddr)
					}
					return nil
				}

				if _, err := r.pc1.WriteTo((*b)[:n], raddr); err != nil {
					return err
				}

				if r.logger != nil {
					r.logger.Tracef("%s <<< %s data: %d",
						r.pc2.LocalAddr(), raddr, n)

				}

				return nil
			}()

			if err != nil {
				errc <- err
				return
			}
		}
	}()

	return <-errc
}
