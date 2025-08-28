package main

import (
	"log"
	"strconv"
	"time"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/codec"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/router/inbuilt/middleware"
)

func IndexSay(request *http.Request) *http.Response {
	if request.Headers.Value("talking") != "allowed" {
		return http.Code(request, status.UnavailableForLegalReasons)
	}

	body, err := request.Body.String()
	if err != nil {
		return http.Error(request, err)
	}

	log.Println("someone whispered:", strconv.Quote(body))

	return request.Respond()
}

func World(request *http.Request) *http.Response {
	return http.String(request, `<h1>Hello, world!</h1>`)
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
	s := config.Default()
	s.NET.ReadTimeout = time.Hour

	app := indigo.New(":8080").
		TLS(":8443", indigo.LocalCert()).
		Tune(s).
		Codec(codec.Suit()...).
		OnBind(func(addr string) {
			log.Printf("running on %s\n", addr)
		})

	r := inbuilt.New().
		Use(middleware.LogRequests()).
		Alias("/", "/static/index.html", method.GET).
		Alias("/favicon.ico", "/static/favicon.ico", method.GET).
		Static("/static", "examples/demo/static")

	r.Get("/stress", Stressful, middleware.Recover)

	r.Resource("/").
		Post(IndexSay)

	r.Post("/shutdown", func(request *http.Request) (_ *http.Response) {
		// TODO: this will stop the server BEFORE the answer can be sent.
		// TODO: There gotta be a better way to stop the server AFTER the handler exits.
		app.Stop()

		return http.Code(request, status.Teapot)
	})

	r.Group("/hello").
		Get("/world", World).
		Get("/easter", Easter)

	if err := app.Serve(r); err != nil {
		log.Fatal(err)
	}
}
