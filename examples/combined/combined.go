package main

import (
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

const (
	host = "0.0.0.0"
	port = 8080
)

func IndexSay(request *http.Request) *http.Response {
	if talking := request.Headers.Value("talking"); talking != "allowed" {
		return http.Code(request, status.UnavailableForLegalReasons)
	}

	body, err := request.Body.String()
	if err != nil {
		return http.Error(request, err)
	}

	log.Println("Somebody said:", strconv.Quote(body))

	return request.Respond()
}

func World(request *http.Request) *http.Response {
	return request.Respond().String(
		`<h1>Hello, world!</h1>`,
	)
}

func Easter(request *http.Request) *http.Response {
	if request.Headers.Has("easter") {
		return request.Respond().
			Code(status.Teapot).
			Header("Easter", "Egg").
			String("You have discovered an easter egg! Congratulations!")
	}

	return request.Respond().String("Pretty ordinary page, isn't it?")
}

func Stressful(request *http.Request) *http.Response {
	resp := request.Respond().
		Header("Should", "never be seen").
		String("Hello, world!")

	panic("TOO MUCH STRESS")

	return resp
}

func main() {
	r := inbuilt.New().
		Get("/stress", Stressful, middleware.Recover)

	r.Resource("/").
		Post(IndexSay).
		Static("/static", "./examples/combined/static")

	r.Group("/hello").
		Get("/world", World).
		Get("/easter", Easter)

	s := settings.Default()
	s.TCP.ReadTimeout = time.Hour

	app := indigo.NewApp(host, port)
	log.Println("Listening on", host, port)

	if err := app.Serve(r, s); err != nil {
		log.Fatal(err)
	}
}
