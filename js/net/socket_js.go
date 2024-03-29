package net

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"syscall/js"
	"time"
)

var uint8Array = js.Global().Get("Uint8Array")

// socket wraps Socket into a net.Conn.
//
// https://nodejs.org/api/net.html#net_class_net_socket
type socket struct {
	s    js.Value
	fns  []js.Func
	rbuf *messageBuffer
	r    io.Reader
	done chan struct{} // Signals close and protects error.
	err  error
}

// newSocket initializes a new node net.Socket() with event handlers for
// transporting messages between Go and JS.
func newSocket(addr string) *socket {
	require := js.Global().Get("require")
	net := require.Invoke("net")

	s := &socket{
		s:    net.Get("Socket").New(),
		rbuf: newMessageBuffer(),
		done: make(chan struct{}),
	}

	s.on("data", s.recv)
	s.on("timeout", s.timeout)

	// As we don't support half-open sockets, so no need to handle "end".
	s.on("error", s.setCloseError)
	s.on("close", s.close)

	return s
}

// recv socket messages and store them in the buffer.
func (s *socket) recv(this js.Value, args []js.Value) interface{} {
	v := args[0]
	b := make([]byte, v.Length())
	js.CopyBytesToGo(b, v)

	select {
	case <-s.done:
		return nil
	default:
	}

	s.rbuf.store(b)
	return nil
}

func (s *socket) timeout(js.Value, []js.Value) interface{} {
	go func() {
		log.Println("js/net: Socket timeout")
		s.Close()
	}()
	return nil
}

// setCloseError sets the error that closed the connection.
func (s *socket) setCloseError(this js.Value, args []js.Value) interface{} {
	m := args[0].Get("message").String()
	s.err = errors.New(strings.ToLower(m))
	return nil
}

// close triggers connection shutdown.
func (s *socket) close(this js.Value, args []js.Value) interface{} {
	close(s.done)
	go s.Close()
	return nil
}

// on registers an event handler that is cleaned up on close.
func (s *socket) on(event string, fn func(this js.Value, args []js.Value) interface{}) interface{} {
	jsFn := js.FuncOf(fn)
	s.fns = append(s.fns, jsFn)
	return s.s.Call("on", event, jsFn)
}

func (s *socket) connect(addr string) error {
	addrParts := strings.Split(addr, ":")
	host, port := addrParts[0], addrParts[1]

	var cb js.Func
	ready := make(chan struct{})
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			log.Printf("js/net: Connected to %s!", addr)
		}()
		close(ready)
		cb.Release() // Release the function.
		return nil
	})
	s.s.Call("connect", port, host, cb)

	select {
	case <-ready:
	case <-s.done:
		return s.err
	}

	return nil
}

// Read implements io.Reader.
func (s *socket) Read(b []byte) (n int, err error) {
	if s.r == nil {
		select {
		case m := <-s.rbuf.get():
			s.rbuf.load()
			s.r = bytes.NewReader(m)
		case <-s.done:
			return 0, s.err
		}
	}

	n, err = s.r.Read(b)
	if err == io.EOF {
		s.r = nil
		err = nil

		if n == 0 {
			return s.Read(b)
		}
	}

	return n, err
}

// Write implements io.Writer.
func (s *socket) Write(p []byte) (n int, err error) {
	buf := uint8Array.New(len(p))
	js.CopyBytesToJS(buf, p)
	ret := s.s.Call("write", buf)

	var ok bool
	switch ret.Type() {
	case js.TypeBoolean:
		ok = ret.Bool()
	default:
		<-s.done
		return 0, s.err
	}

	if !ok {
		return 0, errors.New("socket: write failed")
	}
	return len(p), nil
}

func (s *socket) Close() error {
	s.s.Call("end")
	<-s.done
	s.s.Call("unref")
	for _, fn := range s.fns {
		fn.Release()
	}
	s.fns = nil
	s.rbuf.clear()
	return s.err
}

type tcpAddr string

func (a tcpAddr) Network() string { return "tcp" }
func (a tcpAddr) String() string  { return string(a) }

func (s *socket) LocalAddr() net.Addr {
	ip := s.s.Call("localAddress").String()
	port := s.s.Call("localPort").Int()
	return tcpAddr(fmt.Sprintf("%s:%d", ip, port))
}

func (s *socket) RemoteAddr() net.Addr {
	ip := s.s.Call("remoteAddress").String()
	port := s.s.Call("remotePort").Int()
	return tcpAddr(fmt.Sprintf("%s:%d", ip, port))
}

func (s *socket) SetDeadline(t time.Time) error {
	d := 0
	if !t.IsZero() {
		d = int(t.Sub(time.Now()).Milliseconds())
	}
	s.s.Call("setTimeout", d)
	return nil
}

func (s *socket) SetReadDeadline(t time.Time) error {
	return s.SetDeadline(t)
}

func (s *socket) SetWriteDeadline(t time.Time) error {
	return s.SetDeadline(t)
}
