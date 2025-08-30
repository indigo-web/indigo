package transport

import (
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/status"
	"github.com/stretchr/testify/require"
)

type transportMock struct {
	stopped     *atomic.Bool
	closed      bool
	bound       bool
	once        bool
	loop        time.Duration
	returnError error
}

func newMock(loop time.Duration, returnError error, once bool) *transportMock {
	return &transportMock{
		stopped:     new(atomic.Bool),
		once:        once,
		loop:        loop,
		returnError: returnError,
	}
}

func (t *transportMock) Bind(string) error {
	t.bound = true
	return nil
}

func (t *transportMock) Listen(config.NET, func(conn net.Conn)) error {
	for !t.stopped.Load() && !t.once {
		time.Sleep(t.loop)
	}

	return t.returnError
}

func (t *transportMock) Stop() {
	t.stopped.Store(true)
}

func (t *transportMock) Close() {
	t.closed = true
}

func (t *transportMock) Wait() {
	for !t.stopped.Load() {
		time.Sleep(1 * time.Millisecond)
	}
}

func runParallel(fn func() error) chan error {
	c := make(chan error)

	go func() {
		c <- fn()
	}()

	return c
}

func runAtMost(sup *Supervisor, timeout time.Duration) error {
	select {
	case err := <-runParallel(func() error {
		return sup.Run(config.Default().NET)
	}):
		return err
	case <-time.After(timeout):
		return fmt.Errorf("supervisor timeouted")
	}
}

func TestSupervisor(t *testing.T) {
	newSupervisor := func(ts ...*transportMock) (*Supervisor, error) {
		sup := NewSupervisor()
		for _, transport := range ts {
			if err := sup.Add("", transport, nil); err != nil {
				return nil, err
			}
		}

		return &sup, nil
	}

	t.Run("die without error", func(t *testing.T) {
		sup, err := newSupervisor(
			newMock(100*time.Millisecond, nil, false),
			newMock(200*time.Millisecond, nil, true),
		)
		require.NoError(t, err)
		require.NoError(t, runAtMost(sup, 300*time.Millisecond))
	})

	t.Run("die with error", func(t *testing.T) {
		sup, err := newSupervisor(
			newMock(100*time.Millisecond, nil, false),
			newMock(200*time.Millisecond, status.ErrCloseConnection, true),
		)
		require.NoError(t, err)
		require.EqualError(t, runAtMost(sup, 300*time.Millisecond), status.ErrCloseConnection.Error())
	})

	t.Run("stop", func(t *testing.T) {
		sup, err := newSupervisor(
			newMock(100*time.Millisecond, nil, false),
			newMock(200*time.Millisecond, nil, false),
		)
		require.NoError(t, err)
		c := runParallel(func() error {
			return sup.Run(config.Default().NET)
		})
		time.Sleep(200 * time.Millisecond)
		c2 := runParallel(func() error {
			sup.Stop()
			return nil
		})

		select {
		case err = <-c2:
			require.NoError(t, err)
		case <-time.After(300 * time.Millisecond):
			require.Fail(t, "supervisor did not stop on time")
		}

		select {
		case err = <-c:
			require.NoError(t, err)
		case <-time.After(50 * time.Millisecond):
			require.Fail(t, "supervisor did not stop running on time")
		}
	})
}
