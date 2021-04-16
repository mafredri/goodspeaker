package net

import (
	"context"
	"net"
	"time"
)

type Dialer struct {
	Timeout time.Duration // TODO(mafredri): Add support for timeout.
}

func (d *Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if network != "tcp" {
		panic("unsupported network type")
	}

	s := newSocket(addr)

	// Establish connection.
	err := s.connect(addr)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Dial creates a new net.Socket and connects to the remote.
func Dial(network, addr string) (net.Conn, error) {
	d := Dialer{}
	return d.DialContext(context.Background(), network, addr)
}
