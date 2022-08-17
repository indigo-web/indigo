package main

import (
	"fmt"
	"indigo"
	"indigo/router"
	"indigo/types"
	"log"
)

var addr = "localhost:9090"

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
