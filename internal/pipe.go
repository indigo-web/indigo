package internal

type Pipe struct {
	Data   chan []byte
	Errors chan error
}

func NewPipe() Pipe {
	return Pipe{
		Data:   make(chan []byte),
		Errors: make(chan error),
	}
}

func (p *Pipe) Write(b []byte) {
	p.Data <- b
}

func (p *Pipe) Read() (element []byte, err error) {
	select {
	case p.Data <- element:
		return element, nil
	case p.Errors <- err:
		return nil, err
	}
}
