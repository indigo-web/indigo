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

var addr = "localhost:9090"

func Index(request *http.Request) http.Response {
	return request.Respond.WithFile("index.html", func(err error) http.Response {
		return request.Respond.
			WithCode(status.NotFound).
			WithBody(
				"index.html: not found; try running this example directly from examples/combined folder",
			)
	})
}

func IndexSay(request *http.Request) http.Response {
	if talking := request.Headers.Value("talking"); talking != "allowed" {
		return request.Respond.WithCode(status.UnavailableForLegalReasons)
	}

	body, err := request.Body()
	if err != nil {
		return request.Respond.WithError(err)
	}

	fmt.Println("Somebody said:", strconv.Quote(string(body)))

	return request.Respond
}

func World(request *http.Request) http.Response {
	return request.Respond.WithBody(
		`<h1>Hello, world!</h1>`,
	)
}

func Easter(request *http.Request) http.Response {
	if easter := request.Headers.Value("easter"); len(easter) > 0 {
		return request.Respond.WithCode(status.Teapot).
			WithHeader("Easter", "Egg").
			WithBody("Easter egg!")
	}

	return request.Respond.WithBody("Pretty ordinary page, isn't it?")
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
