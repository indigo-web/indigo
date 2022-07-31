package types

import (
	"indigo/internal"
	"io"
)

type (
	onBodyCallback         func(b []byte) error
	onBodyCompleteCallback func(err error)
)

type requestBody struct {
	body internal.Pipe
	read bool
}

func (r *requestBody) Read(bodyCb onBodyCallback, completeCb onBodyCompleteCallback) error {
	defer func() {
		r.read = true
	}()

	for {
		piece, err := r.body.Read()

		switch err {
		case nil:
			if err = bodyCb(piece); err != nil {
				completeCb(err)
				return err
			}
		case io.EOF:
			completeCb(nil)
			return nil
		default:
			completeCb(err)
			return err
		}
	}
}

func (r *requestBody) Write(b []byte) {
	r.body.Write(b)
}

func (r *requestBody) Reset() {
	r.eraseBody()
	r.read = false
}

func (r *requestBody) eraseBody() {
	for !r.read {
		if _, err := r.body.Read(); err != nil {
			break
		}
	}

	r.read = true
}
