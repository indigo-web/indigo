package main

import (
	"fmt"
	"log"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/router/inbuilt"
)

var addr = "localhost:9090"

func MyDynamicHandler(request *http.Request) *http.Response {
	worldName := request.Params["world-name"]

	return request.Respond().WithBody("your world-name is " + worldName)
}

func main() {
	r := inbuilt.New()

	r.Get("/hello/{world-name}", MyDynamicHandler)

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(r))
}
