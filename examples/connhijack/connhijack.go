package main

import (
	"fmt"
	"indigo"
	"indigo/http/status"
	"indigo/router/inbuilt"
	"indigo/types"
	"log"
	"strconv"
)

var addr = "localhost:9090"

func MyHandler(request *types.Request) types.Response {
	conn, err := request.Hijack()
	if err != nil {
		return types.WithResponse.
			WithCode(status.BadRequest).
			WithBody("bad body")
	}

	readBuff := make([]byte, 1024)

	for {
		n, err := conn.Read(readBuff)
		if n == 0 || err != nil {
			conn.Close()

			return types.WithResponse
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
