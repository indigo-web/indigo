package main

import (
	"fmt"
	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt"
	"log"
	"strconv"
	"strings"
)

func Shout(r *http.Request) *http.Response {
	body, err := r.Body.String()
	if err != nil {
		return http.Error(r, err)
	}

	fmt.Println(strings.ToUpper(strconv.Quote(body)))
	return http.Respond(r)
}

func main() {
	app := indigo.New(":8080")

	r := inbuilt.New().
		Get("/", inbuilt.File("index.html")).
		Post("/submit", Shout)

	if err := app.Serve(r); err != nil {
		log.Fatal(err)
	}
}
