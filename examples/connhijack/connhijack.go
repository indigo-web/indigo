package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/fakefloordiv/indigo/http"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/router/inbuilt"
)

var addr = "localhost:9090"

func MyHandler(request *http.Request) http.Response {
	conn, err := request.Hijack()
	if err != nil {
		// in case error occurred, it may be only an error with a network, so
		// no response may be sent anyway
		return http.RespondTo(request)
	}

	readBuff := make([]byte, 1024)

	for {
		n, err := conn.Read(readBuff)
		if n == 0 || err != nil {
			_ = conn.Close()

			// no matter what we return here as after handler exits, even if connection was
			// not explicitly closed, server will do it implicitly
			return http.RespondTo(request)
		}

		fmt.Println("somebody says:", strconv.Quote(string(readBuff[:n])))
	}
}

func main() {
	r := inbuilt.NewRouter()
	r.Get("/", MyHandler)

	app := indigo.NewApp(addr)
	fmt.Println("Listening on", addr)
	log.Fatal(app.Serve(r))
}
