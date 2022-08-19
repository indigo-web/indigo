package internal

/*
BodyGateway is an embedded solution instead of pipe

The idea is:
- Core pushes some data into the Data channel
- Userspace receives and checks whether Err is not nil
- After userspace completes processing collected data, it writes
  nil to channel (and sets error to something non-nil)
- Core waits for that nil, and when got, it checks whether Err is
  nil
*/
type BodyGateway struct {
	Data chan []byte
	Err  error
}

func NewBodyGateway() *BodyGateway {
	return &BodyGateway{
		Data: make(chan []byte),
	}
}

// WriteErr is simply a sugar for setting an error and sending nil to the channel
func (b *BodyGateway) WriteErr(err error) {
	b.Err = err
	b.Data <- nil
}
