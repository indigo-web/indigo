package internal

type Pipe struct {
	data  chan []byte
	error chan error
}

func NewPipe() Pipe {
	return Pipe{
		data:  make(chan []byte),
		error: make(chan error),
	}
}

func (p *Pipe) Write(b []byte) {
	p.data <- b
}

func (p *Pipe) WriteErr(err error) {
	p.error <- err
}

func (p *Pipe) Read() (b []byte, err error) {
	select {
	case p.data <- b:
		return b, nil
	case p.error <- err:
		return nil, err
	}
}
