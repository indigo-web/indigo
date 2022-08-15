package main

import (
	"fmt"
	methods "indigo/http/method"
	"log"

	"indigo"
	"indigo/http/status"
	"indigo/router"
	"indigo/types"
)

var addr = "localhost:9090"

func MyHandler(request *types.Request) types.Response {
	return types.WithResponse.
		WithCode(status.OK).
		WithHeader("Hello", "world").
		WithBody("<h1>How are you doing?</h1>")
}

func main() {
	myRouter := router.NewDefaultRouter()
	myRouter.Route(methods.GET, "/", MyHandler)

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(myRouter))
}
