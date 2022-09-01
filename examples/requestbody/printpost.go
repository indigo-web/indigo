package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/types"
)

var addr = "localhost:9090"

func MyHandler(request *types.Request) types.Response {
	body, err := request.Body()
	if err != nil {
		return types.WithResponse.WithCode(status.BadRequest)
	}

	fmt.Println("somebody said:", strconv.Quote(string(body)))

	return types.WithResponse.WithBody("Received and processed! Thank you!")
}

func main() {
	r := inbuilt.NewRouter()
	r.Post("/say", MyHandler)

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(r))
}
