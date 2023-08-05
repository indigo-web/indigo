package main

import (
	"fmt"
	"log"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/inbuilt"
)

var addr = "localhost:9090"

func MyHandler(request *http.Request) http.Response {
	return request.Respond().
		WithCode(status.OK).
		WithHeader("Hello", "world").
		WithBody("<h1>How are you doing?</h1>")
}

func main() {
	myRouter := inbuilt.New()
	myRouter.Get("/", MyHandler)

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(myRouter))
}
