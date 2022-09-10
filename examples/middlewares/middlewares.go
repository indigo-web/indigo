package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/types"
)

var addr = "localhost:9090"

func HelloWorldMiddleware(ctx context.Context, next inbuilt.HandlerFunc, request *types.Request) types.Response {
	fmt.Println("running middleware before handler")
	response := next(ctx, request)
	fmt.Println("running middleware after handler")

	return response
}

func SecondMiddleware(ctx context.Context, next inbuilt.HandlerFunc, request *types.Request) types.Response {
	fmt.Println("running second middleware before first one")
	response := next(ctx, request)
	fmt.Println("running second middleware after first one")

	return response
}

func MyBeautifulHandler(_ context.Context, _ *types.Request) types.Response {
	fmt.Println("running handler")

	return types.WithResponse
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
