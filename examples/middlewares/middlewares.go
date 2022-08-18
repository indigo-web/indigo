package main

import (
	"fmt"
	"indigo"
	"indigo/router"
	"indigo/types"
	"log"
)

var addr = "localhost:9090"

func HelloWorldMiddleware(next router.HandlerFunc, request *types.Request) types.Response {
	fmt.Println("running middleware before handler")
	response := next(request)
	fmt.Println("running middleware after handler")

	return response
}

func SecondMiddleware(next router.HandlerFunc, request *types.Request) types.Response {
	fmt.Println("running second middleware before first one")
	response := next(request)
	fmt.Println("running second middleware after first one")

	return response
}

func MyBeautifulHandler(_ *types.Request) types.Response {
	fmt.Println("running handler")

	return types.WithResponse
}

func main() {
	r := router.NewDefaultRouter()

	api := r.Group("/api")
	api.Use(HelloWorldMiddleware)

	v1 := api.Group("/v1")
	v1.Use(SecondMiddleware)
	v1.Get("/hello", MyBeautifulHandler)

	fmt.Println("listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(r))
}
