package transport

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/internal/timer"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type listener interface {
	net.Listener
	SetDeadline(t time.Time) error
}

type TCP struct {
	l    listener
	wg   *sync.WaitGroup
	stop *atomic.Bool
}

func NewTCP() *TCP {
	tcp := newTCP(nil)
	return &tcp
}

func newTCP(l listener) TCP {
	return TCP{
		l:    l,
		wg:   new(sync.WaitGroup),
		stop: new(atomic.Bool),
	}
}

func bindTCP(addr string) (*net.TCPListener, error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	return net.ListenTCP("tcp", tcpaddr)
}

func (t *TCP) Bind(addr string) (err error) {
	t.l, err = bindTCP(addr)
	return err
}

func (t *TCP) Listen(cfg config.TCP, cb func(conn net.Conn)) error {
	for !t.stop.Load() {
		err := t.l.SetDeadline(timer.Now().Add(cfg.AcceptLoopInterruptPeriod))
		if err != nil {
			return err
		}

		conn, err := t.l.Accept()
		if err != nil {
			if err.(*net.OpError).Err.Error() == os.ErrDeadlineExceeded.Error() {
				continue
			}

			return err
		}

		go func(conn net.Conn) {
			t.wg.Add(1)
			cb(conn)
			_ = conn.Close()
			t.wg.Done()
		}(conn)
	}

	return nil
}

func (t *TCP) Stop() {
	t.stop.Store(true)
}

func (t *TCP) Close() {
	_ = t.l.Close()
}

func (t *TCP) Wait() {
	t.wg.Wait()
}
