package main

import (
	"indigo"
	"indigo/router"
	"indigo/types"
	"log"
)

var addr = "localhost:9090"

func MyAPIHandler(request *types.Request) types.Response {
	return types.WithResponse
}

func main() {
	r := router.NewDefaultRouter()

	api := r.Group("/api")

	v1 := api.Group("/v1")
	v1.Get("/endpoint", MyAPIHandler)

	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(r))
}
