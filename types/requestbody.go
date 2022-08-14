package types

import (
	"indigo/internal"
)

type (
	onBodyCallback     func(b []byte) error
	onCompleteCallback func(err error)
)

// requestBody is a struct that handles request body. Separated from types.Request
// because contains a lot of internal logic and conventions that are exotic
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

func (r *requestBody) Read(onBody onBodyCallback, onComplete onCompleteCallback) (err error) {
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

func (r *requestBody) Reset() error {
	if r.read {
		return nil
	}

	r.read = false

	for {
		if <-r.body.Data == nil {
			return r.body.Err
		}
	}
}
