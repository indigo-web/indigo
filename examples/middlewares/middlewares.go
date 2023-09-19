package main

import (
	"fmt"
	"log"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo/router/inbuilt/types"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/router/inbuilt"
)

var addr = "localhost:8080"

func HelloWorldMiddleware(next types.Handler, request *http.Request) *http.Response {
	fmt.Println("running middleware before handler")
	response := next(request)
	fmt.Println("running middleware after handler")

	return response
}

func SecondMiddleware(next types.Handler, request *http.Request) *http.Response {
	fmt.Println("running second middleware before first one")
	response := next(request)
	fmt.Println("running second middleware after first one")

	return response
}

func MyBeautifulHandler(request *http.Request) *http.Response {
	fmt.Println("running handler")

	return request.Respond()
}

func main() {
	r := inbuilt.New()

	api := r.Group("/api")
	api.Use(HelloWorldMiddleware)

	v1 := api.Group("/v1")
	v1.Use(SecondMiddleware)
	v1.Get("/hello", MyBeautifulHandler)

	fmt.Println("listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(r))
}
