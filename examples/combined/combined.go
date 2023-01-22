package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
)

var (
	addr  = "localhost:9090"
	index = "index.html"
)

func Index(request *http.Request) http.Response {
	return http.RespondTo(request).WithFile(index, func(err error) http.Response {
		return http.RespondTo(request).
			WithCode(status.NotFound).
			WithBody(
				index + ": not found; try running this example directly from examples/combined folder",
			)
	})
}

func IndexSay(request *http.Request) http.Response {
	if talking := request.Headers.Value("talking"); talking != "allowed" {
		return http.RespondTo(request).WithCode(status.UnavailableForLegalReasons)
	}

	body, err := request.Body()
	if err != nil {
		return http.RespondTo(request).WithError(err)
	}

	fmt.Println("Somebody said:", strconv.Quote(string(body)))

	return http.RespondTo(request)
}

func World(request *http.Request) http.Response {
	return http.RespondTo(request).WithBody(
		`<h1>Hello, world!</h1>`,
	)
}

func Easter(request *http.Request) http.Response {
	if request.Headers.Has("easter") {
		return http.RespondTo(request).
			WithCode(status.Teapot).
			WithHeader("Easter", "Egg").
			WithBody("You have discovered an easter egg! Congratulations!")
	}

	return http.RespondTo(request).WithBody("Pretty ordinary page, isn't it?")
}

func main() {
	r := inbuilt.NewRouter()

	root := r.Resource("/")
	root.Get(Index)
	root.Post(IndexSay)

	hello := r.Group("/hello")
	hello.Get("/world", World)
	hello.Get("/easter", Easter)

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(r))
}
