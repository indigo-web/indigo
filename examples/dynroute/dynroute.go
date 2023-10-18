package main

import (
	"log"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/router/inbuilt"
)

const (
	host = "0.0.0.0"
	port = 8080
)

func MyDynamicHandler(request *http.Request) *http.Response {
	worldName := request.Params["world-name"]

	return request.Respond().WithBody("your world-name is " + worldName)
}

func main() {
	r := inbuilt.New()

	r.Get("/hello/{world-name}", MyDynamicHandler)

	app := indigo.NewApp(host, port)
	log.Println("Listening on", host, port)
	log.Fatal(app.Serve(r))
}
