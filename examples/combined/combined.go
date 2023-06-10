package main

import (
	"fmt"
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
	addr  = "localhost:9090"
	index = "index.html"
)

func Index(request *http.Request) http.Response {
	resp, err := http.RespondTo(request).WithFile(index)
	if err != nil {
		return http.RespondTo(request).
			WithCode(status.NotFound).
			WithBody(
				index + ": not found; try running this example directly from examples/combined folder",
			)
	}

	return resp
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

func Stressful(request *http.Request) http.Response {
	resp := http.RespondTo(request).
		WithHeader("Should", "never be seen").
		WithBody("Hello, world!")

	panic("TOO MUCH STRESS")

	return resp
}

func main() {
	r := inbuilt.NewRouter()

	r.Get("/stress", Stressful, middleware.Recover)

	root := r.Resource("/")
	root.Get(Index)
	root.Post(IndexSay)

	hello := r.Group("/hello")
	hello.Get("/world", World)
	hello.Get("/easter", Easter)

	s := settings.Default()
	s.TCP.ReadTimeout = time.Hour

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)

	if err := app.Serve(r, s); err != nil {
		log.Fatal(err)
	}
}
