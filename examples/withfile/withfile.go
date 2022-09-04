package main

import (
	"fmt"
	"log"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/types"
)

var addr = "localhost:9090"

func IndexHandler(_ *types.Request) types.Response {
	return types.WithResponse.WithFile("index.html", func(err error) types.Response {
		return types.WithResponse.
			WithCode(status.InternalServerError).
			WithBody("Error: " + err.Error())
	})
}

func main() {
	r := inbuilt.NewRouter()
	r.Get("/", IndexHandler)

	app := indigo.NewApp(addr)
	fmt.Println("Listening on", addr)
	log.Fatal(app.Serve(r))
}
