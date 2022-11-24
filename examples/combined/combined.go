package main

import (
	"fmt"
	"github.com/fakefloordiv/indigo/http"
	"log"
	"strconv"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/router/inbuilt"
)

var (
	addr  = "localhost:9090"
	index = "index.html"
)

func Index(request *http.Request) http.Response {
	return http.Respond(request).WithFile(index, func(err error) http.Response {
		return http.Respond(request).
			WithCode(status.NotFound).
			WithBody(
				index + ": not found; try running this example directly from examples/combined folder",
			)
	})
}

func IndexSay(request *http.Request) http.Response {
	if talking := request.Headers.Value("talking"); talking != "allowed" {
		return http.Respond(request).WithCode(status.UnavailableForLegalReasons)
	}

	body, err := request.Body()
	if err != nil {
		return http.Respond(request).WithError(err)
	}

	fmt.Println("Somebody said:", strconv.Quote(string(body)))

	return http.Respond(request)
}

func World(request *http.Request) http.Response {
	return http.Respond(request).WithBody(
		`<h1>Hello, world!</h1>`,
	)
}

func Easter(request *http.Request) http.Response {
	if easter := request.Headers.Value("easter"); len(easter) > 0 {
		return http.Respond(request).
			WithCode(status.Teapot).
			WithHeader("Easter", "Egg").
			WithBody("You have discovered an easter egg! Congratulations!")
	}

	return http.Respond(request).WithBody("Pretty ordinary page, isn't it?")
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
