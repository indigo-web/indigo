package main

import (
	"log"
	"strconv"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/router/inbuilt"
)

const addr = ":8080"

func MyHandler(request *http.Request) *http.Response {
	client, err := request.Hijack()
	if err != nil {
		// in case error occurred, it may be only an error with a network, so
		// no response may be sent anyway
		return request.Respond()
	}

	// connection will be closed automatically as we'll be finished here

	for {
		data, err := client.Read()
		if err != nil {
			// after hijacking it makes no difference, what will be returned
			return nil
		}

		log.Println("somebody says:", strconv.Quote(string(data)))
	}
}

func main() {
	r := inbuilt.New()
	r.Get("/", MyHandler)

	app := indigo.New(addr).
		OnBind(func(addr string) {
			log.Printf("running on %s\n", addr)
		})

	log.Fatal(app.Serve(r))
}
