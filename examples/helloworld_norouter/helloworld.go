package main

import (
	"log"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/simple"
)

const addr = ":8080"

func MyHandler(request *http.Request) *http.Response {
	return request.Respond().
		Code(status.OK).
		Header("Hello", "world").
		String("<h1>How are you doing?</h1>")
}

func main() {
	r := simple.New(MyHandler, func(request *http.Request) *http.Response {
		return http.Error(request, request.Env.Error)
	})
	app := indigo.New(addr).
		OnListenerStart(func(addr string) {
			log.Printf("running on %s\n", addr)
		})

	log.Fatal(app.Serve(r))
}
