package examples

import (
	"fmt"
	"indigo"
	"indigo/http"
	"indigo/router"
	"indigo/types"
	"log"
)

/*
MyManualHandler is actually not much faster, even if response is static and cached
*/
func MyManualHandler(request *types.Request) types.Response {
	return types.Response{
		Body: []byte("<h1>Hello, world! From kit</h1>"),
	}
}

func MyBeautifulHandler(request *types.Request) types.Response {
	return types.NewResponse().
		WithCode(http.StatusOk).
		WithHeader("Hello", "world").
		WithBody("<h1>How are you doing?</h1>")
}

func main() {
	myRouter := router.NewDefaultRouter()
	myRouter.Route("/", MyManualHandler)
	myRouter.Route("/hello", MyBeautifulHandler)

	fmt.Println("Listening on localhost:9090")
	app := indigo.NewApp("localhost", 9090)
	log.Fatal(app.Serve(myRouter))
}
