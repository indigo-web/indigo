package main

import (
	"log"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/router/inbuilt"
)

const addr = ":8080"

func MyDynamicHandler(request *http.Request) *http.Response {
	worldName := request.Params.Value("world-name")

	return request.Respond().String("your world-name is " + worldName)
}

func main() {
	r := inbuilt.New()

	r.Get("/hello/{world-name}", MyDynamicHandler)

	app := indigo.New(addr).
		OnBind(func(addr string) {
			log.Printf("running on %s\n", addr)
		})

	log.Fatal(app.Serve(r))
}
