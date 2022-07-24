package internal

import "sync/atomic"

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
	defer atomic.AddInt32(&p.entries, -1)

	select {
	case b = <-p.data:
		return b, nil
	case err = <-p.error:
		return nil, err
	}
}

func (p *Pipe) Readable() bool {
	return p.entries > 0
}
