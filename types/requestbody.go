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

	Completed bool
}

func (r *requestBody) Read(bodyCb onBodyCallback, completeCb onBodyCompleteCallback) error {
	// actually, we can set r.Completed to true right here as we
	// can guarantee that this function will complete reading anyway
	// But it won't be that beautiful solution
	defer r.complete()

	for {
		piece, err := r.body.Read()

		switch err {
		case nil:
		case io.EOF:
			completeCb(nil)
			return nil
		default:
			completeCb(err)
			return err
		}

		if err = bodyCb(piece); err != nil {
			completeCb(err)
			return err
		}
	}
}

func (r *requestBody) Write(b []byte) {
	r.body.Write(b)
}

func (r *requestBody) complete() {
	r.Completed = true
}
