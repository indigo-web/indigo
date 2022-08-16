package main

import (
	"fmt"
	"indigo"
	methods "indigo/http/method"
	"indigo/http/status"
	"indigo/router"
	"indigo/types"
	"log"
)

var addr = "localhost:9090"

func MyHandler(_ *types.Request) types.Response {
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
