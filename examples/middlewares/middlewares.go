package main

import (
	"fmt"
	"github.com/fakefloordiv/indigo/http"
	"log"

	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/router/inbuilt"
)

var addr = "localhost:9090"

func HelloWorldMiddleware(next routertypes.HandlerFunc, request *http.Request) http.Response {
	fmt.Println("running middleware before handler")
	response := next(request)
	fmt.Println("running middleware after handler")

	return response
}

func SecondMiddleware(next routertypes.HandlerFunc, request *http.Request) http.Response {
	fmt.Println("running second middleware before first one")
	response := next(request)
	fmt.Println("running second middleware after first one")

	return response
}

func MyBeautifulHandler(request *http.Request) http.Response {
	fmt.Println("running handler")

	return http.RespondTo(request)
}

func main() {
	r := inbuilt.NewRouter()

	api := r.Group("/api")
	api.Use(HelloWorldMiddleware)

	v1 := api.Group("/v1")
	v1.Use(SecondMiddleware)
	v1.Get("/hello", MyBeautifulHandler)

	fmt.Println("listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(r))
}
