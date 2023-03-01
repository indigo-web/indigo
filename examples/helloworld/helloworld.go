package main

import (
	"fmt"
	"log"

	"github.com/indigo-web/indigo/v2/http"

	"github.com/indigo-web/indigo/v2"
	"github.com/indigo-web/indigo/v2/http/status"
	"github.com/indigo-web/indigo/v2/router/inbuilt"
)

var addr = "localhost:9090"

func MyHandler(request *http.Request) http.Response {
	return http.RespondTo(request).
		WithCode(status.OK).
		WithHeader("Hello", "world").
		WithBody("<h1>How are you doing?</h1>")
}

func main() {
	myRouter := inbuilt.NewRouter()
	myRouter.Get("/", MyHandler)

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(myRouter))
}
