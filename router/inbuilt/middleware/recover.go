package middleware

import (
	"log"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
)

// Recover is a basic middleware that catches any panics, logs it and returns 500 Internal Server Error
func Recover(next inbuilt.Handler, req *http.Request) (resp *http.Response) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %v\n", r)
			resp = http.Error(req, status.ErrInternalServerError)
		}
	}()

	resp = next(req)
	return
}
