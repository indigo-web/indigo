package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/types"
)

var addr = "localhost:9090"

func Index(_ *types.Request) types.Response {
	return types.WithResponse.WithFile("index.html", func(err error) types.Response {
		return types.WithResponse.
			WithCode(status.NotFound).
			WithBody(
				"index.html: not found; try running this example directly from examples/combined folder",
			)
	})
}

func IndexSay(request *types.Request) types.Response {
	if talking, found := request.Headers["talking"]; !found || talking[0].Value != "allowed" {
		return types.WithResponse.
			WithCode(status.UnavailableForLegalReasons)
	}

	body, err := request.Body()
	if err != nil {
		return types.WithResponse.WithError(err)
	}

	fmt.Println("Somebody said:", strconv.Quote(string(body)))

	return types.WithResponse
}

func World(_ *types.Request) types.Response {
	return types.WithResponse.WithBody(
		`<h1>Hello, world!</h1>`,
	)
}

func Easter(request *types.Request) types.Response {
	if _, found := request.Headers["easter"]; found {
		return types.WithResponse.
			WithCode(status.Teapot).
			WithHeader("Easter", "Egg").
			WithBody("Easter egg!")
	}

	return types.WithResponse.WithBody("Pretty ordinary page, isn't it?")
}

func main() {
	r := inbuilt.NewRouter()

	r.Get("/", Index)
	r.Post("/", IndexSay)

	hello := r.Group("/hello")
	hello.Get("/world", World)
	hello.Get("/easter", Easter)

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(r))
}
