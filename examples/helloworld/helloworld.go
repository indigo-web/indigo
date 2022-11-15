package main

import (
	"fmt"
	"github.com/fakefloordiv/indigo/http"
	"log"

	"github.com/fakefloordiv/indigo"
	"github.com/fakefloordiv/indigo/http/status"
	"github.com/fakefloordiv/indigo/router/inbuilt"
)

var addr = "localhost:9090"

func MyHandler(request *http.Request) http.Response {
	return request.Respond.
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
