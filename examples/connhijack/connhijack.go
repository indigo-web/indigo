package main

import (
	"fmt"
	"github.com/fakefloordiv/indigo/http"
	"log"
	"strconv"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/router/inbuilt"
)

var addr = "localhost:9090"

func MyHandler(request *http.Request) http.Response {
	conn, err := request.Hijack()
	if err != nil {
		return request.Respond.WithCode(status.BadRequest).
			WithBody("bad body")
	}

	readBuff := make([]byte, 1024)

	for {
		n, err := conn.Read(readBuff)
		if n == 0 || err != nil {
			_ = conn.Close()

			return request.Respond
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
