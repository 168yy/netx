package sshd

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/168yy/netx/core/dialer"
	md "github.com/168yy/netx/core/metadata"
	ssh_util "github.com/168yy/netx/x/internal/util/ssh"
	"golang.org/x/crypto/ssh"
)

type sshdDialer struct {
	sessions     map[string]*ssh_util.Session
	sessionMutex sync.Mutex
	md           metadata
	options      dialer.Options
}

func NewDialer(opts ...dialer.Option) dialer.IDialer {
	options := dialer.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	return &sshdDialer{
		sessions: make(map[string]*ssh_util.Session),
		options:  options,
	}
}

func (d *sshdDialer) Init(md md.IMetaData) (err error) {
	if err = d.parseMetadata(md); err != nil {
		return
	}

	return nil
}

// Multiplex implements dialer.IMultiplexer interface.
func (d *sshdDialer) Multiplex() bool {
	return true
}

func (d *sshdDialer) Dial(ctx context.Context, addr string, opts ...dialer.DialOption) (conn net.Conn, err error) {
	d.sessionMutex.Lock()
	defer d.sessionMutex.Unlock()

	session, ok := d.sessions[addr]
	if session != nil && session.IsClosed() {
		delete(d.sessions, addr) // session is dead
		ok = false
	}
	if !ok {
		var options dialer.DialOptions
		for _, opt := range opts {
			opt(&options)
		}

		conn, err = options.NetDialer.Dial(ctx, "tcp", addr)
		if err != nil {
			return
		}

		if d.md.handshakeTimeout > 0 {
			conn.SetDeadline(time.Now().Add(d.md.handshakeTimeout))
			defer conn.SetDeadline(time.Time{})
		}

		session, err = d.initSession(ctx, addr, conn)
		if err != nil {
			conn.Close()
			return nil, err
		}
		if d.md.keepalive {
			go session.Keepalive(d.md.keepaliveInterval, d.md.keepaliveTimeout, d.md.keepaliveRetries)
		}
		go session.Wait()
		go session.WaitClose()

		d.sessions[addr] = session
	}

	return ssh_util.NewClientConn(session), nil
}

func (d *sshdDialer) initSession(ctx context.Context, addr string, conn net.Conn) (*ssh_util.Session, error) {
	config := ssh.ClientConfig{
		Timeout:         d.md.handshakeTimeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if d.options.Auth != nil {
		config.User = d.options.Auth.Username()
		if password, _ := d.options.Auth.Password(); password != "" {
			config.Auth = []ssh.AuthMethod{
				ssh.Password(password),
			}
		}
	}
	if d.md.signer != nil {
		config.Auth = append(config.Auth, ssh.PublicKeys(d.md.signer))
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, &config)
	if err != nil {
		return nil, err
	}

	return ssh_util.NewSession(conn, ssh.NewClient(sshConn, chans, reqs), d.options.Logger), nil
}
