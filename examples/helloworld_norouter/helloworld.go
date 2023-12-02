package main

import (
	"log"

	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/router/simple"
)

const (
	host = "0.0.0.0"
	port = 8080
)

func MyHandler(request *http.Request) *http.Response {
	return request.Respond().
		Code(status.OK).
		Header("Hello", "world").
		String("<h1>How are you doing?</h1>")
}

func main() {
	myRouter := simple.NewRouter(MyHandler, http.Error)

	app := indigo.NewApp(host, port)
	log.Println("Listening on", host, port)
	log.Fatal(app.Serve(myRouter))
}
