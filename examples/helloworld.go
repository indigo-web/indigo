package examples

import (
	"indigo"
	"indigo/http"
	"indigo/router"
	"indigo/types"
	"log"
)

func MyHandler(request *types.Request) *types.ResponseStruct {
	return types.Response().
		WithCode(http.StatusOk).
		WithBody([]byte("<h1>Hello, world!</h1>"))
}

func main() {
	myRouter := router.NewDefaultRouter()
	myRouter.Route("/", MyHandler)

	app := indigo.NewApp("localhost", 9090)
	log.Fatal(app.Serve(myRouter))
}
