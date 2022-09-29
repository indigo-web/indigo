package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/types"
)

var addr = "localhost:9090"

func Index(_ context.Context, _ *types.Request) types.Response {
	return types.WithFile("index.html", func(err error) types.Response {
		return types.WithResponse.
			WithCode(status.NotFound).
			WithBody(
				"index.html: not found; try running this example directly from examples/combined folder",
			)
	})
}

func IndexSay(_ context.Context, request *types.Request) types.Response {
	if talking, found := request.Headers["talking"]; !found || talking[0].Value != "allowed" {
		return types.WithCode(status.UnavailableForLegalReasons)
	}

	body, err := request.Body()
	if err != nil {
		return types.WithError(err)
	}

	fmt.Println("Somebody said:", strconv.Quote(string(body)))

	return types.OK()
}

func World(_ context.Context, _ *types.Request) types.Response {
	return types.WithBody(
		`<h1>Hello, world!</h1>`,
	)
}

func Easter(_ context.Context, request *types.Request) types.Response {
	if _, found := request.Headers["easter"]; found {
		return types.
			WithCode(status.Teapot).
			WithHeader("Easter", "Egg").
			WithBody("Easter egg!")
	}

	return types.WithBody("Pretty ordinary page, isn't it?")
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
