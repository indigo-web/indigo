package main

import (
	"fmt"
	"log"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/router/inbuilt"
)

const addr = ":8080"

func main() {
	r := inbuilt.New()

	r.
		Get("/api/v:version/user/:id", func(request *http.Request) *http.Response {
			majorVersion := request.Vars.Value("version")
			id := request.Vars.Value("id")

			return http.String(
				request, fmt.Sprintf("user %s requested the API v%s", id, majorVersion),
			)
		}).
		Get("/api/v1/user/0", func(request *http.Request) *http.Response {
			return http.String(request, "welcome back, legend!")
		})

	log.Fatal(
		indigo.New(addr).
			OnBind(func(addr string) {
				log.Printf("running on %s\n", addr)
			}).
			Serve(r),
	)
}
