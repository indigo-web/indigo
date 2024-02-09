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

const addr = ":8080"

func IndexSay(request *http.Request) *http.Response {
	if request.Headers.Value("talking") != "allowed" {
		return http.Code(request, status.UnavailableForLegalReasons)
	}

	body, err := request.Body.String()
	if err != nil {
		return http.Error(request, err)
	}

	log.Println("someone says:", strconv.Quote(body))

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
	s := settings.Default()
	s.TCP.ReadTimeout = time.Hour

	app := indigo.New(addr).
		Tune(s).
		AutoHTTPS(8443).
		NotifyOnStart(func() {
			log.Println("initialized")
		})

	r := inbuilt.New().
		Use(middleware.LogRequests()).
		Alias("/", "/static/index.html").
		Alias("/favicon.ico", "/static/favicon.ico").
		Static("/static", "examples/combined/static")

	r.Get("/stress", Stressful, middleware.Recover)

	r.Resource("/").
		Post(IndexSay)

	r.Post("/shutdown", func(request *http.Request) (_ *http.Response) {
		app.GracefulStop()

		return http.Code(request, status.Teapot)
	})

	r.Group("/hello").
		Get("/world", World).
		Get("/easter", Easter)

	if err := app.Serve(r); err != nil {
		log.Fatal(err)
	}
}
