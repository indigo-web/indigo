package middleware

import (
	"github.com/indigo-web/indigo/v2/http"
	"github.com/indigo-web/indigo/v2/http/status"
	"github.com/indigo-web/indigo/v2/router/inbuilt/types"
)

// Recover is a basic middleware that catches any panics, and returns 500 Internal Server Error
// instead. Response headers and body are being discarded in consistency purposes, avoiding
// half-cooked response being sent
func Recover(next types.HandlerFunc, req *http.Request) (resp http.Response) {
	defer func() {
		if r := recover(); r != nil {
			resp = http.RespondTo(req).
				DiscardHeaders().
				WithBody("").
				WithError(status.ErrInternalServerError)
		}
	}()

	resp = next(req)
	return
}
