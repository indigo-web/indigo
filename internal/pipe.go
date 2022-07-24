package internal

import (
	"sync/atomic"
)

/*
Pipe is a simple implementation of io.Pipe, except re-usability. This means
that after error is written pipe is still usable as it was before. This was
made in optimization purposes only

Also pipe provides Readable() method that tells whether we have anything to
read. This implemented by atomic int32 counter of how much is written and
unread. As type is int32 (not *int32), copying pipe corrupts it. So be care,
I lost like half and hour debugging the problem was caused because of that
*/
type Pipe struct {
	data  chan []byte
	error chan error

	entries int32
}

func NewChanSizedPipe(dataChanSize, errChanSize int) *Pipe {
	return &Pipe{
		data:  make(chan []byte, dataChanSize),
		error: make(chan error, errChanSize),
	}
}

func NewPipe() *Pipe {
	return NewChanSizedPipe(0, 0)
}

func (p *Pipe) Write(b []byte) {
	atomic.AddInt32(&p.entries, 1)
	p.data <- b
}

func (p *Pipe) WriteErr(err error) {
	atomic.AddInt32(&p.entries, 1)
	p.error <- err
}

func (p *Pipe) Read() (b []byte, err error) {
	atomic.AddInt32(&p.entries, -1)
	select {
	case b = <-p.data:
		return b, nil
	case err = <-p.error:
		return nil, err
	}
}

func (p Pipe) Readable() bool {
	return p.entries > 0
}
