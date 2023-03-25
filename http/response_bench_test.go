package http

import (
	"errors"
	"github.com/indigo-web/indigo/http/status"
	"testing"
)

func BenchmarkResponse_WithError(b *testing.B) {
	resp := NewResponse()
	knownErr := status.ErrBadRequest
	unknownErr := errors.New("some crap happened, unable to recover")

	b.Run("KnownError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp.WithError(knownErr)
		}
	})

	b.Run("UnknownError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp.WithError(unknownErr)
		}
	})
}
