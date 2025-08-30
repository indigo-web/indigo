package transport

import (
	"net"
	"sync/atomic"

	"github.com/indigo-web/indigo/config"
)

type Transport interface {
	Bind(addr string) error
	Listen(cfg config.NET, cb func(conn net.Conn)) error
	Stop()
	Close()
	Wait()
}

type Supervisor struct {
	stopped *atomic.Bool
	ts      []boundTransport
	stopch  chan struct{}
}

func NewSupervisor() Supervisor {
	return Supervisor{
		stopped: new(atomic.Bool),
		stopch:  make(chan struct{}),
	}
}

func (s *Supervisor) Add(addr string, transport Transport, cb func(net.Conn)) error {
	err := transport.Bind(addr)
	if err != nil {
		s.close()
		return err
	}

	s.ts = append(s.ts, boundTransport{
		cb: cb,
		t:  transport,
	})

	return nil
}

func (s *Supervisor) Run(cfg config.NET) error {
	if len(s.ts) == 0 {
		return nil
	}

	errch := make(chan error)

	for _, t := range s.ts {
		go func(t boundTransport, ch chan<- error) {
			errch <- t.t.Listen(cfg, t.cb)
		}(t, errch)
	}

	select {
	case err := <-errch:
		s.stop()
		drain(errch, len(s.ts)-1)

		return err
	case <-s.stopch:
		s.stop()
		drain(errch, len(s.ts))
		s.stopch <- struct{}{}

		return nil
	}
}

func (s *Supervisor) Stop() {
	if !s.stopped.Load() {
		s.stopch <- struct{}{}
		<-s.stopch
	}
}

func (s *Supervisor) stop() {
	if s.stopped.Load() {
		return
	}

	s.stopped.Store(true)

	for _, t := range s.ts {
		t.t.Stop()
	}

	for _, t := range s.ts {
		t.t.Wait()
		t.t.Close()
	}
}

func (s *Supervisor) close() {
	for _, t := range s.ts {
		t.t.Close()
	}
}

type boundTransport struct {
	cb func(conn net.Conn)
	t  Transport
}

func drain(ch <-chan error, n int) {
	for range n {
		<-ch
	}
}
