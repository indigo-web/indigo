package main

import (
	"log"
	"strconv"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/router/inbuilt"
)

const (
	host = "0.0.0.0"
	port = 8080
)

func MyHandler(request *http.Request) *http.Response {
	conn, err := request.Hijack()
	if err != nil {
		// in case error occurred, it may be only an error with a network, so
		// no response may be sent anyway
		return request.Respond()
	}

	readBuff := make([]byte, 1024)

	for {
		n, err := conn.Read(readBuff)
		if n == 0 || err != nil {
			_ = conn.Close()

			// no matter what we return here as after handler exits, even if connection was
			// not explicitly closed, server will do it implicitly
			return request.Respond()
		}

		log.Println("somebody says:", strconv.Quote(string(readBuff[:n])))
	}
}

func main() {
	r := inbuilt.New()
	r.Get("/", MyHandler)

	app := indigo.NewApp(host, port)
	log.Println("Listening on", host, port)
	log.Fatal(app.Serve(r))
}
