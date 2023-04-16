package main

import (
	"fmt"
	"github.com/indigo-web/indigo"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/simple"
	"log"
)

func main() {
	myRouter := simple.NewRouter(MyHandler, func(request *http.Request, err error) http.Response {
		return http.RespondTo(request).WithError(err)
	})

	fmt.Println("Listening on", addr)
	app := indigo.NewApp(addr)
	log.Fatal(app.Serve(myRouter))
}
