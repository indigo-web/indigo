package main

import (
	"fmt"
	"github.com/indigo-web/indigo/http/decoder"
	"github.com/indigo-web/indigo/router/inbuilt/middleware"
	"github.com/indigo-web/indigo/settings"
	"log"
	"strconv"
	"time"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
)

var (
	addr  = "localhost:8080"
	index = "index.html"
)

func Index(request *http.Request) *http.Response {
	resp, err := request.Respond().WithFile(index)
	if err != nil {
		return request.Respond().
			WithCode(status.NotFound).
			WithBody(
				index + ": not found; try running this example directly from examples/combined folder",
			)
	}

	return resp
}

func IndexSay(request *http.Request) *http.Response {
	if talking := request.Headers.Value("talking"); talking != "allowed" {
		return request.Respond().WithCode(status.UnavailableForLegalReasons)
	}

	body, err := request.Body().Full()
	if err != nil {
		return request.Respond().WithError(err)
	}

	fmt.Println("Somebody said:", strconv.Quote(string(body)))

	return request.Respond()
}

func World(request *http.Request) *http.Response {
	return request.Respond().WithBody(
		`<h1>Hello, world!</h1>`,
	)
}

func Easter(request *http.Request) *http.Response {
	if request.Headers.Has("easter") {
		return request.Respond().
			WithCode(status.Teapot).
			WithHeader("Easter", "Egg").
			WithBody("You have discovered an easter egg! Congratulations!")
	}

	return request.Respond().WithBody("Pretty ordinary page, isn't it?")
}

func Stressful(request *http.Request) *http.Response {
	resp := request.Respond().
		WithHeader("Should", "never be seen").
		WithBody("Hello, world!")

	panic("TOO MUCH STRESS")

	return resp
}

func main() {
	r := inbuilt.New().
		Get("/stress", Stressful, middleware.Recover)

	r.Resource("/").
		Get(Index).
		Post(IndexSay)

	r.Group("/hello").
		Get("/world", World).
		Get("/easter", Easter)

	s := settings.Default()
	s.TCP.ReadTimeout = time.Hour

	app := indigo.NewApp(addr)
	app.AddContentDecoder("gzip", decoder.NewGZIPDecoder)
	fmt.Println("Listening on", addr)

	if err := app.Serve(r, s); err != nil {
		log.Fatal(err)
	}
}
