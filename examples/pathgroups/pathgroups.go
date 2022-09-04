package main

import (
	"fmt"
	"log"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/types"
)

var addr = "localhost:9090"

func MyAPIHandler(_ *types.Request) types.Response {
	return types.WithResponse
}

func main() {
	r := inbuilt.NewRouter()

	api := r.Group("/api")

	v1 := api.Group("/v1")
	v1.Get("/endpoint", MyAPIHandler)

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(r))
}
