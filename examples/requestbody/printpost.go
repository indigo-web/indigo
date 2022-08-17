package main

import (
	"fmt"
	"indigo"
	"indigo/http/status"
	"indigo/router"
	"indigo/types"
	"log"
	"strconv"
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
	r := router.NewDefaultRouter()
	r.Post("/say", MyHandler)

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(r))
}
