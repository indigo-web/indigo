package main

import (
	"fmt"
	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/codec"
	"github.com/indigo-web/indigo/router/inbuilt"
	"log"
	"strconv"
)

func Shout(r *http.Request) *http.Response {
	body, err := r.Body.String()
	if err != nil {
		return http.Error(r, err)
	}

	fmt.Println("a quiet whisper comes from somewhere:", strconv.Quote(body))
	return http.Respond(r)
}

func main() {
	app := indigo.New(":8080").
		Codec(codec.NewGZIP()).
		OnBind(func(addr string) {
			fmt.Println("Listening on", addr)
		})

	r := inbuilt.New().
		Get("/", inbuilt.File("./examples/compression/index.html")).
		Post("/submit", Shout)

	if err := app.Serve(r); err != nil {
		log.Fatal(err)
	}
}
