package types

import (
	"errors"
	"github.com/fakefloordiv/indigo/internal"
)

type (
	onBodyCallback     func(b []byte) error
	onCompleteCallback func(err error)
)

var ErrRead = errors.New("body has been already read")

// requestBody is a struct that handles request body. Separated from types.Request
// because contains a lot of internal logic and conventions
// TODO: add some struct that implements io.Reader interface. I propose implement it
//       as a struct equal to this one, but sends signal about completion on a next
//       Read() call (such a structure looks pretty unstable as for me, but no choice)
type requestBody struct {
	body *internal.BodyGateway
	read bool
}

func newRequestBody() (requestBody, *internal.BodyGateway) {
	gateway := internal.NewBodyGateway()

	return requestBody{
		body: gateway,
	}, gateway
}

// Read reads a body until nil is met. After it is, onComplete is called with error
// presented in r.body.Err attribute. If onBody returned error, it will be also passed
// into the onComplete callback and returned
// Due to conventions, we have to receive a piece of body, process it (and copy, otherwise
// it will be rewritten as passed slice is a part of buffer for reading from socket),
// and send nil back (and put an error into the r.body.Err attribute if we have; this will
// close the connection without calling router.OnError callback). If we received nil from
// body channel, this means that body is over - in this case, onComplete with r.body.Err
// is called, and r.body.Err is returned. Also, when we meet end of body, we are not supposed
// to notify server back when we processed it because it has no sense
func (r *requestBody) Read(onBody onBodyCallback, onComplete onCompleteCallback) (err error) {
	if r.read {
		return ErrRead
	}

	r.read = true

	for {
		data := <-r.body.Data
		if data == nil {
			onComplete(r.body.Err)

			return r.body.Err
		}

		if err = onBody(data); err != nil {
			r.body.Err = err
			r.body.Data <- nil

			return err
		}

		r.body.Data <- nil
	}
}

// Reset resets body by reading it into nowhere if it was not read by user
// If it is already read, doing nothing and returning nil-error
func (r *requestBody) Reset() error {
	if r.read {
		r.read = false

		return nil
	}

	for {
		if <-r.body.Data == nil {
			return r.body.Err
		}
	}
}
