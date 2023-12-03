package main

import (
	"log"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/router/inbuilt"
)

const addr = ":8080"

func HelloWorldMiddleware(next inbuilt.Handler, request *http.Request) *http.Response {
	log.Println("running middleware before handler")
	response := next(request)
	log.Println("running middleware after handler")

	return response
}

func SecondMiddleware(next inbuilt.Handler, request *http.Request) *http.Response {
	log.Println("running second middleware before first one")
	response := next(request)
	log.Println("running second middleware after first one")

	return response
}

func MyBeautifulHandler(request *http.Request) *http.Response {
	log.Println("running handler")

	return request.Respond()
}

func main() {
	r := inbuilt.New()

	api := r.Group("/api")
	api.Use(HelloWorldMiddleware)

	v1 := api.Group("/v1")
	v1.Use(SecondMiddleware)
	v1.Get("/hello", MyBeautifulHandler)

	app := indigo.NewApp(addr)
	log.Println("listening on", addr)
	log.Fatal(app.Serve(r))
}
